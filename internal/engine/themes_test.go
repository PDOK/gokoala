package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStaticEndpoint(t *testing.T) {
	// given
	engine, err := NewEngine("internal/engine/testdata/config_minimal.yaml", "internal/engine/testdata/test_theme.yaml", "", false, true)
	assert.NoError(t, err)

	tests := []struct {
		input    string
		expected string
	}{
		{"fake/unrelative/path.file", "/fake/unrelative/path.file"},
		{"./fake/relative/path.file", "/fake/relative/path.file"},
		{"/fake/absolute/path.file", "/fake/absolute/path.file"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, newStaticEndppoint(engine, tt.input))
		})
	}
}
