package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Theme struct {
	Logo  *ThemeLogo   `yaml:"logo,omitempty" json:"logo,omitempty" validate:"omitempty"`
	Color *ThemeColors `yaml:"color,omitempty" json:"color,omitempty" validate:"omitempty"`
	Path  string
}

type ThemeLogo struct {
	Header    string `yaml:"header,omitempty" json:"header,omitempty" validate:"omitempty"`
	Footer    string `yaml:"footer,omitempty" json:"footer,omitempty" validate:"omitempty"`
	Opengraph string `yaml:"opengraph,omitempty" json:"opengraph,omitempty" validate:"omitempty"`
	Favicon   string `yaml:"favicon,omitempty" json:"favicon,omitempty" validate:"omitempty"`
	Favicon16 string `yaml:"favicon16,omitempty" json:"favicon16,omitempty" validate:"omitempty"`
	Favicon32 string `yaml:"favicon32,omitempty" json:"favicon32,omitempty" validate:"omitempty"`
}

type ThemeColors struct {
	Primary   string `yaml:"primary,omitempty" json:"primary,omitempty" validate:"hexcolor,omitempty"`
	Secondary string `yaml:"secondary,omitempty" json:"secondary,omitempty" validate:"hexcolor,omitempty"`
	Link      string `yaml:"link,omitempty" json:"link,omitempty" validate:"hexcolor,omitempty"`
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
	return theme, nil
}
