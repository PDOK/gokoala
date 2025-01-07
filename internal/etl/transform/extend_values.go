package transform

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/PDOK/gomagpie/internal/engine/util"
)

func extendFieldValues(fieldValuesByName map[string]string, substitutionsFile, synonymsFile string) ([]map[string]string, error) {
	substitutions, err := readCsvFile(substitutionsFile)
	if err != nil {
		return nil, err
	}
	synonyms, err := readCsvFile(synonymsFile)
	if err != nil {
		return nil, err
	}

	var fieldValuesByNameWithAllValues = make(map[string][]string)
	for key, value := range fieldValuesByName {
		valueLower := strings.ToLower(value)
		// Get all substitutions
		substitutedValues := extendValues([]string{valueLower}, substitutions)
		// Get all synonyms for these substituted values
		// -> one way
		synonymsValuesOneWay := extendValues(substitutedValues, synonyms)
		// <- reverse way
		allValues := extendValues(synonymsValuesOneWay, util.Inverse(synonyms))
		// Create map with for each key a slice of []values
		fieldValuesByNameWithAllValues[key] = allValues
	}
	return generateAllFieldValuesByName(fieldValuesByNameWithAllValues), err
}

// Transform a map[string][]string into a []map[string]string using the cartesian product, i.e.
// - both maps have the same keys
// - values exist for all possible combinations
func generateAllFieldValuesByName(input map[string][]string) []map[string]string {
	keys := []string{}
	values := [][]string{}

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
	results = append(results, input...)

	for j := range input {
		for oldChar, newChar := range mapping {
			if strings.Contains(input[j], oldChar) {
				for i := 0; i < strings.Count(input[j], oldChar); i++ {
					extendedInput := replaceNth(input[j], oldChar, newChar, i+1)
					subCombinations := extendValues([]string{extendedInput}, mapping)
					results = append(results, subCombinations...)
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
	substitutions := make(map[string]string)

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

	for _, row := range records {
		substitutions[row[0]] = row[1]
	}
	return substitutions, nil
}
