package config

import (
	"encoding/json"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type TestEmbeddedURL struct {
	U URL `json:"u" yaml:"u"`
}

func TestNewURL(t *testing.T) {
	t.Setenv("TEST_BASE_URL", "https://example.com")

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid HTTP URL",
			input:    "http://localhost:8080",
			expected: "http://localhost:8080",
			wantErr:  false,
		},
		{
			name:     "Valid HTTPS URL with trailing slash (should be trimmed)",
			input:    "https://example.com/",
			expected: "https://example.com",
			wantErr:  false,
		},
		{
			name:     "Environment variable resolution",
			input:    "${TEST_BASE_URL}/api",
			expected: "https://example.com/api",
			wantErr:  false,
		},
		{
			name:    "Invalid URL scheme",
			input:   "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "Plain text string (invalid URL)",
			input:   "not-a-url",
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewURL(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.expected, got.String())
			}
		})
	}
}

func TestURL_DeepCopy(t *testing.T) {
	tests := []struct {
		url *URL
	}{
		{
			url: &URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset/"}},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := tt.url.DeepCopy()
			assert.Equal(t, tt.url, got, "DeepCopy")
			assert.NotSamef(t, tt.url, got, "DeepCopy")
		})
	}
}

func TestURL_DeepCopyInto(t *testing.T) {
	tests := []struct {
		url *URL
	}{
		{
			url: &URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset/"}},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := &URL{}
			tt.url.DeepCopyInto(got)
			assert.Equal(t, tt.url, got, "DeepCopyInto")
			assert.NotSamef(t, tt.url, got, "DeepCopyInto")
		})
	}
}

func TestURL_Marshalling_JSON(t *testing.T) {
	tests := []struct {
		url     *URL
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			url:     &URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset/"}},
			want:    `"https://tiles.foobar.example/somedataset/"`,
			wantErr: assert.NoError,
		},
		{
			url:     &URL{},
			want:    `""`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			marshalled, err := json.Marshal(tt.url)
			if !tt.wantErr(t, err, errors.New("json.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "json.Marshal")

			// non-pointer
			marshalled, err = json.Marshal(*tt.url)
			if !tt.wantErr(t, err, errors.New("json.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "json.Marshal")
		})
	}
}

func TestURL_Unmarshalling_JSON(t *testing.T) {
	tests := []struct {
		url     string
		want    *URL
		wantErr assert.ErrorAssertionFunc
	}{
		{
			url:     `"https://tiles.foobar.example/somedataset/"`,
			want:    &URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset"}}, // no trailing slash
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			unmarshalled := &URL{}
			err := json.Unmarshal([]byte(tt.url), unmarshalled)
			if !tt.wantErr(t, err, errors.New("json.Unmarshal")) {
				return
			}
			assert.Equalf(t, tt.want, unmarshalled, "json.Unmarshal")

			// non-pointer
			unmarshalledEmbedded := &TestEmbeddedURL{}
			err = json.Unmarshal([]byte(`{"U": `+tt.url+`}`), unmarshalledEmbedded)
			if !tt.wantErr(t, err, errors.New("json.Unmarshal")) {
				return
			}
			assert.Equalf(t, &TestEmbeddedURL{U: *tt.want}, unmarshalledEmbedded, "json.Unmarshal")
		})
	}
}

func TestURL_Marshalling_YAML(t *testing.T) {
	tests := []struct {
		url     *URL
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			url:     &URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset/"}},
			want:    `https://tiles.foobar.example/somedataset/` + "\n",
			wantErr: assert.NoError,
		},
		{
			url:     &URL{},
			want:    `""` + "\n",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			marshalled, err := yaml.Marshal(tt.url)
			if !tt.wantErr(t, err, errors.New("yaml.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "yaml.Marshal")

			// non-pointer
			marshalled, err = yaml.Marshal(*tt.url)
			if !tt.wantErr(t, err, errors.New("yaml.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "yaml.Marshal")
		})
	}
}

func TestURL_Unmarshalling_YAML(t *testing.T) {
	tests := []struct {
		url     string
		want    *URL
		wantErr assert.ErrorAssertionFunc
	}{
		{
			url:     `https://tiles.foobar.example/somedataset/` + "\n",
			want:    &URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset"}}, // no trailing slash
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			unmarshalled := &URL{}
			err := yaml.Unmarshal([]byte(tt.url), unmarshalled)
			if !tt.wantErr(t, err, errors.New("yaml.Unmarshal")) {
				return
			}
			assert.Equalf(t, tt.want, unmarshalled, "yaml.Unmarshal")

			// non-pointer
			unmarshalledEmbedded := &TestEmbeddedURL{}
			err = yaml.Unmarshal([]byte(`{"u": `+tt.url+`}`), unmarshalledEmbedded)
			if !tt.wantErr(t, err, errors.New("yaml.Unmarshal")) {
				return
			}
			assert.Equalf(t, &TestEmbeddedURL{U: *tt.want}, unmarshalledEmbedded, "yaml.Unmarshal")
		})
	}
}
