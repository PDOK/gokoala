package config

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	gokoalaIDRegexp = regexp.MustCompile("^[a-z0-9]([a-z0-9_-]*[a-z0-9]+|)$")
)

const (
	gokoalaID = "gokoala_id"
)

func RegisterAllValidators(v *validator.Validate) error {
	return v.RegisterValidation(gokoalaID, GokoalaID)
}

// GokoalaID is the validation function for validating if the current field
// is not empty and contains only lowercase chars, numbers, hyphens or underscores.
// It's similar to RFC 1035 DNS label but not the same.
func GokoalaID(fl validator.FieldLevel) bool {
	valAsString := fl.Field().String()
	return gokoalaIDRegexp.MatchString(valAsString)
}
