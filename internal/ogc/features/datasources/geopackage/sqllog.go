package geopackage

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PDOK/gokoala/internal/ogc/features/datasources/common"
)

type contextKey int

const (
	sqlContextKey contextKey = iota
)

// SQLLog query logging for debugging purposes
type SQLLog struct {
	LogSQL bool
}

// NewSQLLogFromEnv build a SQLLog based on the `LOG_SQL` environment variable
func NewSQLLogFromEnv() *SQLLog {
	var err error
	logSQL := false
	if os.Getenv(common.EnvLogSQL) != "" {
		logSQL, err = strconv.ParseBool(os.Getenv(common.EnvLogSQL))
		if err != nil {
			log.Fatalf("invalid %s value provided, must be a boolean", common.EnvLogSQL)
		}
	}
	return &SQLLog{LogSQL: logSQL}
}

// Before callback prior to execution of the given SQL query
func (s *SQLLog) Before(ctx context.Context, _ string, _ ...any) (context.Context, error) {
	return context.WithValue(ctx, sqlContextKey, time.Now()), nil
}

// After callback once execution of the given SQL query is done
func (s *SQLLog) After(ctx context.Context, query string, args ...any) (context.Context, error) {
	start := ctx.Value(sqlContextKey).(time.Time)
	timeSpent := time.Since(start)
	if s.LogSQL {
		query = replaceBindVars(query, args)
		log.Printf("\n--- SQL:\n%s\n--- SQL query took: %s\n", query, timeSpent)
	}
	return ctx, nil
}

// replaceBindVars replaces '?' bind vars with actual values to log a complete query (best effort)
func replaceBindVars(query string, args []any) string {
	for _, arg := range args {
		query = strings.Replace(query, "?", fmt.Sprintf("%v", arg), 1)
	}
	return query
}
