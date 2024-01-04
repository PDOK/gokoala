package util

import (
	"bytes"
	"encoding/json"
	"log"

	"dario.cat/mergo"
)

func PrettyPrintJSON(content []byte, name string) []byte {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, content, "", " "); err != nil {
		log.Print(string(content))
		log.Fatalf("invalid json in %s: %v, see json output above", name, err)
	}
	return pretty.Bytes()
}

// MergeJSON merges the two JSON byte slices containing x1 and x2,
// preferring x1 over x2 except where x1 and x2 are
// JSON objects, in which case the keys from both objects
// are included and their values merged recursively.
//
// It returns an error if x1 or x2 cannot be JSON-unmarshalled,
// or the merged JSON is invalid.
func MergeJSON(x1, x2 []byte) ([]byte, error) {
	var j1 map[string]any
	err := json.Unmarshal(x1, &j1)
	if err != nil {
		return nil, err
	}
	var j2 map[string]any
	err = json.Unmarshal(x2, &j2)
	if err != nil {
		return nil, err
	}
	err = mergo.Merge(&j1, &j2)
	if err != nil {
		return nil, err
	}
	return json.Marshal(j1)
}
