package search

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"

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

func (s SubstAndSynonyms) generate(term string) [][]string {
	var result = make([][]string, 2)
	result[0] = []string{term}
	return result
}

func readCsvFile(filepath string) (map[string]string, error) {
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

	result := make(map[string]string)
	for _, row := range records {
		result[row[0]] = row[1]
	}
	return result, nil
}
