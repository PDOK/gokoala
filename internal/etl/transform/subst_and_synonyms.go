package transform

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/PDOK/gomagpie/internal/engine/util"
)

type SubstAndSynonyms struct {
	substitutions   map[string]string
	synonyms        map[string]string
	synonymsInverse map[string]string
}

func NewSubstAndSynonyms(substitutionsFile, synonymsFile string) (*SubstAndSynonyms, error) {
	substitutions, substErr := readCsvFile(substitutionsFile)
	synonyms, synErr := readCsvFile(synonymsFile)
	return &SubstAndSynonyms{
		substitutions:   substitutions,
		synonyms:        synonyms,
		synonymsInverse: util.Inverse(synonyms),
	}, errors.Join(substErr, synErr)
}

func (s SubstAndSynonyms) generate(fieldValuesByName map[string]string) []map[string]string {
	var fieldValuesByNameWithAllValues = make(map[string][]string)
	for key, value := range fieldValuesByName {
		valueLower := strings.ToLower(value)
		// Get all substitutions
		substitutedValues := extendValues([]string{valueLower}, s.substitutions)
		// Get all synonyms for these substituted values
		// -> one way
		synonymsValuesOneWay := extendValues(substitutedValues, s.synonyms)
		// <- reverse way
		allValues := extendValues(synonymsValuesOneWay, s.synonymsInverse)
		// Create map with for each key a slice of []values
		fieldValuesByNameWithAllValues[key] = allValues
	}
	combinations := generateAllCombinations(fieldValuesByNameWithAllValues)
	return combinations
}

// Transform a map[string][]string into a []map[string]string using the cartesian product, i.e.
// - both maps have the same keys
// - values exist for all possible combinations
func generateAllCombinations(input map[string][]string) []map[string]string {
	var keys []string
	var values [][]string

	for key, vals := range input {
		keys = append(keys, key)
		values = append(values, vals)
	}

	return generateCombinations(keys, values)
}

func generateCombinations(keys []string, values [][]string) []map[string]string {
	if len(keys) == 0 || len(values) == 0 {
		return nil
	}
	result := []map[string]string{{}} // contains empty map so the first iteration works
	for keyDepth := 0; keyDepth < len(keys); keyDepth++ {
		var newResult []map[string]string
		for _, entry := range result {
			for _, val := range values[keyDepth] {
				newEntry := make(map[string]string)
				for k, v := range entry {
					newEntry[k] = v
				}
				newEntry[keys[keyDepth]] = val
				newResult = append(newResult, newEntry)
			}
		}
		result = newResult
	}
	return result
}

func extendValues(input []string, mapping map[string]string) []string {
	var results []string

	for len(input) > 0 {
		// Pop the first element from the input slice
		current := input[0]
		input = input[1:]

		// Add the current string to the results
		results = append(results, current)

		// Generate new strings based on the mapping
		for oldChar, newChar := range mapping {
			if strings.Contains(current, oldChar) {
				for i := 0; i < strings.Count(current, oldChar); i++ {
					if strings.HasPrefix(newChar, oldChar) {
						// skip to prevent endless loop for cases such as
						// oldChar = "foo", newChar = "foos" and input = "foosball", which would otherwise result in "foosssssssssssssssball"
						continue
					}
					extendedInput := replaceNth(current, oldChar, newChar, i+1)
					input = append(input, extendedInput)
				}
			}
		}
	}

	// Possible performance improvement here by avoiding duplicates in the first place
	return uniqueSlice(results)
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

func readCsvFile(filepath string) (map[string]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV file: %w", err)
	}

	result := make(map[string]string)
	for _, row := range records {
		result[row[0]] = row[1]
	}
	return result, nil
}
