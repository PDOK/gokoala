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
	if os.Getenv("LOG_SQL") != "" {
		query = ReplaceBindVars(query, args)
		start := ctx.Value(sqlContextKey).(time.Time)

		log.Printf("SQL:\n%s\nSQL query took: %s\n", query, time.Since(start))
	}
	return ctx, nil
}

// ReplaceBindVars replaces $1, $2, $3, etc bind vars in order to log a complete query
func ReplaceBindVars(query string, args []interface{}) string {
	for i, arg := range args {
		query = strings.ReplaceAll(query, fmt.Sprintf("$%d", i+1), fmt.Sprintf("%v", arg))
	}
	return query
}
