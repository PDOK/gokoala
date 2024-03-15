package config

import (
	"encoding/json"
	"net/url"
	"strings"

	"gopkg.in/yaml.v3"
)

// URL Custom net.URL compatible with YAML and JSON (un)marshalling and kubebuilder.
// In addition, it also removes trailing slash if present, so we can easily
// append a longer path without having to worry about double slashes.
//
// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Format=uri
// +kubebuilder:validation:Pattern=`^https?://.+`
type URL struct {
	*url.URL
}

// UnmarshalYAML parses a string to URL and also removes trailing slash if present,
// so we can easily append a longer path without having to worry about double slashes.
func (u *URL) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	if parsedURL, err := parseURL(s); err != nil {
		return err
	} else if parsedURL != nil {
		u.URL = parsedURL
	}
	return nil
}

// MarshalJSON turns URL into JSON.
func (u *URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.URL.String())
}

// UnmarshalJSON parses a string to URL and also removes trailing slash if present,
// so we can easily append a longer path without having to worry about double slashes.
func (u *URL) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, u)
}

// MarshalYAML turns URL into YAML.
func (u *URL) MarshalYAML() (interface{}, error) {
	return u.URL.String(), nil
}

// DeepCopyInto copies the receiver, writes into out.
func (u *URL) DeepCopyInto(out *URL) {
	if out != nil {
		*out = *u
	}
}

// DeepCopy copies the receiver, creates a new URL.
func (u *URL) DeepCopy() *URL {
	if u == nil {
		return nil
	}
	out := &URL{}
	u.DeepCopyInto(out)
	return out
}

func parseURL(s string) (*url.URL, error) {
	return url.ParseRequestURI(strings.TrimSuffix(s, "/"))
}
