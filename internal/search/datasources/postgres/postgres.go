package postgres

import (
	"context"
	"fmt"
	"log"

	d "github.com/PDOK/gomagpie/internal/search/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
	pgxgeom "github.com/twpayne/pgx-geom"

	"time"
)

type Postgres struct {
	db  *pgxpool.Pool
	ctx context.Context

	queryTimeout time.Duration
	searchIndex  string

	rankNormalization        int
	exactMatchMultiplier     float64
	primarySuggestMultiplier float64
	rankThreshold            int
	preRankLimit             int
}

func NewPostgres(dbConn string, queryTimeout time.Duration, searchIndex string, rankNormalization int, exactMatchMultiplier float64, primarySuggestMultiplier float64, rankThreshold int, preRankLimit int) (*Postgres, error) {
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(dbConn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// add support for Go <-> PostGIS conversions
	config.AfterConnect = pgxgeom.Register

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &Postgres{
		db,
		ctx,
		queryTimeout,
		searchIndex,
		rankNormalization,
		exactMatchMultiplier,
		primarySuggestMultiplier,
		rankThreshold,
		preRankLimit,
	}, nil
}

func (p *Postgres) Close() {
	p.db.Close()
}

func (p *Postgres) SearchFeaturesAcrossCollections(ctx context.Context, searchQuery d.SearchQuery,
	collections d.CollectionsWithParams, srid d.SRID, limit int) (*d.FeatureCollection, error) {

	queryCtx, cancel := context.WithTimeout(ctx, p.queryTimeout)
	defer cancel()

	sql := makeSQL(p.searchIndex, srid)
	wildcardQuery := searchQuery.ToWildcardQuery()
	exactMatchQuery := searchQuery.ToExactMatchQuery()
	names, versions, relevance := collections.NamesAndVersionsAndRelevance()
	log.Printf("\nSEARCH QUERY (wildcard): %s\n", wildcardQuery)

	// Execute search query
	rows, err := p.db.Query(
		queryCtx,
		sql,
		limit,
		wildcardQuery,
		exactMatchQuery,
		names,
		versions,
		relevance,
		p.rankNormalization,
		p.exactMatchMultiplier,
		p.primarySuggestMultiplier,
		p.rankThreshold,
		p.preRankLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", sql, err)
	}
	defer rows.Close()

	// Turn rows into FeatureCollection
	return mapRowsToFeatures(queryCtx, rows)
}

//nolint:funlen
func makeSQL(index string, srid d.SRID) string {
	// language=postgresql
	return fmt.Sprintf(`
	WITH query_wildcard AS (
		SELECT to_tsquery('custom_dict', $2) query
	),
	query_exact AS (
		SELECT to_tsquery('custom_dict', $3) query
	),
	results AS NOT MATERIALIZED ( -- the results query is called twice, materializing it results in a non optimal query plan for one of the calls
		SELECT
			r.display_name,
			r.feature_id,
			r.external_fid,
			r.collection_id,
			r.collection_version,
			r.geometry_type,
			r.bbox,
			r.geometry,
			r.suggest,
			CASE WHEN r.display_name = r.suggest THEN
				(
				    ts_rank_cd(r.ts, (SELECT query FROM query_exact), $7) * $8 * $9 + ts_rank_cd(r.ts, (SELECT query FROM query_wildcard), $7)
				) * rel.relevance
			ELSE
				(
				    ts_rank_cd(r.ts, (SELECT query FROM query_exact), $7) * $8 + ts_rank_cd(r.ts, (SELECT query FROM query_wildcard), $7)
				) * rel.relevance
			END AS rank,
			ts_headline('custom_dict', r.suggest, (SELECT query FROM query_wildcard)) AS highlighted_text
		FROM
			%[1]s r
		LEFT JOIN
			(SELECT * FROM unnest($4::text[], $6::float[]) rel(collection_id,relevance)) rel
		ON
			rel.collection_id = r.collection_id
		WHERE
			r.ts @@ (SELECT query FROM query_wildcard) AND (r.collection_id, r.collection_version) IN (
				-- make a virtual table by creating tuples from the provided arrays.
				SELECT * FROM unnest($4::text[], $5::int[])
			)
	),
	results_count AS (
	    SELECT
	    	COUNT(*) c
	    FROM (
	        SELECT
	        	r.id
	        FROM
	        	%[1]s r
			WHERE
				r.ts @@ (SELECT query FROM query_wildcard) AND (r.collection_id, r.collection_version) IN (
					-- make a virtual table by creating tuples from the provided arrays.
					SELECT * FROM unnest($4::text[], $5::int[])
				)
	        LIMIT $10
	    )
	)
	SELECT
	    rn.display_name,
		rn.feature_id,
		rn.external_fid,
		rn.collection_id,
		rn.collection_version,
		rn.geometry_type,
		st_transform(rn.bbox, %[2]d)::geometry AS bbox,
		st_transform(rn.geometry, %[2]d)::geometry AS geometry,
		rn.rank,
		rn.highlighted_text
	FROM (
		SELECT
			*,
			ROW_NUMBER() OVER (
				PARTITION BY
					r.display_name,
					r.collection_id,
					r.collection_version,
					r.feature_id,
					r.external_fid
				ORDER BY -- use same "order by" clause everywhere
					r.rank DESC,
					r.display_name COLLATE "custom_numeric" ASC
			) AS row_number
		FROM (
		    -- a UNION ALL is used, because a CASE in the ORDER BY clause causes a sequence scan instead of an index scan
			-- because of 1 = 1 in the WHERE clauses below the results are only added if WHEN is true, otherwise the results are ignored
			(
				SELECT
					*
				FROM
					results r
				WHERE
				    -- less then rank threshold results don't need to be pre-ranked, they can be ranked based on score
					CASE WHEN (SELECT c from results_count) < $10 THEN 1 = 1 END
			) UNION ALL (
		    	SELECT
					*
				FROM
					results r
				WHERE
				    -- pre-rank more then rank threshold results by ordering on suggest length and display_name
					CASE WHEN (SELECT c from results_count) = $10 THEN 1 = 1 END
				ORDER BY
					char_length(r.suggest) ASC,
					r.display_name COLLATE "custom_numeric" ASC
				LIMIT $11 -- return limited pre-ranked results for ranking based on score
			)
		) r
	) rn
	WHERE rn.row_number = 1
	ORDER BY -- use same "order by" clause everywhere
	    rn.rank DESC,
	    rn.display_name COLLATE "custom_numeric" ASC
	LIMIT $1`, index, srid) // don't add user input here, use $X params for user input!
}

func mapRowsToFeatures(queryCtx context.Context, rows pgx.Rows) (*d.FeatureCollection, error) {
	fc := d.FeatureCollection{Features: make([]*d.Feature, 0)}
	for rows.Next() {
		var displayName, highlightedText, featureID, collectionID, collectionVersion, geomType string
		var rank float64
		var bbox, geometry geom.T
		var externalFid *string

		if err := rows.Scan(&displayName, &featureID, &externalFid, &collectionID, &collectionVersion, &geomType,
			&bbox, &geometry, &rank, &highlightedText); err != nil {
			return nil, err
		}
		geojsonBbox, err := encodeBBox(bbox)
		if err != nil {
			return nil, err
		}
		geojsonGeom, err := geojson.Encode(geometry, geojson.EncodeGeometryWithMaxDecimalDigits(10))
		if err != nil {
			return nil, err
		}
		fc.Features = append(fc.Features, &d.Feature{
			ID:       getFeatureID(externalFid, featureID),
			Geometry: geojsonGeom,
			Bbox:     geojsonBbox,
			Properties: map[string]any{
				d.PropCollectionID:      collectionID,
				d.PropCollectionVersion: collectionVersion,
				d.PropGeomType:          geomType,
				d.PropDisplayName:       displayName,
				d.PropHighlight:         highlightedText,
				d.PropScore:             rank,
			},
		})
		fc.NumberReturned = len(fc.Features)
	}
	return &fc, queryCtx.Err()
}

func getFeatureID(externalFid *string, featureID string) string {
	if externalFid != nil {
		return *externalFid
	}
	return featureID
}

// adapted from https://github.com/twpayne/go-geom/blob/b22fd061f1531a51582333b5bd45710a455c4978/encoding/geojson/geojson.go#L525
// encodeBBox encodes b as a GeoJson Bounding Box.
func encodeBBox(bbox geom.T) (*[]float64, error) {
	if bbox == nil {
		return nil, nil
	}
	b := bbox.Bounds()
	switch l := b.Layout(); l {
	case geom.XY, geom.XYM:
		return &[]float64{b.Min(0), b.Min(1), b.Max(0), b.Max(1)}, nil
	case geom.XYZ, geom.XYZM, geom.NoLayout:
		return nil, fmt.Errorf("unsupported type: %d", rune(l))
	default:
		return nil, fmt.Errorf("unsupported type: %d", rune(l))
	}
}
