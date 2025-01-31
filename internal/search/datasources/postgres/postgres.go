package postgres

import (
	"context"
	"fmt"

	d "github.com/PDOK/gomagpie/internal/search/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pggeom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
	pgxgeom "github.com/twpayne/pgx-geom"

	"strings"
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

func (p *Postgres) SearchFeaturesAcrossCollections(ctx context.Context, searchTerm string, collections d.CollectionsWithParams,
	srid d.SRID, limit int) (*d.FeatureCollection, error) {

	queryCtx, cancel := context.WithTimeout(ctx, p.queryTimeout)
	defer cancel()

	// Split terms by spaces and append :* to each term
	termsWildcard := strings.Fields(searchTerm)
	for i, term := range termsWildcard {
		termsWildcard[i] = term + ":*"
	}
	termsWildcardConcat := strings.Join(termsWildcard, " & ")
	termExactConcat := strings.Join(strings.Fields(searchTerm), " | ")
	query := makeSearchQuery(p.searchIndex, srid)

	// Execute search query
	names, versions, relevance := collections.NamesAndVersionsAndRelevance()
	rows, err := p.db.Query(queryCtx, query, limit, termsWildcardConcat, termExactConcat, names, versions, relevance)
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	// Turn rows into FeatureCollection
	return mapRowsToFeatures(queryCtx, rows)
}

func makeSearchQuery(index string, srid d.SRID) string {
	// language=postgresql
	query := fmt.Sprintf(`
	WITH query_wildcard AS (
		SELECT to_tsquery('simple', $2) query
	),
	query_exact AS (
		SELECT to_tsquery('simple', $3) query
	)
	SELECT
	    rn.display_name,
		rn.feature_id,
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
					r.feature_id
				ORDER BY
					r.rank DESC,
					r.display_name ASC
			) AS row_number
		FROM (
			SELECT
				r.display_name,
				r.feature_id,
				r.collection_id,
				r.collection_version,
				r.geometry_type,
				r.bbox,
				CASE WHEN r.display_name=r.suggest THEN
					(ts_rank(r.ts, (SELECT query FROM query_exact), 1) + 0.01 + ts_rank(r.ts, (SELECT query FROM query_wildcard), 1)) * rel.relevance
				ELSE
				    (ts_rank(r.ts, (SELECT query FROM query_exact), 1) + ts_rank(r.ts, (SELECT query FROM query_wildcard), 1)) * rel.relevance
				END AS rank,
				ts_headline('simple', r.suggest, (SELECT query FROM query_wildcard)) AS highlighted_text
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
			ORDER BY -- keep the same as outer and row_number 'order by' clause
			    rank DESC,
			    r.display_name ASC
			LIMIT 500
		) r
	) rn
	WHERE rn.row_number = 1
	ORDER BY
	    rn.rank DESC,
	    rn.display_name ASC
	LIMIT $1`, index, srid) // don't add user input here, use $X params for user input!

	return query
}

func mapRowsToFeatures(queryCtx context.Context, rows pgx.Rows) (*d.FeatureCollection, error) {
	fc := d.FeatureCollection{Features: make([]*d.Feature, 0)}
	for rows.Next() {
		var displayName, highlightedText, featureID, collectionID, collectionVersion, geomType string
		var rank float64
		var bbox pggeom.T

		if err := rows.Scan(&displayName, &featureID, &collectionID, &collectionVersion, &geomType,
			&bbox, &rank, &highlightedText); err != nil {
			return nil, err
		}
		geojsonGeom, err := geojson.Encode(bbox)
		if err != nil {
			return nil, err
		}
		fc.Features = append(fc.Features, &d.Feature{
			ID:       featureID,
			Geometry: *geojsonGeom,
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
