package datasources

import (
	"bytes"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSQLLogFromEnv(t *testing.T) {
	tests := []struct {
		name          string
		logSQL        string
		slowQueryTime string
		wantLogSQL    bool
		wantDuration  time.Duration
	}{
		{"Both environment variables are set", "true", "5s", true, 5 * time.Second},
		{"Only logSql is set", "false", "", false, 5 * time.Second},
		{"Only slowQueryTime is set", "", "10s", false, 10 * time.Second},
		{"Neither environment variables are set", "", "", false, 5 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.logSQL != "" {
				t.Setenv(envLogSQL, tt.logSQL)
			} else {
				t.Setenv(envLogSQL, "")
			}
			if tt.slowQueryTime != "" {
				t.Setenv(envSlowQueryTime, tt.slowQueryTime)
			} else {
				t.Setenv(envSlowQueryTime, "")
			}
			got := NewSQLLogFromEnv()
			assert.Equal(t, tt.wantLogSQL, got.LogSQL)
			assert.Equal(t, tt.wantDuration, got.SlowQueryTime)
		})
	}
}

func TestSQLLog_CheckLogMessageWhenExplicitEnabled(t *testing.T) {
	var capturedLogOutput bytes.Buffer
	log.SetOutput(&capturedLogOutput)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	s := &SQLLog{LogSQL: true, SlowQueryTime: 10 * time.Hour}

	ctx, err := s.Before(t.Context(), "SELECT * FROM test WHERE id = ?", 123)
	assert.NoError(t, err)

	_, err = s.After(ctx, "SELECT * FROM test WHERE id = ?", 123)
	assert.NoError(t, err)

	assert.Contains(t, capturedLogOutput.String(), "SQL:\nSELECT * FROM test WHERE id = 123\n--- SQL query took")
}

func TestSQLLog_CheckLogInCaseOfSlowQuery(t *testing.T) {
	var capturedLogOutput bytes.Buffer
	log.SetOutput(&capturedLogOutput)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	s := &SQLLog{LogSQL: false, SlowQueryTime: 1 * time.Nanosecond}

	ctx, err := s.Before(t.Context(), "SELECT * FROM test WHERE id = ?", 123)
	assert.NoError(t, err)

	_, err = s.After(ctx, "SELECT * FROM test WHERE id = ?", 123)
	assert.NoError(t, err)

	assert.Contains(t, capturedLogOutput.String(), "SQL:\nSELECT * FROM test WHERE id = 123\n--- SQL query took")
}
