package config

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"dario.cat/mergo"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	defaultThemeConfig = "themes/pdok/theme.yaml"
)

type Theme struct {
	Logo        *ThemeLogo     `yaml:"logo" json:"logo" validate:"required"`
	Color       *ThemeColors   `yaml:"color" json:"color" validate:"required"`
	Includes    *ThemeIncludes `yaml:"includes" json:"includes"`
	Path        string
	defaultPath string
}
type ThemeLogo struct {
	Header    string `yaml:"header" json:"header" validate:"required"`
	Footer    string `yaml:"footer" json:"footer" validate:"required"`
	Opengraph string `yaml:"opengraph" json:"opengraph" validate:"required"`
	Favicon   string `yaml:"favicon" json:"favicon" validate:"required"`
	Favicon16 string `yaml:"favicon16" json:"favicon16" validate:"required"`
	Favicon32 string `yaml:"favicon32" json:"favicon32" validate:"required"`
}

type ThemeColors struct {
	Primary   string `yaml:"primary" json:"primary" validate:"required,hexcolor"`
	Secondary string `yaml:"secondary" json:"secondary" validate:"required,hexcolor"`
	Link      string `yaml:"link" json:"link" validate:"required,hexcolor"`
}

type ThemeIncludes struct {
	HTMLFile   string `yaml:"html"`
	ParsedHTML template.HTML
}

func NewTheme(cfg string) (theme *Theme, err error) {
	theme, err = getThemeFromFile(defaultThemeConfig)
	if err != nil {
		return nil, err
	}

	var customTheme *Theme
	if cfg != "" {
		// If a custom theme is present, also fetch it
		customTheme, err = getThemeFromFile(cfg)
		if err != nil {
			return nil, err
		}

		// Overwrite the basetheme
		err = mergo.Merge(theme, customTheme, mergo.WithOverride)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
			return nil, err
		}
	}

	// check 'validate' tags
	v := validator.New()
	err = v.Struct(theme)
	if err != nil {
		return nil, formatValidationErr(err)
	}
	// if valid, set theme location
	theme.ParseHTML()
	return theme, nil
}

func (t *Theme) ParseHTML() {
	if t.Includes == nil {
		t.Includes = &ThemeIncludes{}
	}
	path := filepath.Join(t.Path, t.Includes.HTMLFile)
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("failed to read html file %v", err)
		t.Includes.ParsedHTML = ""
		return
	}

	// #nosec G203 - trusted html so no threat
	t.Includes.ParsedHTML = template.HTML(content)
}

func getThemeFromFile(path string) (theme *Theme, err error) {
	yamlData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file %w", err)
	}

	if err = yaml.Unmarshal(yamlData, &theme); err != nil {
		return nil, fmt.Errorf("failed to unmarshal theme file, error: %w", err)
	}

	theme.Path = filepath.Dir(path)

	return
}
