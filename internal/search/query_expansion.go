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

// Position is a substring match in the given search term
type Position struct {
	start       int
	length      int
	alternative string
}

func expandSynonyms(input string, mapping map[string][]string) []string {
	results := []string{input}
	continueExpanding := true

	for continueExpanding {
		var currentResults []string
		continueExpanding = false

		for _, variant := range results {
			positions := mapPositions(variant, mapping)

			// sort by longest length, when equal by smallest start position
			sort.Slice(positions, func(i, j int) bool {
				if positions[i].length != positions[j].length {
					return positions[i].length > positions[j].length
				}
				return positions[i].start < positions[j].start
			})

			for _, newVariant := range generateCandidateVariants(variant, positions) {
				if !slices.Contains(results, newVariant) && !slices.Contains(currentResults, newVariant) {
					currentResults = append(currentResults, newVariant)
					continueExpanding = true
				}
			}
		}
		results = append(results, currentResults...)
	}
	return results
}

func mapPositions(input string, mapping map[string][]string) []Position {
	var results []Position
	words := strings.Fields(input)
	wordsPos := 0

	for _, word := range words {
		// try to match whole words first
		if alternatives, exists := mapping[word]; exists {
			for _, alternative := range alternatives {
				results = append(results, Position{
					start:       wordsPos,
					length:      len(word),
					alternative: alternative,
				})
			}
		} else {
			// then try to find matches within the word
			for original, alternatives := range mapping {
				for pos := 0; pos < len(word); {
					originalPos := strings.Index(word[pos:], original)
					if originalPos == -1 {
						break
					}

					actualPos := wordsPos + pos + originalPos
					for _, alternative := range alternatives {
						results = append(results, Position{
							start:       actualPos,
							length:      len(original),
							alternative: alternative,
						})
					}
					pos += originalPos + 1
				}
			}
		}
		wordsPos += len(word) + 1 // +1 for the space
	}
	return results
}

func generateCandidateVariants(input string, positions []Position) []string {
	var results []string
	for _, pos := range positions {
		if !hasOverlap(pos, positions) {
			variant := input[:pos.start] + pos.alternative + input[pos.start+pos.length:]
			results = append(results, variant)
		}
	}
	return results
}

func hasOverlap(current Position, all []Position) bool {
	currentEnd := current.start + current.length
	for _, other := range all {
		if other.length <= current.length {
			continue
		}

		otherEnd := other.start + other.length
		if current.start < otherEnd && other.start < currentEnd {
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
