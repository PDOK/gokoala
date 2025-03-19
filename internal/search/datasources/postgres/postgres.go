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
	"github.com/twpayne/go-geom/encoding/wkt"
	pgxgeom "github.com/twpayne/pgx-geom"

	"time"
)

type Postgres struct {
	db  *pgxpool.Pool
	ctx context.Context

	queryTimeout    time.Duration
	searchIndex     string
	searchIndexSrid d.SRID

	rankNormalization        int
	exactMatchMultiplier     float64
	primarySuggestMultiplier float64
	rankThreshold            int
	preRankLimitMultiplier   int
	synonymsExactMatch       bool
}

func NewPostgres(dbConn string, queryTimeout time.Duration, searchIndex string, searchIndexSrid d.SRID,
	rankNormalization int, exactMatchMultiplier float64, primarySuggestMultiplier float64, rankThreshold int,
	preRankLimitMultiplier int, synonymsExactMatch bool) (*Postgres, error) {

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
		searchIndexSrid,
		rankNormalization,
		exactMatchMultiplier,
		primarySuggestMultiplier,
		rankThreshold,
		preRankLimitMultiplier,
		synonymsExactMatch,
	}, nil
}

func (p *Postgres) Close() {
	p.db.Close()
}

func (p *Postgres) SearchFeaturesAcrossCollections(ctx context.Context, searchQuery d.SearchQuery,
	collections d.CollectionsWithParams, srid d.SRID, bbox *geom.Bounds, bboxSRID d.SRID, limit int) (*d.FeatureCollection, error) {

	queryCtx, cancel := context.WithTimeout(ctx, p.queryTimeout)
	defer cancel()

	bboxFilter, bboxQueryArgs, err := parseBbox(bbox, bboxSRID, p.searchIndexSrid)
	if err != nil {
		return nil, err
	}
	sql := makeSQL(p.searchIndex, srid, bboxFilter)
	wildcardQuery := searchQuery.ToWildcardQuery()
	exactMatchQuery := searchQuery.ToExactMatchQuery(p.synonymsExactMatch)
	names, versions, relevance := collections.NamesAndVersionsAndRelevance()
	log.Printf("\nSEARCH QUERY (wildcard): %s\n", wildcardQuery)

	// Execute search query
	queryArgs := append([]any{limit,
		wildcardQuery,
		exactMatchQuery,
		names,
		versions,
		relevance,
		p.rankNormalization,
		p.exactMatchMultiplier,
		p.primarySuggestMultiplier,
		p.rankThreshold,
		p.preRankLimitMultiplier}, bboxQueryArgs...)
	rows, err := p.db.Query(queryCtx, sql, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", sql, err)
	}
	defer rows.Close()

	// Turn rows into FeatureCollection
	return mapRowsToFeatures(queryCtx, rows)
}

//nolint:funlen
func makeSQL(index string, srid d.SRID, bboxFilter string) string {
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
			r.ts
		FROM
			%[1]s r
		WHERE
			r.ts @@ (SELECT query FROM query_wildcard) AND (r.collection_id, r.collection_version) IN (
				-- make a virtual table by creating tuples from the provided arrays.
				SELECT * FROM unnest($4::text[], $5::int[])
			)
		%[3]s -- bounding box intersect filter
	),
	results_count AS (
	    SELECT
	    	COUNT(*) c
	    FROM (
	        SELECT
				r.feature_id
	        FROM
				results r
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
		ts_headline('custom_dict', rn.suggest, (SELECT query FROM query_wildcard)) AS highlighted_text
	FROM (
		SELECT
			r.*,
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
			SELECT
				u.*,
				CASE WHEN u.display_name = u.suggest THEN (
					ts_rank_cd(u.ts, (SELECT query FROM query_exact), $7) * $8 * $9 + ts_rank_cd(u.ts, (SELECT query FROM query_wildcard), $7)
				) * rel.relevance
				ELSE (
					ts_rank_cd(u.ts, (SELECT query FROM query_exact), $7) * $8 + ts_rank_cd(u.ts, (SELECT query FROM query_wildcard), $7)
				) * rel.relevance
				END AS rank
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
						array_length(string_to_array(r.suggest, ' '), 1) ASC,
						r.display_name COLLATE "custom_numeric" ASC
					LIMIT $1::int * $11::int -- return limited pre-ranked results for ranking based on score
				)
			) u
			LEFT JOIN
				(SELECT * FROM unnest($4::text[], $6::float[]) rel(collection_id,relevance)) rel
			ON
				rel.collection_id = u.collection_id
		) r
	) rn
	WHERE rn.row_number = 1
	ORDER BY -- use same "order by" clause everywhere
	    rn.rank DESC,
	    rn.display_name COLLATE "custom_numeric" ASC
	LIMIT $1`, index, srid, bboxFilter) // don't add user input here, use $X params for user input!
}

func parseBbox(bbox *geom.Bounds, bboxSRID d.SRID, searchIndexSRID d.SRID) (string, []any, error) {
	var bboxFilter, bboxWkt string
	var bboxQueryArgs []any
	var err error
	if bbox != nil {
		bboxFilter = fmt.Sprintf(`AND
			(st_intersects(r.geometry, st_transform(st_geomfromtext($12::text, $13::int), %[1]d)) OR st_intersects(r.bbox, st_transform(st_geomfromtext($12::text, $13::int), %[1]d)))
		`, searchIndexSRID)
		bboxWkt, err = wkt.Marshal(bbox.Polygon())
		if err != nil {
			return "", []any{}, err
		}
		bboxQueryArgs = append(bboxQueryArgs, bboxWkt, bboxSRID)
	}
	return bboxFilter, bboxQueryArgs, err
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
