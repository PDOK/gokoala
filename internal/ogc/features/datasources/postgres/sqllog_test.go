package postgres

import (
	"testing"

	"github.com/PDOK/gokoala/internal/ogc/features/datasources/common"
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
				t.Setenv(common.EnvLogSQL, tt.logSQL)
			} else {
				t.Setenv(common.EnvLogSQL, "")
			}
			got := NewSQLLogFromEnv()
			assert.Equal(t, tt.wantLogSQL, got.LogSQL)
		})
	}
}
