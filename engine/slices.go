package engine

// Contains reports whether v is present in s.
//
// Source: https://github.com/golang/exp/blob/master/slices/slices.go
func Contains[E comparable](s []E, v E) bool {
	return Index(s, v) >= 0
}

// Index returns the index of the first occurrence of v in s,
// or -1 if not present.
//
// Source: https://github.com/golang/exp/blob/master/slices/slices.go
func Index[E comparable](s []E, v E) int {
	for i, vs := range s {
		if v == vs {
			return i
		}
	}
	return -1
}
