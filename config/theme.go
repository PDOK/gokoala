package config

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Theme struct {
	Logo     *ThemeLogo     `yaml:"logo" json:"logo" validate:"required"`
	Color    *ThemeColors   `yaml:"color" json:"color" validate:"required"`
	Includes *ThemeIncludes `yaml:"includes" json:"includes"`
	Path     string
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
	yamlData, err := os.ReadFile(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file %w", err)
	}

	err = yaml.Unmarshal(yamlData, &theme)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal theme file, error: %w", err)
	}

	// check 'validate' tags
	v := validator.New()
	err = v.Struct(theme)
	if err != nil {
		return nil, formatValidationErr(err)
	}
	// if valid, set theme location
	theme.Path = filepath.Dir(cfg)
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
