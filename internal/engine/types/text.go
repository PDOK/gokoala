package types

import (
	"unicode"
	"unicode/utf8"
)

// IsValidString returns true if the given string is valid UTF-8
func IsValidString(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}

	for _, r := range s {
		// check for non-printable control characters
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
