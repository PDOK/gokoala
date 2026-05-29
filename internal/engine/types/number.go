package types

import "fmt"

// IsNumeric return true when v is a number, false otherwise.
func IsNumeric(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		complex64, complex128:
		return true
	default:
		return false
	}
}

// IsFloat return true when float has decimals, false otherwise.
func IsFloat(f float64) bool {
	return f != float64(int64(f))
}

// ToInt64 converts v to int64.
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
