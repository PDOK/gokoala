package config

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type TestEmbeddedDuration struct {
	D Duration `json:"D" yaml:"D"`
}

func TestDuration_DeepCopy(t *testing.T) {
	tests := []struct {
		duration *Duration
	}{
		{
			duration: &Duration{15},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := tt.duration.DeepCopy()
			assert.Equal(t, tt.duration, got, "DeepCopy")
			assert.NotSamef(t, tt.duration, got, "DeepCopy")
		})
	}
}

func TestDuration_DeepCopyInto(t *testing.T) {
	tests := []struct {
		duration *Duration
	}{
		{
			duration: &Duration{15},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := &Duration{}
			tt.duration.DeepCopyInto(got)
			assert.Equal(t, tt.duration, got, "DeepCopyInto")
			assert.NotSamef(t, tt.duration, got, "DeepCopyInto")
		})
	}
}

func TestDuration_Marshalling_JSON(t *testing.T) {
	tests := []struct {
		duration *Duration
		want     string
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			duration: &Duration{14},
			want:     `"14ns"`,
			wantErr:  assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			marshalled, err := json.Marshal(tt.duration)
			if !tt.wantErr(t, err, errors.New("json.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "json.Marshal")

			unmarshalled := &Duration{}
			err = json.Unmarshal(marshalled, unmarshalled)
			if !tt.wantErr(t, err, errors.New("json.Unmarshal")) {
				return
			}
			assert.Equalf(t, tt.duration, unmarshalled, "json.Unmarshal")

			// non-pointer
			unmarshalledEmbedded := &TestEmbeddedDuration{}
			err = yaml.Unmarshal([]byte(`{"D": `+tt.want+`}`), unmarshalledEmbedded)
			if !tt.wantErr(t, err, errors.New("yaml.Unmarshal")) {
				return
			}
			assert.EqualValuesf(t, &TestEmbeddedDuration{D: *tt.duration}, unmarshalledEmbedded, "yaml.Unmarshal")
		})
	}
}

func TestDuration_Marshalling_YAML(t *testing.T) {
	tests := []struct {
		duration *Duration
		want     string
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			duration: &Duration{14},
			want:     `14ns` + "\n",
			wantErr:  assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			marshalled, err := yaml.Marshal(tt.duration)
			if !tt.wantErr(t, err, errors.New("yaml.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "yaml.Marshal")

			unmarshalled := &Duration{}
			err = yaml.Unmarshal(marshalled, unmarshalled)
			if !tt.wantErr(t, err, errors.New("yaml.Unmarshal")) {
				return
			}
			assert.Equalf(t, tt.duration, unmarshalled, "yaml.Unmarshal")
		})
	}
}
