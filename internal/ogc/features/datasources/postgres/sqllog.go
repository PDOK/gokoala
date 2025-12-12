package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/PDOK/gokoala/internal/ogc/features/datasources/common"
	"github.com/jackc/pgx/v5/tracelog"
)

type stdoutLogger struct {
	logger *log.Logger
}

// SQLLog query logging for debugging purposes.
type SQLLog struct {
	LogSQL bool
	Tracer *tracelog.TraceLog
}

// NewSQLLogFromEnv build a SQLLog based on the `LOG_SQL` environment variable.
func NewSQLLogFromEnv() *SQLLog {
	var err error
	logSQL := false
	var tracer *tracelog.TraceLog
	if os.Getenv(common.EnvLogSQL) != "" {
		logSQL, err = strconv.ParseBool(os.Getenv(common.EnvLogSQL))
		if err != nil {
			log.Fatalf("invalid %s value provided, must be a boolean", common.EnvLogSQL)
		}

		loggerAdapter := &stdoutLogger{logger: log.New(os.Stdout, "POSTGRES: ", log.LstdFlags)}
		tracer = &tracelog.TraceLog{
			Logger:   loggerAdapter,
			LogLevel: tracelog.LogLevelTrace, // Set to Trace to see all query details
		}
	}

	return &SQLLog{LogSQL: logSQL, Tracer: tracer}
}

func (s *stdoutLogger) Log(_ context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	logMessage := fmt.Sprintf("%s: %s", level, msg)
	if data != nil {
		logMessage = fmt.Sprintf("%s %v", logMessage, data)
	}
	s.logger.Println(logMessage)
}
