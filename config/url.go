package config

import (
	"encoding/json"
	"net/url"
	"strings"
)

// URL Custom net.URL compatible with YAML and JSON (un)marshalling and kubebuilder.
// In addition, it also removes trailing slash if present, so we can easily
// append a longer path without having to worry about double slashes.
//
// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Format=uri
// +kubebuilder:validation:Pattern=`^https?://`
// +kubebuilder:object:generate=true
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
	if parsedURL, err := parse(s); err != nil {
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
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if parsedURL, err := parse(s); err != nil {
		return err
	} else if parsedURL != nil {
		*u = URL{parsedURL}
	}
	return nil
}

// DeepCopyInto copy the receiver, write into out. in must be non-nil.
func (u *URL) DeepCopyInto(out *URL) {
	*out = *u
}

// DeepCopy copy the receiver, create a new URL.
func (u *URL) DeepCopy() *URL {
	if u == nil {
		return nil
	}
	out := &URL{}
	u.DeepCopyInto(out)
	return out
}

func parse(s string) (*url.URL, error) {
	return url.ParseRequestURI(strings.TrimSuffix(s, "/"))
}
