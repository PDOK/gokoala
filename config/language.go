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
// Value instead of pointer receiver because only that way it can be used for both.
func (l Language) MarshalJSON() ([]byte, error) {
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

// DeepCopyInto copy the receiver, write into out. in must be non-nil.
func (l *Language) DeepCopyInto(out *Language) {
	*out = *l
}

// DeepCopy copy the receiver, create a new Language.
func (l *Language) DeepCopy() *Language {
	if l == nil {
		return nil
	}
	out := &Language{}
	l.DeepCopyInto(out)
	return out
}
