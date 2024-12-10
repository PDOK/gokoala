package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/PDOK/gomagpie/internal/search/domain"
	"github.com/jackc/pgx/v5"
	pggeom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
	pgxgeom "github.com/twpayne/pgx-geom"

	"strings"
	"time"
)

type Postgres struct {
	db  *pgx.Conn
	ctx context.Context

	queryTimeout time.Duration
	searchIndex  string
}

func NewPostgres(dbConn string, queryTimeout time.Duration, searchIndex string) (*Postgres, error) {
	ctx := context.Background()
	db, err := pgx.Connect(ctx, dbConn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	// add support for Go <-> PostGIS conversions
	if err := pgxgeom.Register(ctx, db); err != nil {
		return nil, err
	}
	return &Postgres{db, ctx, queryTimeout, searchIndex}, nil
}

func (p *Postgres) Close() {
	_ = p.db.Close(p.ctx)
}

func (p *Postgres) Suggest(ctx context.Context, searchTerm string, collections map[string]map[string]string,
	srid domain.SRID, limit int) (*domain.FeatureCollection, error) {

	queryCtx, cancel := context.WithTimeout(ctx, p.queryTimeout)
	defer cancel()

	// Prepare dynamic full-text search query
	// Split terms by spaces and append :* to each term
	terms := strings.Fields(searchTerm)
	for i, term := range terms {
		terms[i] = term + ":*"
	}
	termsConcat := strings.Join(terms, " & ")
	searchQuery := makeSearchQuery(termsConcat, p.searchIndex)

	// Execute search query
	rows, err := p.db.Query(queryCtx, searchQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", searchQuery, err)
	}
	defer rows.Close()

	// Turn rows into FeatureCollection
	fc := domain.FeatureCollection{Features: make([]*domain.Feature, 0)}
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
		f := domain.Feature{
			ID:       featureID,
			Geometry: *geojsonGeom,
			Properties: map[string]any{
				"collectionId":           collectionID,
				"collectionVersion":      collectionVersion,
				"collectionGeometryType": geomType,
				"displayName":            displayName,
				"highlight":              highlightedText,
				"score":                  rank,
			},
		}
		log.Printf("collections %s, srid %v", collections, srid) // TODO  use params
		fc.Features = append(fc.Features, &f)
		fc.NumberReturned = len(fc.Features)
	}
	return &fc, queryCtx.Err()
}

func makeSearchQuery(term string, index string) string {
	// language=postgresql
	return fmt.Sprintf(`
	select r.display_name as display_name, 
	       max(r.feature_id) as feature_id,
		   max(r.collection_id) as collection_id,
		   max(r.collection_version) as collection_version,
		   max(r.geometry_type) as geometry_type,
		   cast(max(r.bbox) as geometry) as bbox,
		   max(r.rank) as rank, 
		   max(r.highlighted_text) as highlighted_text
	from (
		select display_name, feature_id, collection_id, collection_version, geometry_type, bbox,
	           ts_rank_cd(ts, to_tsquery('%[1]s'), 1) as rank,
	    	   ts_headline('dutch', suggest, to_tsquery('%[1]s')) as highlighted_text
		from %[2]s
		where ts @@ to_tsquery('%[1]s') 
		limit 500
	) r
	group by r.display_name
	order by rank desc, display_name asc
	limit $1`, term, index)
}
