package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		srid     SRID
		expected int
	}{
		{"Positive SRID", SRID(28992), 28992},
		{"Zero SRID", SRID(0), WGS84SRID},
		{"Negative SRID", SRID(-1), WGS84SRID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.srid.GetOrDefault())
		})
	}
}

func TestEpsgToSrid(t *testing.T) {
	tests := []struct {
		name        string
		srs         string
		expected    SRID
		expectError bool
	}{
		{"Valid EPSG", "EPSG:28992", SRID(28992), false},
		{"Invalid prefix", "INVALID:28992", SRID(-1), true},
		{"Non-numeric EPSG code", "EPSG:ABC", SRID(-1), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EpsgToSrid(tt.srs)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
