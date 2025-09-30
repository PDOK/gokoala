package config

import (
	"log"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	lowercaseIDRegexp = regexp.MustCompile("^[a-z0-9\"]([a-z0-9_-]*[a-z0-9\"]+|)$")
)

const (
	lowercaseID = "lowercase_id"
)

// LowercaseID is the validation function for validating if the current field
// is not empty and contains only lowercase chars, numbers, hyphens or underscores.
// It's similar to RFC 1035 DNS label but not the same.
func LowercaseID(fl validator.FieldLevel) bool {
	valAsString := fl.Field().String()
	valid := lowercaseIDRegexp.MatchString(valAsString)
	if !valid {
		log.Printf("Invalid ID %s", valAsString)
	}

	return valid
}
