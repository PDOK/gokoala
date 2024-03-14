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

func (d *Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, &d.Duration)
}

func (d *Duration) MarshalYAML() (interface{}, error) {
	return d.Duration, nil
}

func (d *Duration) UnmarshalYAML(unmarshal func(any) error) error {
	return unmarshal(&d.Duration)
}

func (d *Duration) DeepCopyInto(out *Duration) {
	if out != nil {
		*out = *d
	}
}

func (d *Duration) DeepCopy() *Duration {
	if d == nil {
		return nil
	}
	out := &Duration{}
	d.DeepCopyInto(out)
	return out
}
