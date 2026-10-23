package records

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/PDOK/gokoala/config"
)

type sortKey struct {
	name       string
	descending bool
}

func sortRecords(records config.Records, collection config.RecordsCollection, query map[string][]string) (config.Records, error) {
	requested, hasRequested := query["sortby"]
	sortOrder := make([]sortKey, 0)
	if hasRequested {
		allowed := make(map[string]bool)
		for _, name := range sortableNames(collection) {
			allowed[name] = true
		}
		for _, value := range requested {
			for _, rawKey := range strings.Split(value, ",") {
				key := sortKey{}
				if strings.HasPrefix(rawKey, "-") {
					key.descending = true
					key.name = strings.TrimPrefix(rawKey, "-")
				} else {
					key.name = strings.TrimPrefix(rawKey, "+")
				}
				if key.name == "" || !allowed[key.name] {
					return nil, fmt.Errorf("property %q is not sortable", key.name)
				}
				sortOrder = append(sortOrder, key)
			}
		}
	} else {
		for _, configured := range collection.DefaultSortOrder {
			sortOrder = append(sortOrder, sortKey{name: configured.Field, descending: configured.Direction == "desc"})
		}
	}
	if len(sortOrder) == 0 {
		return records, nil
	}
	result := append(config.Records(nil), records...)
	sort.SliceStable(result, func(left, right int) bool {
		for _, key := range sortOrder {
			comparison := compareSortableValues(result[left].Properties[key.name], result[right].Properties[key.name])
			if comparison == 0 {
				continue
			}
			if key.descending {
				return comparison > 0
			}
			return comparison < 0
		}
		return false
	})
	return result, nil
}

func sortableNames(collection config.RecordsCollection) []string {
	if len(collection.Sortables) > 0 {
		return append([]string(nil), collection.Sortables...)
	}
	names := make(map[string]bool)
	for _, record := range collection.Records {
		for name, value := range record.Properties {
			if isSortableValue(value) {
				names[name] = true
			}
		}
	}
	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func sortableType(collection config.RecordsCollection, name string) string {
	for _, record := range collection.Records {
		value, ok := record.Properties[name]
		if !ok {
			continue
		}
		decoded, err := value.Any()
		if err != nil || decoded == nil {
			continue
		}
		switch decoded.(type) {
		case bool:
			return "boolean"
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return "integer"
		case float32, float64:
			return "number"
		case json.Number:
			if _, err := decoded.(json.Number).Int64(); err == nil {
				return "integer"
			}
			return "number"
		case string:
			return "string"
		}
	}
	return "string"
}

func isSortableValue(value any) bool {
	if encoded, ok := value.(config.JSONValue); ok {
		decoded, err := encoded.Any()
		return err == nil && isSortableValue(decoded)
	}
	switch value.(type) {
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string:
		return true
	default:
		return false
	}
}

func compareSortableValues(left, right any) int {
	left = decodeSortableValue(left)
	right = decodeSortableValue(right)
	if left == nil && right == nil {
		return 0
	}
	if left == nil {
		return 1
	}
	if right == nil {
		return -1
	}
	if leftNumber, ok := numericValue(left); ok {
		if rightNumber, rightOK := numericValue(right); rightOK {
			switch {
			case leftNumber < rightNumber:
				return -1
			case leftNumber > rightNumber:
				return 1
			default:
				return 0
			}
		}
	}
	return strings.Compare(propertyString(left), propertyString(right))
}

func numericValue(value any) (float64, bool) {
	switch number := value.(type) {
	case int:
		return float64(number), true
	case int8:
		return float64(number), true
	case int16:
		return float64(number), true
	case int32:
		return float64(number), true
	case int64:
		return float64(number), true
	case uint:
		return float64(number), true
	case uint8:
		return float64(number), true
	case uint16:
		return float64(number), true
	case uint32:
		return float64(number), true
	case uint64:
		return float64(number), true
	case float32:
		return float64(number), true
	case float64:
		return number, true
	case json.Number:
		parsed, err := number.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func decodeSortableValue(value any) any {
	if encoded, ok := value.(config.JSONValue); ok {
		decoded, err := encoded.Any()
		if err == nil {
			return decoded
		}
	}
	return value
}
