package postgres

import (
	"context"
	"fmt"

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
}

func NewPostgres(dbConn string, queryTimeout time.Duration, searchIndex string) (*Postgres, error) {
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

	return &Postgres{db, ctx, queryTimeout, searchIndex}, nil
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

	// Execute search query
	rows, err := p.db.Query(queryCtx, sql, limit, wildcardQuery, exactMatchQuery, names, versions, relevance)
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
				(ts_rank(r.ts, (SELECT query FROM query_exact), 1) + 0.01 + ts_rank(r.ts, (SELECT query FROM query_wildcard), 1)) * rel.relevance
			ELSE
				(ts_rank(r.ts, (SELECT query FROM query_exact), 1) + ts_rank(r.ts, (SELECT query FROM query_wildcard), 1)) * rel.relevance
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
	        LIMIT 40000
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
				    -- less then 40000 results don't need to be pre-ranked, they can be ranked based on score
					CASE WHEN (SELECT c from results_count) < 40000 THEN 1 = 1 END
			) UNION ALL (
		    	SELECT
					*
				FROM
					results r
				WHERE
				    -- pre-rank more then 40000 results by ordering on suggest length and display_name
					CASE WHEN (SELECT c from results_count) = 40000 THEN 1 = 1 END
				ORDER BY
					char_length(r.suggest) ASC,
					r.display_name COLLATE "custom_numeric" ASC
				LIMIT 400 -- return 400 pre-ranked results for ranking based on score
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
		geojsonBbox, err := geojson.Encode(bbox, geojson.EncodeGeometryWithMaxDecimalDigits(10))
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
