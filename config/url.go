package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	validURLRegexp = regexp.MustCompile(`^(https?://.+|\$\{.+\}.*)$`) // https://regex101.com/r/IvhP6H/1
)

// URL Custom net.URL compatible with YAML and JSON (un)marshalling and kubebuilder.
// In addition, it also removes trailing slash if present, so we can easily
// append a longer path without having to worry about double slashes.
//
// Allow only http/https URLs or environment variables like ${FOOBAR}
// +kubebuilder:validation:Pattern=`^(https?://.+)|(\$\{.+\}.*)`
// +kubebuilder:validation:Type=string
type URL struct {
	// This is a pointer so the wrapper can directly be used in templates, e.g.: {{ .Config.BaseURL }}
	// Otherwise you would need .String() or template.URL(). (Might be a bug.)
	*url.URL
}

// NewURL parses a string to URL and also resolves any environment variables present in the given string.
func NewURL(s string) (*URL, error) {
	resolvedURL := os.ExpandEnv(s)
	parsedURL, err := parseURL(resolvedURL)
	if err != nil {
		return nil, err
	}
	return &URL{URL: parsedURL}, nil
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
// Value instead of pointer receiver because only that way it can be used for both.
func (u URL) MarshalJSON() ([]byte, error) {
	if u.URL == nil {
		return json.Marshal("")
	}

	return json.Marshal(u.String())
}

// UnmarshalJSON parses a string to URL and also removes trailing slash if present,
// so we can easily append a longer path without having to worry about double slashes.
func (u *URL) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, u)
}

// MarshalYAML turns URL into YAML.
// Value instead of pointer receiver because only that way it can be used for both.
func (u URL) MarshalYAML() (any, error) {
	if u.URL == nil {
		return "", nil
	}

	return u.String(), nil
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
	if !validURLRegexp.MatchString(s) {
		return nil, fmt.Errorf("invalid URL: %s", s)
	}

	return url.Parse(strings.TrimSuffix(s, "/"))
}
