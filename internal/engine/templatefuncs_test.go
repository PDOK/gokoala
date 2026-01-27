package engine

import (
	"html/template"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarkdown(t *testing.T) {
	tests := []struct {
		input    *string
		expected template.HTML
	}{
		{nil, ""},
		{ptrTo("**bold**"), "<p><strong>bold</strong></p>\n"},
		{ptrTo("# Heading"), "<h1>Heading</h1>\n"},
		{ptrTo("Some [link](https://example.com)"), "<p>Some <a href=\"https://example.com\" target=\"_blank\">link</a></p>\n"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, markdown(tt.input))
		})
	}
}

func TestUnmarkdown(t *testing.T) {
	tests := []struct {
		input    *string
		expected string
	}{
		{nil, ""},
		{ptrTo("**bold**"), "bold"},
		{ptrTo("# Heading"), "Heading"},
		{ptrTo("Some [link](https://example.com)"), "Some link"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, unmarkdown(tt.input))
		})
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		input    string
		limit    int
		expected string
	}{
		{"This text is not too long.", 50, "This text is not too long."},
		{"", 50, ""},
		{"This text is longer than the configured limit allows it to be.", 50, "This text is longer than the configured limit..."},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, *truncateText(&tt.input, tt.limit))
		})
	}
}

func TestTruncateSlice(t *testing.T) {
	tests := []struct {
		input    string
		limit    int
		expected string
	}{
		{"This text is not too long.", 50, "This text is not too long."},
		{"", 50, ""},
		{"This text is longer than the configured limit allows it to be.", 50, "This text is longer than the configured limit..."},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, truncateSlice(tt.input, tt.limit))
		})
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{int64(1000), "1kB"},
		{float64(1000), "1kB"},
		{1000.00, "1kB"},
		{"1000", "1kB"},
		{"1000000", "1MB"},
		{"invalid", "0"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, humanSize(tt.input))
		})
	}
}

func TestBytesSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1 kB", 1000},
		{"1 MB", 1000000},
		{"1.1 GB", 1100000000},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, bytesSize(tt.input))
		})
	}
}

func TestIsDate(t *testing.T) {
	tests := []struct {
		input    any
		expected bool
	}{
		{time.Now(), true},
		{"not a date", false},
		{12345, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, isDate(tt.input))
		})
	}
}

func TestIsLink(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "Valid HTTP URL",
			input:    "http://example.com",
			expected: true,
		},
		{
			name:     "Valid HTTPS URL",
			input:    "https://example.com",
			expected: true,
		},
		{
			name:     "Invalid URL without scheme",
			input:    "example.com",
			expected: false,
		},
		{
			name:     "Invalid string with no URL",
			input:    "not a url",
			expected: false,
		},
		{
			name:     "Non-string input (integer)",
			input:    12345,
			expected: false,
		},
		{
			name:     "Non-string input (struct)",
			input:    struct{}{},
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "URL with trailing characters",
			input:    "http://example.com foo bar",
			expected: false,
		},
		{
			name:     "URL with leading characters",
			input:    "foo bar http://example.com",
			expected: false,
		},
		{
			name:     "URL with special characters",
			input:    "http://example.com/path?query=param#fragment",
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isLink(tt.input))
		})
	}
}

func TestIsStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "valid string slice",
			input:    []string{"foo", "bar"},
			expected: true,
		},
		{
			name:     "empty string slice",
			input:    []string{},
			expected: true,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: false,
		},
		{
			name:     "string input",
			input:    "foo",
			expected: false,
		},
		{
			name:     "int slice",
			input:    []int{1, 2},
			expected: false,
		},
		{
			name:     "interface slice containing strings",
			input:    []any{"foo", "bar"},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isStringSlice(tt.input))
		})
	}
}

func TestHasField(t *testing.T) {
	type sample struct {
		Exported   string
		unexported int
	}

	var nilSamplePtr *sample = nil

	tests := []struct {
		name      string
		structRef any
		fieldName string
		want      bool
	}{
		{
			name:      "struct value - existing exported field",
			structRef: sample{Exported: "x"},
			fieldName: "Exported",
			want:      true,
		},
		{
			name:      "struct value - existing unexported field",
			structRef: sample{unexported: 1},
			fieldName: "unexported",
			want:      true,
		},
		{
			name:      "struct value - missing field",
			structRef: sample{},
			fieldName: "DoesNotExist",
			want:      false,
		},
		{
			name:      "pointer to struct - existing field",
			structRef: &sample{Exported: "x"},
			fieldName: "Exported",
			want:      true,
		},
		{
			name:      "pointer to struct (nil)",
			structRef: nilSamplePtr,
			fieldName: "Exported",
			want:      false,
		},
		{
			name:      "non-struct (int)",
			structRef: 123,
			fieldName: "Anything",
			want:      false,
		},
		{
			name:      "pointer to non-struct",
			structRef: func() any { x := 1; return &x }(),
			fieldName: "Anything",
			want:      false,
		},
		{
			name:      "nil interface",
			structRef: nil,
			fieldName: "Anything",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasField(tt.structRef, tt.fieldName))
		})
	}
}

func ptrTo(s string) *string {
	return &s
}
