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
	terms := strings.Fields(searchTerm)
	for i, term := range terms {
		terms[i] = term + ":*"
	}
	termsConcat := strings.Join(terms, " & ")
	query := makeSearchQuery(p.searchIndex, srid)

	// Execute search query
	names, ints := collections.NamesAndVersions()
	rows, err := p.db.Query(queryCtx, query, limit, termsConcat, names, ints)
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
	with query as (
		select to_tsquery('dutch', $2) query
	)
	select 	r.display_name as display_name, 
			max(r.feature_id) as feature_id,
			max(r.collection_id) as collection_id,
			max(r.collection_version) as collection_version,
			max(r.geometry_type) as geometry_type,
			st_transform(max(r.bbox), %[2]d)::geometry as bbox,
			max(r.rank) as rank, 
			max(r.highlighted_text) as highlighted_text
	from (
		select display_name, feature_id, collection_id, collection_version, geometry_type, bbox,
				ts_rank_cd(ts, (select query from query), 1) as rank,
				ts_headline('dutch', display_name, (select query from query)) as highlighted_text
		from %[1]s
		where ts @@ (select query from query) and (collection_id, collection_version) in (
		    -- make a virtual table by creating tuples from the provided arrays.
			select * from unnest($3::text[], $4::int[])
		)
		order by rank desc, display_name asc -- keep the same as outer 'order by' clause
		limit 500
	) r
	group by r.display_name, r.collection_id, r.collection_version, r.feature_id
	order by rank desc, display_name asc
	limit $1`, index, srid) // don't add user input here, use $X params for user input!

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
