package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToWildcardQuery(t *testing.T) {
	tests := []struct {
		name             string
		words            []string
		withoutSynonyms  map[string]struct{}
		withSynonyms     map[string][]string
		expectedWildcard string
	}{
		{
			name:             "empty words",
			words:            []string{},
			withoutSynonyms:  map[string]struct{}{},
			withSynonyms:     map[string][]string{},
			expectedWildcard: "",
		},
		{
			name:             "single word without synonym",
			words:            []string{"foo"},
			withoutSynonyms:  map[string]struct{}{"foo": {}},
			withSynonyms:     map[string][]string{},
			expectedWildcard: "foo:*",
		},
		{
			name:             "single word with synonyms",
			words:            []string{"bar"},
			withoutSynonyms:  map[string]struct{}{},
			withSynonyms:     map[string][]string{"bar": {"baz", "qux"}},
			expectedWildcard: "(bar:* | baz:* | qux:*)",
		},
		{
			name:  "multiple words with mixed settings",
			words: []string{"foo", "bar", "baz"},
			withoutSynonyms: map[string]struct{}{
				"foo": {},
			},
			withSynonyms: map[string][]string{
				"bar": {"baz", "qux"},
				"baz": {"quux"},
			},
			expectedWildcard: "foo:* & (bar:* | baz:* | qux:*) & (baz:* | quux:*)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := NewSearchQuery(tt.words, tt.withoutSynonyms, tt.withSynonyms)
			assert.Equal(t, tt.expectedWildcard, query.ToWildcardQuery())
		})
	}
}

func TestToExactMatchQuery(t *testing.T) {
	tests := []struct {
		name               string
		words              []string
		useSynonyms        bool
		withoutSynonyms    map[string]struct{}
		withSynonyms       map[string][]string
		expectedExactMatch string
	}{
		{
			name:               "empty words",
			words:              []string{},
			useSynonyms:        false,
			withoutSynonyms:    map[string]struct{}{},
			withSynonyms:       map[string][]string{},
			expectedExactMatch: "",
		},
		{
			name:               "single word without synonym",
			words:              []string{"foo"},
			useSynonyms:        false,
			withoutSynonyms:    map[string]struct{}{"foo": {}},
			withSynonyms:       map[string][]string{},
			expectedExactMatch: "foo",
		},
		{
			name:               "single word with synonyms and useSynonyms = false",
			words:              []string{"bar"},
			useSynonyms:        false,
			withoutSynonyms:    map[string]struct{}{},
			withSynonyms:       map[string][]string{"bar": {"baz", "qux"}},
			expectedExactMatch: "(bar)",
		},
		{
			name:               "single word with synonyms and useSynonyms = true",
			words:              []string{"bar"},
			useSynonyms:        true,
			withoutSynonyms:    map[string]struct{}{},
			withSynonyms:       map[string][]string{"bar": {"baz", "qux"}},
			expectedExactMatch: "(bar | baz | qux)",
		},
		{
			name:        "multiple words with mixed settings and useSynonyms = true",
			words:       []string{"foo", "bar", "baz"},
			useSynonyms: true,
			withoutSynonyms: map[string]struct{}{
				"foo": {},
			},
			withSynonyms: map[string][]string{
				"bar": {"baz", "qux"},
				"baz": {"quux"},
			},
			expectedExactMatch: "foo & (bar | baz | qux) & (baz | quux)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := NewSearchQuery(tt.words, tt.withoutSynonyms, tt.withSynonyms)
			assert.Equal(t, tt.expectedExactMatch, query.ToExactMatchQuery(tt.useSynonyms))
		})
	}
}
