package types

import "time"

// IsDate return true when time.Time doesn't contain a time component, false otherwise
func IsDate(t time.Time) bool {
	return t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0
}

// IsFloat return true when float has decimals, false otherwise
func IsFloat(f float64) bool {
	return f != float64(int64(f))
}
