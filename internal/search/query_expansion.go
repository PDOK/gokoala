package search

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/PDOK/gomagpie/internal/search/domain"
)

// QueryExpansion query expansion involves evaluating a user's input (what words were typed into the search query area)
// and expanding the search query to match additional results, see https://en.wikipedia.org/wiki/Query_expansion
type QueryExpansion struct {
	rewrites map[string][]string
	synonyms map[*regexp.Regexp][]string
}

func NewQueryExpansion(rewritesFile, synonymsFile string) (*QueryExpansion, error) {
	rewrites, rewErr := readCsvFile(rewritesFile, false)
	synonyms, synErr := readCsvFile(synonymsFile, true)

	// Turn keys into regexes for efficient matching of whole words
	synonymsWithRegexKey := make(map[*regexp.Regexp][]string)
	for k, v := range synonyms {
		wholeWordRegex, err := regexp.Compile(`\b` + regexp.QuoteMeta(k) + `\b`)
		if err != nil {
			return nil, err
		}
		synonymsWithRegexKey[wholeWordRegex] = v
	}

	return &QueryExpansion{
		rewrites: rewrites,
		synonyms: synonymsWithRegexKey,
	}, errors.Join(rewErr, synErr)
}

// Expand Perform query expansion, see https://en.wikipedia.org/wiki/Query_expansion
func (s QueryExpansion) Expand(searchTerms string) domain.SearchQuery {
	result := rewrite(searchTerms, s.rewrites)
	results := expandSynonyms(result, s.synonyms)
	return domain.NewSearchQuery(results)
}

func rewrite(input string, mapping map[string][]string) string {
	terms := strings.ToLower(input)
	for original, alternatives := range mapping {
		for _, alternative := range alternatives {
			terms = strings.ReplaceAll(terms, alternative, original)
		}
	}
	return terms
}

func expandSynonyms(input string, mapping map[*regexp.Regexp][]string) []string {
	results := []string{input}

	for original, alternatives := range mapping {
		currentResults := make([]string, 0)
		for _, existing := range results {
			if original.MatchString(existing) {
				for _, alternative := range alternatives {
					// Replace only complete words to avoid situations like:
					// original = "foo", alternative = "foos" and input = "foosball",
					// which would otherwise result in "foossball"
					updated := original.ReplaceAllString(existing, alternative)
					currentResults = append(currentResults, updated)
				}
			}
		}
		results = append(results, currentResults...)
	}

	return uniqueSlice(results)
}

func uniqueSlice(s []string) []string {
	var results []string
	seen := make(map[string]bool)
	for _, entry := range s {
		if !seen[entry] {
			seen[entry] = true
			results = append(results, entry)
		}
	}
	return results
}

func readCsvFile(filepath string, bidi bool) (map[string][]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // allow variable number of columns per row, also allow blank lines
	reader.Comment = '#'        // allow comments in CSV

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV file: %w", err)
	}

	result := make(map[string][]string)
	for _, row := range records {
		key := strings.ToLower(row[0])
		result[key] = make([]string, 0)

		// add all alternatives
		for i := 1; i < len(row); i++ {
			result[key] = append(result[key], strings.ToLower(row[i]))
		}

		if bidi {
			// make result map bidirectional
			for _, alt := range result[key] {
				if _, exists := result[alt]; !exists {
					result[alt] = make([]string, 0)
				}
				result[alt] = append(result[alt], key)
				for _, other := range result[key] {
					if other != alt { // skip self
						result[alt] = append(result[alt], other)
					}
				}
			}
		}
	}
	return result, nil
}
