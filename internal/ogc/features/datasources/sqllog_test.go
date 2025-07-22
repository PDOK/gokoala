package datasources

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSQLLogFromEnv(t *testing.T) {
	tests := []struct {
		name       string
		logSQL     string
		wantLogSQL bool
	}{
		{"logSql is set to false", "false", false},
		{"logSql is set to true", "true", true},
		{"logSql is not set", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.logSQL != "" {
				t.Setenv(envLogSQL, tt.logSQL)
			} else {
				t.Setenv(envLogSQL, "")
			}
			got := NewSQLLogFromEnv()
			assert.Equal(t, tt.wantLogSQL, got.LogSQL)
		})
	}
}

func TestSQLLog_CheckLogMessageWhenExplicitEnabled(t *testing.T) {
	var capturedLogOutput bytes.Buffer
	log.SetOutput(&capturedLogOutput)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	s := &SQLLog{LogSQL: true}

	ctx, err := s.Before(t.Context(), "SELECT * FROM test WHERE id = ?", 123)
	assert.NoError(t, err)

	_, err = s.After(ctx, "SELECT * FROM test WHERE id = ?", 123)
	assert.NoError(t, err)

	assert.Contains(t, capturedLogOutput.String(), "SQL:\nSELECT * FROM test WHERE id = 123\n--- SQL query took")
}
