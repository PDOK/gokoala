package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONMerge_identical_json_input_should_not_result_differences(t *testing.T) {
	// given
	file, err := filepath.Abs("../testdata/ogcapi-tiles-1.bundled.json")
	if err != nil {
		t.Fatalf("can't locate testdata %v", err)
	}
	fileContent, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("can't read testdata %v", err)
	}
	fileContentJSON, err := json.Marshal(fileContent)
	if err != nil {
		t.Fatalf("can't parse testdata %v", err)
	}
	var expected []byte
	err = json.Unmarshal(fileContentJSON, &expected)
	if err != nil {
		t.Fatalf("can't turn testdata back into JSON %v", err)
	}

	// when
	actual, err := MergeJSON(fileContent, fileContent)
	if err != nil {
		t.Fatalf("JSON merge failed %v", err)
	}

	// then
	assert.JSONEq(t, string(expected), string(actual))
}
