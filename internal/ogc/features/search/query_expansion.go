package search

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

// QueryExpansion query expansion involves evaluating a user's input (what words were typed into the search query area)
// and expanding the search query to match additional results, see https://en.wikipedia.org/wiki/Query_expansion
type QueryExpansion struct {
	rewrites map[string][]string
	synonyms map[string][]string
}

func NewQueryExpansion(rewritesFile, synonymsFile string) (*QueryExpansion, error) {
	if rewritesFile == "" && synonymsFile == "" {
		return nil, nil
	}
	rewrites, rewErr := readCsvFile(rewritesFile, false)
	synonyms, synErr := readCsvFile(synonymsFile, true)

	// avoid too short synonyms to prevent to many invalid synonym/combinations
	for k, v := range synonyms {
		if err := assertSynonymLength(k); err != nil {
			return nil, err
		}
		for _, variant := range v {
			if err := assertSynonymLength(variant); err != nil {
				return nil, err
			}
		}
	}

	return &QueryExpansion{
		rewrites: rewrites,
		synonyms: synonyms,
	}, errors.Join(rewErr, synErr)
}

// Expand Perform query expansion, see https://en.wikipedia.org/wiki/Query_expansion
func (s QueryExpansion) Expand(ctx context.Context, searchTerms string) (*domain.SearchQuery, error) {
	expandCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	rewritten, err := rewrite(expandCtx, strings.ToLower(searchTerms), s.rewrites)
	if err != nil {
		return nil, err
	}
	words, wordsWithoutSynonyms, wordsWithSynonyms, err := expandSynonyms(expandCtx, rewritten, s.synonyms)
	if err != nil {
		return nil, err
	}
	return domain.NewSearchQuery(words, wordsWithoutSynonyms, wordsWithSynonyms), expandCtx.Err()
}

func rewrite(ctx context.Context, input string, mapping map[string][]string) (string, error) {
	for original, alternatives := range mapping {
		for _, alternative := range alternatives {
			input = strings.ReplaceAll(input, alternative, original)
		}
	}
	return input, ctx.Err()
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

func expandSynonyms(ctx context.Context, input string, mapping map[string][]string) ([]string, map[string]struct{},
	map[string][]string, error) {

	words := uniqueSlice(strings.Fields(input))

	wordsWithSynonyms := make(map[string][]string)
	for _, word := range words {
		variants := []string{word}
		for i := 0; i < len(variants); i++ {
			existingVariant := variants[i]
			positions := mapPositions(existingVariant, mapping)

			// sort by longest length, when equal by smallest start position
			sort.Slice(positions, func(i, j int) bool {
				if positions[i].length != positions[j].length {
					return positions[i].length > positions[j].length
				}
				return positions[i].start < positions[j].start
			})

			for _, newVariant := range generateNewVariants(existingVariant, positions) {
				if err := ctx.Err(); err != nil {
					return nil, nil, nil, err // timeout encountered
				}
				if !slices.Contains(variants, newVariant) {
					variants = append(variants, newVariant) // continue for-loop by appending to slice
					wordsWithSynonyms[word] = append(wordsWithSynonyms[word], newVariant)
				}
			}
		}
	}

	wordsWithoutSynonyms := make(map[string]struct{})
	for _, word := range words {
		if _, ok := wordsWithSynonyms[word]; ok {
			continue
		}
		wordsWithoutSynonyms[word] = struct{}{}
	}
	return words, wordsWithoutSynonyms, wordsWithSynonyms, ctx.Err()
}

// replaces all duplicates in a slice (note: slices.compact() only replaces consecutive duplicates).
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

func mapPositions(input string, mapping map[string][]string) []position {
	var results []position

	for original, alternatives := range mapping {
		for i := 0; i < len(input); {
			originalPos := strings.Index(input[i:], original)
			if originalPos == -1 {
				break
			}

			for _, alternative := range alternatives {
				results = append(results, position{
					start:       i + originalPos,
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

		// add all alternatives
		result[key] = make([]string, 0)
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

func assertSynonymLength(syn string) error {
	if len(syn) < 2 {
		return fmt.Errorf("failed to parse CSV file: synonym '%s' is too short, should be at least 2 chars long", syn)
	}
	return nil
}
