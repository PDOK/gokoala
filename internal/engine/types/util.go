// Package types package contains generic types
package types

import (
	"fmt"
	"time"
)

// IsDate return true when time.Time doesn't contain a time component, false otherwise.
func IsDate(t time.Time) bool {
	return t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0
}

// IsFloat return true when float has decimals, false otherwise.
func IsFloat(f float64) bool {
	return f != float64(int64(f))
}

func ToInt64(v any) (int64, error) {
	switch i := v.(type) {
	case int:
		return int64(i), nil
	case int32:
		return int64(i), nil
	case int64:
		return i, nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}
