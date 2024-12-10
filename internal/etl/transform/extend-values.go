package transform

import (
	"encoding/csv"
	"os"
	"strings"
)

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

	return generateCombinations(keys, values, 0, make(map[string]any))
}

// Recursively generate all combinations
func generateCombinations(keys []string, values [][]string, keyDepth int, current map[string]any) []map[string]any {
	var result []map[string]any
	if keyDepth == len(keys) {
		newEntry := make(map[string]any)
		for k, v := range current {
			newEntry[k] = v
		}
		return []map[string]any{newEntry}
	}

	for _, val := range values[keyDepth] {
		current[keys[keyDepth]] = val
		partialResult := generateCombinations(keys, values, keyDepth+1, current)
		result = append(result, partialResult...)
	}

	return result
}

func applySubstitutions(input string, substitutions map[string]string) ([]string, error) {
	inputLower := strings.ToLower(input)
	var results []string
	results = append(results, inputLower)

	for oldChar, newChar := range substitutions {
		if strings.Contains(inputLower, oldChar) {
			for i := 0; i < strings.Count(inputLower, oldChar); i++ {
				substituded, err := replaceNth(inputLower, oldChar, newChar, i+1)
				if err != nil {
					return nil, err
				}
				subCombinations, err := applySubstitutions(substituded, substitutions)
				if err != nil {
					return nil, err
				}
				results = append(results, subCombinations...)
			}
		}
	}

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

func readSubstitutionsFile(filepath string) (map[string]string, error) {
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
