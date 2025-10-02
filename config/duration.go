package config

import (
	"encoding/json"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration Custom time.Duration compatible with YAML and JSON (un)marshalling and kubebuilder.
// (Already supported in yaml/v3 but not encoding/json.)
//
// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Format=duration
type Duration struct {
	time.Duration
}

// MarshalJSON turn duration tag into JSON
// Value instead of pointer receiver because only that way it can be used for both.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, &d.Duration)
}

// MarshalYAML turn duration tag into YAML
// Value instead of pointer receiver because only that way it can be used for both.
func (d Duration) MarshalYAML() (any, error) {
	return d.Duration, nil
}

func (d *Duration) UnmarshalYAML(unmarshal func(any) error) error {
	return unmarshal(&d.Duration)
}

// DeepCopyInto copy the receiver, write into out. in must be non-nil.
func (d *Duration) DeepCopyInto(out *Duration) {
	if out != nil {
		*out = *d
	}
}

// DeepCopy copy the receiver, create a new Duration.
func (d *Duration) DeepCopy() *Duration {
	if d == nil {
		return nil
	}
	out := &Duration{}
	d.DeepCopyInto(out)

	return out
}
