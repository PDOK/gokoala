package util

import (
	"bytes"
	"encoding/json"
	"log"
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
	var j1 interface{}
	err := json.Unmarshal(x1, &j1)
	if err != nil {
		return nil, err
	}
	var j2 interface{}
	err = json.Unmarshal(x2, &j2)
	if err != nil {
		return nil, err
	}
	merged := merge(j1, j2)
	return json.Marshal(merged)
}

func merge(x1, x2 interface{}) interface{} {
	switch x1 := x1.(type) {
	case map[string]interface{}:
		x2, ok := x2.(map[string]interface{})
		if !ok {
			return x1
		}
		for k, v2 := range x2 {
			if v1, ok := x1[k]; ok {
				x1[k] = merge(v1, v2)
			} else {
				x1[k] = v2
			}
		}
	case []string:
		x2, ok := x2.([]string)
		if !ok {
			return x1
		}
		x1 = append(x1, x2...)
		return removeDuplicates(x1)
	case []int:
		x2, ok := x2.([]int)
		if !ok {
			return x1
		}
		x1 = append(x1, x2...)
		return removeDuplicates(x1)
	case nil:
		x2, ok := x2.(map[string]interface{})
		if ok {
			return x2
		}
	}
	return x1
}

func removeDuplicates[T string | int | bool](input []T) []T {
	allKeys := make(map[T]bool)
	var result []T
	for _, item := range input {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			result = append(result, item)
		}
	}
	return result
}
