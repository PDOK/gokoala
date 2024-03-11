package config

import (
	"encoding/json"

	"golang.org/x/text/language"
)

// Language represents a BCP 47 language tag.
// +kubebuilder:validation:Type=string
type Language struct {
	language.Tag
}

// MarshalJSON turn language tag into JSON
func (l *Language) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Tag.String())
}

// UnmarshalJSON turn JSON into Language
func (l *Language) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*l = Language{language.Make(s)}
	return nil
}
