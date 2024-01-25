package datasources

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type contextKey int

const (
	envLogSQL                       = "LOG_SQL"
	envSlowQueryTime                = "SLOW_QUERY_TIME"
	defaultSlowQueryTime            = 5 * time.Second
	sqlContextKey        contextKey = iota
)

// SQLLog query logging for debugging purposes
type SQLLog struct {
	LogSQL        bool
	SlowQueryTime time.Duration
}

// NewSQLLogFromEnv build a SQLLog from environment variables listed in this file
func NewSQLLogFromEnv() *SQLLog {
	var err error
	logSQL := false
	if os.Getenv(envLogSQL) != "" {
		logSQL, err = strconv.ParseBool(os.Getenv(envLogSQL))
		if err != nil {
			log.Fatalf("invalid %s value provided, must be a boolean", envLogSQL)
		}
	}
	slowQueryTime := defaultSlowQueryTime
	if os.Getenv(envSlowQueryTime) != "" {
		slowQueryTime, err = time.ParseDuration(os.Getenv(envSlowQueryTime))
		if err != nil {
			log.Fatalf("invalid %s value provided, value such as '5s' expected", envSlowQueryTime)
		}
	}
	return &SQLLog{LogSQL: logSQL, SlowQueryTime: slowQueryTime}
}

// Before callback prior to execution of the given SQL query
func (s *SQLLog) Before(ctx context.Context, _ string, _ ...any) (context.Context, error) {
	return context.WithValue(ctx, sqlContextKey, time.Now()), nil
}

// After callback once execution of the given SQL query is done
func (s *SQLLog) After(ctx context.Context, query string, args ...any) (context.Context, error) {
	start := ctx.Value(sqlContextKey).(time.Time)
	timeSpent := time.Since(start)
	if timeSpent > s.SlowQueryTime || s.LogSQL {
		query = replaceBindVars(query, args)
		log.Printf("\n--- SQL:\n%s\n--- SQL query took: %s\n", query, timeSpent)
	}
	return ctx, nil
}

// replaceBindVars replaces '?' bind vars in order to log a complete query
func replaceBindVars(query string, args []any) string {
	for _, arg := range args {
		query = strings.Replace(query, "?", fmt.Sprintf("%v", arg), 1)
	}
	return query
}
