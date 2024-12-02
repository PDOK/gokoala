package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
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

func (p *Postgres) Suggest(ctx context.Context, suggestForThis string) ([]string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, p.queryTimeout)
	defer cancel()

	// Prepare dynamic full-text search query
	// Split terms by spaces and append :* to each term
	terms := strings.Fields(suggestForThis)
	for i, term := range terms {
		terms[i] = term + ":*"
	}
	searchTerm := strings.Join(terms, " & ")

	sqlQuery := fmt.Sprintf(
		`SELECT
	r.display_name AS display_name,
	max(r.rank) AS rank,
	max(r.highlighted_text) AS highlighted_text
	FROM (
		SELECT display_name, 
	    ts_rank_cd(ts, to_tsquery('%[1]s'), 1) AS rank,
	    ts_headline('dutch', suggest, to_tsquery('%[2]s')) AS highlighted_text
		FROM
		%[3]s
		WHERE ts @@ to_tsquery('%[4]s') LIMIT 500
	) r
	GROUP BY display_name
	ORDER BY rank DESC, display_name ASC LIMIT 50`,
		searchTerm, searchTerm, p.searchIndex, searchTerm)

	// Execute query
	rows, err := p.db.Query(queryCtx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", sqlQuery, err)
	}
	defer rows.Close()

	if queryCtx.Err() != nil {
		return nil, queryCtx.Err()
	}

	var suggestions []string
	for rows.Next() {
		var displayName, highlightedText string
		var rank float64

		// Scan all selected columns
		if err := rows.Scan(&displayName, &rank, &highlightedText); err != nil {
			return nil, err
		}

		suggestions = append(suggestions, highlightedText) // or displayName, whichever you want to return
	}

	return suggestions, nil
}
