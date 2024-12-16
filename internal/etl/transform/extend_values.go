package transform

import (
	"encoding/csv"
	"github.com/PDOK/gomagpie/internal/engine/util"
	"os"
	"strings"
)

// Return slice of fieldValuesByName
func extendFieldValues(fieldValuesByName map[string]any, substitutionsFile, synonymsFile string) ([]map[string]any, error) {
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
		valueLower := strings.ToLower(value.(string))

		// Get all substitutions
		substitutedValues, err := extendValues([]string{valueLower}, substitutions)
		if err != nil {
			return nil, err
		}
		// Get all synonyms for these substituted values
		// one way
		synonymsValuesOneWay, err := extendValues(substitutedValues, synonyms)
		if err != nil {
			return nil, err
		}
		// reverse way
		allValues, err := extendValues(synonymsValuesOneWay, util.Inverse(synonyms))
		if err != nil {
			return nil, err
		}

		// Create map with for each key a slice of []values
		fieldValuesByNameWithAllValues[key] = allValues
	}
	return generateAllFieldValuesByName(fieldValuesByNameWithAllValues), err
}

// Transform a map[string][]string into a []map[string]string using the cartesian product, i.e.
// - both maps have the same keys
// - values exist for all possible combinations
func generateAllFieldValuesByName(input map[string][]string) []map[string]any {
	keys := []string{}
	values := [][]string{}

	for key, vals := range input {
		keys = append(keys, key)
		values = append(values, vals)
	}

	return generateCombinations(keys, values)
}

func generateCombinations(keys []string, values [][]string) []map[string]any {
	if len(keys) == 0 || len(values) == 0 {
		return nil
	}
	result := []map[string]any{{}} // contains empty map so the first iteration works
	for keyDepth := 0; keyDepth < len(keys); keyDepth++ {
		var newResult []map[string]any
		for _, entry := range result {
			for _, val := range values[keyDepth] {
				newEntry := make(map[string]any)
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

func extendValues(input []string, mapping map[string]string) ([]string, error) {
	var results []string
	results = append(results, input...)

	for j := range input {
		for oldChar, newChar := range mapping {
			if strings.Contains(input[j], oldChar) {
				for i := 0; i < strings.Count(input[j], oldChar); i++ {
					extendedInput, err := replaceNth(input[j], oldChar, newChar, i+1)
					if err != nil {
						return nil, err
					}
					subCombinations, err := extendValues([]string{extendedInput}, mapping)
					if err != nil {
						return nil, err
					}
					results = append(results, subCombinations...)
				}
			}
		}
	}
	// Possible performance improvement here by avoiding duplicates in the first place
	return uniqueSlice(results), nil
}

func replaceNth(input, oldChar, newChar string, nthIndex int) (string, error) {
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
		err := result.WriteByte(input[i])
		if err != nil {
			return "", err
		}
	}
	return result.String(), nil
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
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, row := range records {
		substitutions[row[0]] = row[1]
	}
	return substitutions, err
}
