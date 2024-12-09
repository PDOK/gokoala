package postgres

import (
	"context"
	"fmt"

	"github.com/PDOK/gomagpie/internal/search/domain"
	"github.com/jackc/pgx/v5"
	pggeom "github.com/twpayne/go-geom"
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

func (p *Postgres) Suggest(ctx context.Context, searchTerm string, _ map[string]map[string]string, _ domain.SRID, limit int) ([]string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, p.queryTimeout)
	defer cancel()

	// Prepare dynamic full-text search query
	// Split terms by spaces and append :* to each term
	terms := strings.Fields(searchTerm)
	for i, term := range terms {
		terms[i] = term + ":*"
	}
	searchTermForPostgres := strings.Join(terms, " & ")

	sqlQuery := fmt.Sprintf(`
	select r.display_name as display_name, 
	       max(r.feature_id) as feature_id,
		   max(r.collection_id) as collection_id,
		   max(r.collection_version) as collection_version,
		   cast(max(r.bbox) as geometry) as bbox,
		   max(r.rank) as rank, 
		   max(r.highlighted_text) as highlighted_text
	from (
		select display_name, feature_id, collection_id, collection_version, bbox,
	           ts_rank_cd(ts, to_tsquery('%[1]s'), 1) as rank,
	    	   ts_headline('dutch', suggest, to_tsquery('%[1]s')) as highlighted_text
		from %[2]s
		where ts @@ to_tsquery('%[1]s') 
		limit 500
	) r
	group by r.display_name
	order by rank desc, display_name asc
	limit $1`, searchTermForPostgres, p.searchIndex)

	// Execute query
	rows, err := p.db.Query(queryCtx, sqlQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", sqlQuery, err)
	}
	defer rows.Close()

	var suggestions []string
	for rows.Next() {
		var displayName, highlightedText, featureID, collectionID, collectionVersion string
		var rank float64
		var bbox pggeom.Polygon

		if err := rows.Scan(&displayName, &featureID, &collectionID, &collectionVersion, &bbox, &rank, &highlightedText); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, highlightedText) // or displayName, whichever you want to return
	}

	return suggestions, queryCtx.Err()
}
