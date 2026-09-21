package features_search

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSearchTermsRejectsOperators(t *testing.T) {
	tests := []string{
		"foo|bar",
		"!foo",
		"foo<->bar",
		"foo<2>bar",
	}

	for _, searchTerms := range tests {
		t.Run(searchTerms, func(t *testing.T) {
			actual, err := parseSearchTerms(url.Values{queryParam: {searchTerms}})

			assert.Empty(t, actual)
			require.EqualError(t, err, "provided search terms contain one ore more boolean operators which aren't allowed")
		})
	}
}

func TestParseSearchTermsRejectsInvalidCharacters(t *testing.T) {
	tests := []struct {
		searchTerms   string
		expectedError string
	}{
		{"foo:bar", "provided search terms contain invalid character ':'"},
		{"foo*bar", "provided search terms contain invalid character '*'"},
		{"foo<bar", "provided search terms contain invalid character '<'"},
		{"foo>bar", "provided search terms contain invalid character '>'"},
		{`foo\bar`, `provided search terms contain invalid character '\\'`},
		{"foo\x00bar", `provided search terms contain invalid character '\x00'`},
	}

	for _, tt := range tests {
		t.Run(tt.searchTerms, func(t *testing.T) {
			actual, err := parseSearchTerms(url.Values{queryParam: {tt.searchTerms}})

			assert.Empty(t, actual)
			require.EqualError(t, err, tt.expectedError)
		})
	}
}

func TestParseSearchTermsAllowsAmpersandsAndApostrophes(t *testing.T) {
	tests := []string{
		"'s-Gravenhage",
		"Janssen & Fritsenplein",
	}

	for _, searchTerms := range tests {
		t.Run(searchTerms, func(t *testing.T) {
			actual, err := parseSearchTerms(url.Values{queryParam: {searchTerms}})

			require.NoError(t, err)
			assert.Equal(t, strings.ToLower(searchTerms), actual)
		})
	}
}
