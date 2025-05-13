package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestTemplateKeyWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  []TemplateKeyOption
		expected TemplateKey
	}{
		{
			name:    "default values",
			options: nil,
			expected: TemplateKey{
				Name:               "landing-page.go.json",
				Directory:          "internal/ogc/common/core/templates",
				Format:             "json",
				Language:           language.Dutch,
				InstanceName:       "",
				MediaTypeOverwrite: "",
			},
		},
		{
			name:    "with language",
			options: []TemplateKeyOption{WithLanguage(language.English)},
			expected: TemplateKey{
				Name:               "landing-page.go.json",
				Directory:          "internal/ogc/common/core/templates",
				Format:             "json",
				Language:           language.English,
				InstanceName:       "",
				MediaTypeOverwrite: "",
			},
		},
		{
			name:    "with instance name",
			options: []TemplateKeyOption{WithInstanceName("test-instance")},
			expected: TemplateKey{
				Name:               "landing-page.go.json",
				Directory:          "internal/ogc/common/core/templates",
				Format:             "json",
				Language:           language.Dutch,
				InstanceName:       "test-instance",
				MediaTypeOverwrite: "",
			},
		},
		{
			name:    "with media type",
			options: []TemplateKeyOption{WithMediaTypeOverwrite("application/docx")},
			expected: TemplateKey{
				Name:               "landing-page.go.json",
				Directory:          "internal/ogc/common/core/templates",
				Format:             "json",
				Language:           language.Dutch,
				InstanceName:       "",
				MediaTypeOverwrite: "application/docx",
			},
		},
		{
			name: "with multiple options",
			options: []TemplateKeyOption{
				WithLanguage(language.English),
				WithInstanceName("test-instance"),
				WithMediaTypeOverwrite("application/docx"),
			},
			expected: TemplateKey{
				Name:               "landing-page.go.json",
				Directory:          "internal/ogc/common/core/templates",
				Format:             "json",
				Language:           language.English,
				InstanceName:       "test-instance",
				MediaTypeOverwrite: "application/docx",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := NewTemplateKey("internal/ogc/common/core/templates/landing-page.go.json", tt.options...)
			assert.Equal(t, tt.expected.Name, key.Name)
			assert.Equal(t, tt.expected.Directory, key.Directory)
			assert.Equal(t, tt.expected.Format, key.Format)
			assert.Equal(t, tt.expected.Language, key.Language)
			assert.Equal(t, tt.expected.InstanceName, key.InstanceName)
			assert.Equal(t, tt.expected.MediaTypeOverwrite, key.MediaTypeOverwrite)
		})
	}
}
