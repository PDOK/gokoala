package config

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

type TestEmbeddedLanguage struct {
	L Language `json:"L" yaml:"L"` //nolint:tagliatelle
}

func TestLanguage_DeepCopy(t *testing.T) {
	tests := []struct {
		lang *Language
	}{
		{
			lang: &Language{language.Afrikaans},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := tt.lang.DeepCopy()
			assert.Equal(t, tt.lang, got, "DeepCopy")
			assert.NotSamef(t, tt.lang, got, "DeepCopy")
		})
	}
}

func TestLanguage_DeepCopyInto(t *testing.T) {
	tests := []struct {
		lang *Language
	}{
		{
			lang: &Language{language.Afrikaans},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := &Language{}
			tt.lang.DeepCopyInto(got)
			assert.Equal(t, tt.lang, got, "DeepCopyInto")
			assert.NotSamef(t, tt.lang, got, "DeepCopyInto")
		})
	}
}

func TestLanguage_Marshalling_JSON(t *testing.T) {
	tests := []struct {
		lang    *Language
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			lang:    &Language{language.French},
			want:    `"fr"`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			marshalled, err := json.Marshal(tt.lang)
			if !tt.wantErr(t, err, errors.New("json.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "json.Marshal")

			unmarshalled := &Language{}
			err = json.Unmarshal(marshalled, unmarshalled)
			if !tt.wantErr(t, err, errors.New("json.Unmarshal")) {
				return
			}
			assert.Equalf(t, tt.lang, unmarshalled, "json.Unmarshal")

			// non-pointer
			unmarshalledEmbedded := &TestEmbeddedLanguage{}
			err = yaml.Unmarshal([]byte(`{"L": `+tt.want+`}`), unmarshalledEmbedded)
			if !tt.wantErr(t, err, errors.New("yaml.Unmarshal")) {
				return
			}
			assert.Equalf(t, &TestEmbeddedLanguage{L: *tt.lang}, unmarshalledEmbedded, "yaml.Unmarshal")
		})
	}
}

func TestLanguage_Marshalling_YAML(t *testing.T) {
	tests := []struct {
		lang    *Language
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			lang:    &Language{language.French},
			want:    "fr\n",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			marshalled, err := yaml.Marshal(tt.lang)
			if !tt.wantErr(t, err, errors.New("yaml.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "yaml.Marshal")

			unmarshalled := &Language{}
			err = yaml.Unmarshal(marshalled, unmarshalled)
			if !tt.wantErr(t, err, errors.New("yaml.Unmarshal")) {
				return
			}
			assert.Equalf(t, tt.lang, unmarshalled, "yaml.Unmarshal")
		})
	}
}
