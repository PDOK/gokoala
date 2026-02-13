package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid UTF-8 string",
			input:    "Hello, World!",
			expected: true,
		},
		{
			name:     "string with control character",
			input:    "Hello\x00World",
			expected: false,
		},
		{
			name:     "valid empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "invalid UTF-8 string",
			input:    string([]byte{0xff, 0xfe, 0xfd}),
			expected: false,
		},
		{
			name:     "string with only printable characters",
			input:    "Hello, World!",
			expected: true,
		},
		{
			name:     "string with newline character",
			input:    "Hello\nWorld",
			expected: false,
		},
		{
			name:     "string with tab character",
			input:    "Hello\tWorld",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidString(tt.input)
			assert.Equal(t, tt.expected, result, "IsValidString(%q)", tt.input)
		})
	}
}
