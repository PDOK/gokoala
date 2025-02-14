package search

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/PDOK/gomagpie/internal/engine/util"
	"github.com/PDOK/gomagpie/internal/search/domain"
)

type SubstAndSynonyms struct {
	rewrites        map[string][]string
	synonyms        map[string][]string
	synonymsInverse map[string][]string
}

func NewSubstAndSynonyms(rewritesFile, synonymsFile string) (*SubstAndSynonyms, error) {
	rewrites, rewErr := readCsvFile(rewritesFile)
	synonyms, synErr := readCsvFile(synonymsFile)
	return &SubstAndSynonyms{
		rewrites:        rewrites,
		synonyms:        synonyms,
		synonymsInverse: util.InverseMulti(synonyms),
	}, errors.Join(rewErr, synErr)
}

func (s SubstAndSynonyms) generate(term string) domain.SearchQuery {
	result := rewrite(term, s.rewrites)
	// -> one way
	result = expandSynonyms(result, s.synonyms)
	// <- reverse way
	result = expandSynonyms(result, s.synonymsInverse)
	return domain.NewSearchQuery(result)
}

func rewrite(terms string, mapping map[string][]string) []string {
	for original, alternatives := range mapping {
		for _, alternative := range alternatives {
			terms = strings.ReplaceAll(terms, alternative, original)
		}
	}
	return []string{terms}
}

func expandSynonyms(input []string, mapping map[string][]string) []string {
	var results []string

	for len(input) > 0 {
		// Pop the first element from the input slice
		current := strings.ToLower(input[0])
		input = input[1:]

		// Add the current string to the results
		results = append(results, current)

		// Generate new strings based on the mapping
		for original, alternatives := range mapping {
			for _, alternative := range alternatives {
				if strings.Contains(current, original) {
					for i := 0; i < strings.Count(current, original); i++ {
						if strings.HasPrefix(alternative, original) {
							// skip to prevent endless loop for cases such as
							// original = "foo", alternative = "foos" and input = "foosball", which would otherwise result in "foosssssssssssssssball"
							continue
						}
						extendedInput := replaceNth(current, original, alternative, i+1)
						input = append(input, extendedInput)
					}
				}
			}
		}
	}
	return slices.Compact(results)
}

func replaceNth(input, oldChar, newChar string, nthIndex int) string {
	count := 0
	result := strings.Builder{}

	for i := 0; i < len(input); i++ {
		if strings.HasPrefix(input[i:], oldChar) {
			count++
			if count == nthIndex {
				result.WriteString(newChar)
				i += len(oldChar) - 1
				continue
			}
		}
		result.WriteByte(input[i]) // no need to catch error since "the returned error is always nil"
	}
	return result.String()
}

func readCsvFile(filepath string) (map[string][]string, error) {
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
		for i := 0; i < len(row); i++ {
			// make sure it's all in lowercase since replacement happens in lowercase
			row[i] = strings.ToLower(row[i])
		}
		result[row[0]] = row[1:]
	}
	return result, nil
}
