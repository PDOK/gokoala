package config

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/elnormous/contenttype"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type TestEmbeddedMediaType struct {
	M MediaType `json:"M" yaml:"M"`
}

func TestMediaType_DeepCopy(t *testing.T) {
	tests := []struct {
		mediaType *MediaType
	}{
		{
			mediaType: &MediaType{MediaType: contenttype.MediaType{Type: "application", Subtype: "json", Parameters: make(map[string]string)}},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := tt.mediaType.DeepCopy()
			assert.Equal(t, tt.mediaType, got, "DeepCopy")
			assert.NotSamef(t, tt.mediaType, got, "DeepCopy")
		})
	}
}

func TestMediaType_DeepCopyInto(t *testing.T) {
	tests := []struct {
		mediaType *MediaType
	}{
		{
			mediaType: &MediaType{MediaType: contenttype.MediaType{Type: "application", Subtype: "json", Parameters: make(map[string]string)}},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := &MediaType{}
			tt.mediaType.DeepCopyInto(got)
			assert.Equal(t, tt.mediaType, got, "DeepCopyInto")
			assert.NotSamef(t, tt.mediaType, got, "DeepCopyInto")
		})
	}
}

func TestMediaType_Marshalling_JSON(t *testing.T) {
	tests := []struct {
		mediaType *MediaType
		want      string
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			mediaType: &MediaType{MediaType: contenttype.MediaType{Type: "application", Subtype: "json", Parameters: make(map[string]string)}},
			want:      `"application/json"`,
			wantErr:   assert.NoError,
		},
		{
			mediaType: &MediaType{},
			want:      `""`,
			wantErr:   assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			marshalled, err := json.Marshal(tt.mediaType)
			if !tt.wantErr(t, err, errors.New("json.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "json.Marshal")

			// non-pointer
			marshalled, err = json.Marshal(*tt.mediaType)
			if !tt.wantErr(t, err, errors.New("json.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "json.Marshal")
		})
	}
}

func TestMediaType_Unmarshalling_JSON(t *testing.T) {
	tests := []struct {
		mediaType string
		want      *MediaType
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			mediaType: `"application/json"`,
			want:      &MediaType{MediaType: contenttype.MediaType{Type: "application", Subtype: "json", Parameters: make(map[string]string)}},
			wantErr:   assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			unmarshalled := &MediaType{}
			err := json.Unmarshal([]byte(tt.mediaType), unmarshalled)
			if !tt.wantErr(t, err, errors.New("json.Unmarshal")) {
				return
			}
			assert.Equalf(t, tt.want, unmarshalled, "json.Unmarshal")

			// non-pointer
			unmarshalledEmbedded := &TestEmbeddedMediaType{}
			err = json.Unmarshal([]byte(`{"M": `+tt.mediaType+`}`), unmarshalledEmbedded)
			if !tt.wantErr(t, err, errors.New("json.Unmarshal")) {
				return
			}
			assert.EqualValuesf(t, &TestEmbeddedMediaType{M: *tt.want}, unmarshalledEmbedded, "json.Unmarshal")
		})
	}
}

func TestMediaType_Marshalling_YAML(t *testing.T) {
	tests := []struct {
		mediaType *MediaType
		want      string
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			mediaType: &MediaType{MediaType: contenttype.MediaType{Type: "application", Subtype: "json", Parameters: make(map[string]string)}},
			want:      `application/json` + "\n",
			wantErr:   assert.NoError,
		},
		{
			mediaType: &MediaType{},
			want:      `""` + "\n",
			wantErr:   assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			marshalled, err := yaml.Marshal(tt.mediaType)
			if !tt.wantErr(t, err, errors.New("yaml.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "yaml.Marshal")

			// non-pointer
			marshalled, err = yaml.Marshal(*tt.mediaType)
			if !tt.wantErr(t, err, errors.New("yaml.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "yaml.Marshal")
		})
	}
}

func TestMediaType_Unmarshalling_YAML(t *testing.T) {
	tests := []struct {
		mediaType string
		want      *MediaType
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			mediaType: `application/json` + "\n",
			want:      &MediaType{MediaType: contenttype.MediaType{Type: "application", Subtype: "json", Parameters: make(map[string]string)}},
			wantErr:   assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			unmarshalled := &MediaType{}
			err := yaml.Unmarshal([]byte(tt.mediaType), unmarshalled)
			if !tt.wantErr(t, err, errors.New("yaml.Unmarshal")) {
				return
			}
			assert.Equalf(t, tt.want, unmarshalled, "yaml.Unmarshal")

			// non-pointer
			unmarshalledEmbedded := &TestEmbeddedMediaType{}
			err = yaml.Unmarshal([]byte(`{"M": `+tt.mediaType+`}`), unmarshalledEmbedded)
			if !tt.wantErr(t, err, errors.New("yaml.Unmarshal")) {
				return
			}
			assert.EqualValuesf(t, &TestEmbeddedMediaType{M: *tt.want}, unmarshalledEmbedded, "yaml.Unmarshal")
		})
	}
}
