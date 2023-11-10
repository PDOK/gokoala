package datasources

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type contextKey int

const sqlContextKey contextKey = iota

// SQLLog query logging for debugging purposes
type SQLLog struct{}

// Before callback prior to execution of the given SQL query
func (s *SQLLog) Before(ctx context.Context, _ string, _ ...interface{}) (context.Context, error) {
	return context.WithValue(ctx, sqlContextKey, time.Now()), nil
}

// After callback once execution of the given SQL query is done
func (s *SQLLog) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	if os.Getenv("LOG_SQL") == "true" {
		query = replaceBindVars(query, args)
		start := ctx.Value(sqlContextKey).(time.Time)

		log.Printf("\n--- SQL:\n%s\n--- SQL query took: %s\n", query, time.Since(start))
	}
	return ctx, nil
}

// replaces '?' bind vars in order to log a complete query
func replaceBindVars(query string, args []interface{}) string {
	for _, arg := range args {
		query = strings.Replace(query, "?", fmt.Sprintf("%v", arg), 1)
	}
	return query
}
