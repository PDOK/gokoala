package domain

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamesAndVersionsAndRelevance(t *testing.T) {
	tests := []struct {
		name          string
		input         CollectionsWithParams
		expectedNames []string
		expectedVers  []int
		expectedRel   []float64
	}{
		{
			name: "Valid input with relevance",
			input: CollectionsWithParams{
				"collection1": {VersionParam: "1", RelevanceParam: "0.8"},
				"collection2": {VersionParam: "2", RelevanceParam: "0.5"},
			},
			expectedNames: []string{"collection1", "collection2"},
			expectedVers:  []int{1, 2},
			expectedRel:   []float64{0.8, 0.5},
		},
		{
			name: "Valid input without relevance",
			input: CollectionsWithParams{
				"collection1": {VersionParam: "1"},
				"collection2": {VersionParam: "2"},
			},
			expectedNames: []string{"collection1", "collection2"},
			expectedVers:  []int{1, 2},
			expectedRel:   []float64{DefaultRelevance, DefaultRelevance},
		},
		{
			name: "Invalid version",
			input: CollectionsWithParams{
				"collection1": {VersionParam: "invalid"},
				"collection2": {VersionParam: "2"},
			},
			expectedNames: []string{"collection2"},
			expectedVers:  []int{2},
			expectedRel:   []float64{DefaultRelevance},
		},
		{
			name: "Invalid relevance",
			input: CollectionsWithParams{
				"collection1": {VersionParam: "1", RelevanceParam: "invalid"},
				"collection2": {VersionParam: "2", RelevanceParam: "-1"},
				"collection3": {VersionParam: "3", RelevanceParam: "2"},
			},
			expectedNames: []string{"collection1", "collection2", "collection3"},
			expectedVers:  []int{1, 2, 3},
			expectedRel:   []float64{DefaultRelevance, DefaultRelevance, DefaultRelevance},
		},
		{
			name: "Missing version parameter",
			input: CollectionsWithParams{
				"collection1": {RelevanceParam: "0.8"},
				"collection2": {VersionParam: "2", RelevanceParam: "0.5"},
			},
			expectedNames: []string{"collection2"},
			expectedVers:  []int{2},
			expectedRel:   []float64{0.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualNames, actualVers, actualRel := tt.input.NamesAndVersionsAndRelevance()

			// Sort all slices by collection name to ensure consistent order
			type result struct {
				name string
				ver  int
				rel  float64
			}
			results := make([]result, len(actualNames))
			for i := range actualNames {
				results[i] = result{actualNames[i], actualVers[i], actualRel[i]}
			}

			// Sort by name using sort.Slice
			sort.Slice(results, func(i, j int) bool {
				return results[i].name < results[j].name
			})

			// Extract sorted values
			actualNames = make([]string, len(results))
			actualVers = make([]int, len(results))
			actualRel = make([]float64, len(results))
			for i, r := range results {
				actualNames[i] = r.name
				actualVers[i] = r.ver
				actualRel[i] = r.rel
			}
			assert.Equal(t, tt.expectedNames, actualNames)
			assert.Equal(t, tt.expectedVers, actualVers)
			assert.Equal(t, tt.expectedRel, actualRel)
		})
	}
}
