package config

import (
	"encoding/json"

	"github.com/elnormous/contenttype"
)

// MediaType represents a IANA media type as described in RFC 6838. Media types were formerly known as MIME types.
// +kubebuilder:validation:Type=string
type MediaType struct {
	contenttype.MediaType
}

// MarshalJSON turn MediaType into JSON
// Value instead of pointer receiver because only that way it can be used for both.
func (m MediaType) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

// UnmarshalJSON turn JSON into MediaType.
func (m *MediaType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	mt, err := contenttype.ParseMediaType(s)
	if err != nil {
		return err
	}
	m.MediaType = mt

	return nil
}

// MarshalYAML turns MediaType into YAML.
// Value instead of pointer receiver because only that way it can be used for both.
func (m MediaType) MarshalYAML() (any, error) {
	return m.String(), nil
}

// UnmarshalYAML parses a string to MediaType.
func (m *MediaType) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	mt, err := contenttype.ParseMediaType(s)
	if err != nil {
		return err
	}
	m.MediaType = mt

	return nil
}

// DeepCopyInto copy the receiver, write into out. in must be non-nil.
func (m *MediaType) DeepCopyInto(out *MediaType) {
	*out = *m
}

// DeepCopy copy the receiver, create a new MediaType.
func (m *MediaType) DeepCopy() *MediaType {
	if m == nil {
		return nil
	}
	out := &MediaType{}
	m.DeepCopyInto(out)

	return out
}
