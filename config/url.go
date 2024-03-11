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
type URL struct {
	*url.URL
}

// UnmarshalYAML parses a string to URL and also removes trailing slash if present,
// so we can easily append a longer path without having to worry about double slashes.
func (o *URL) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	parsedURL, err := url.ParseRequestURI(strings.TrimSuffix(s, "/"))
	o.URL = parsedURL
	return err
}

// MarshalJSON turns URL into JSON.
func (o *URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.URL.String())
}

// UnmarshalJSON parses a string to URL and also removes trailing slash if present,
// so we can easily append a longer path without having to worry about double slashes.
func (o *URL) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if parsedURL, err := url.ParseRequestURI(strings.TrimSuffix(s, "/")); err != nil {
		return err
	} else if parsedURL != nil {
		*o = URL{parsedURL}
	}
	return nil
}
