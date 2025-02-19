package search

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/PDOK/gomagpie/internal/search/domain"
)

// QueryExpansion query expansion involves evaluating a user's input (what words were typed into the search query area)
// and expanding the search query to match additional results, see https://en.wikipedia.org/wiki/Query_expansion
type QueryExpansion struct {
	rewrites map[string][]string
	synonyms map[string][]string
}

func NewQueryExpansion(rewritesFile, synonymsFile string) (*QueryExpansion, error) {
	rewrites, rewErr := readCsvFile(rewritesFile, false)
	synonyms, synErr := readCsvFile(synonymsFile, true)

	return &QueryExpansion{
		rewrites: rewrites,
		synonyms: synonyms,
	}, errors.Join(rewErr, synErr)
}

// Expand Perform query expansion, see https://en.wikipedia.org/wiki/Query_expansion
func (s QueryExpansion) Expand(searchTerms string) domain.SearchQuery {
	result := rewrite(strings.ToLower(searchTerms), s.rewrites)
	results := expandSynonyms(result, s.synonyms)
	return domain.NewSearchQuery(results)
}

func rewrite(input string, mapping map[string][]string) string {
	for original, alternatives := range mapping {
		for _, alternative := range alternatives {
			input = strings.ReplaceAll(input, alternative, original)
		}
	}
	return input
}

// position is a substring match in the given search term
type position struct {
	start       int
	length      int
	alternative string
}

func (p position) end() int {
	return p.start + p.length
}

func (p position) replace(input string) string {
	return input[:p.start] + p.alternative + input[p.end():]
}

func expandSynonyms(input string, mapping map[string][]string) []string {
	results := []string{input}

	for i := 0; i < len(results); i++ {
		existing := results[i]
		positions := mapPositions(existing, mapping)

		// sort by longest length, when equal by smallest start position
		sort.Slice(positions, func(i, j int) bool {
			if positions[i].length != positions[j].length {
				return positions[i].length > positions[j].length
			}
			return positions[i].start < positions[j].start
		})

		for _, newVariant := range generateNewVariants(existing, positions) {
			if !slices.Contains(results, newVariant) {
				results = append(results, newVariant)
			}
		}
	}
	return results
}

func mapPositions(input string, mapping map[string][]string) []position {
	var results []position

	for original, alternatives := range mapping {
		for i := 0; i < len(input); {
			originalPos := strings.Index(input[i:], original)
			if originalPos == -1 {
				break
			}

			actualPos := i + originalPos
			for _, alternative := range alternatives {
				results = append(results, position{
					start:       actualPos,
					length:      len(original),
					alternative: alternative,
				})
			}
			i += originalPos + 1
		}
	}
	return results
}

func generateNewVariants(input string, positions []position) []string {
	var results []string
	for _, pos := range positions {
		if !hasOverlap(pos, positions) {
			results = append(results, pos.replace(input))
		}
	}
	return results
}

// We need to check for overlapping synonyms for situations like:
//
// synonyms = goeverneur,goev,gouverneur,gouv
// input = 1e gouverneurstraat
// synonyms key (original) => gouv
// synonyms value (alternative) = goeverneur
// resulting string = 1e goeverneurERNEURstraat <-- not what we want
func hasOverlap(current position, all []position) bool {
	for _, other := range all {
		if other.length <= current.length {
			continue
		}
		if current.start < other.end() && other.start < current.end() {
			return true
		}
	}
	return false
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
			// make result map bidirectional, so:
			// 1e => one,first | 2e => second
			// becomes:
			// 1e => one,first | 2e => second | one => 1e,first | first => 1e,one | second => 2e
			for _, alt := range result[key] {
				if _, ok := result[alt]; !ok {
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
