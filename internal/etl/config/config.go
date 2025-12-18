package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// NewConfig read YAML config file
func NewConfig(configFile string) (*Config, error) {
	yamlData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %w", err)
	}

	// expand environment variables
	yamlData = []byte(os.ExpandEnv(string(yamlData)))

	var config *Config
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file, error: %w", err)
	}
	return config, nil
}

type Config struct {
	// Collections used in this dataset
	Collections []Collection `yaml:"collections" json:"collections" validate:"required"`
}

func (c *Config) CollectionByID(id string) *Collection {
	for _, coll := range c.Collections {
		if coll.ID == id {
			return &coll
		}
	}
	return nil
}

// UnmarshalYAML hooks into unmarshalling to set defaults and validate config
func (c *Config) UnmarshalYAML(unmarshal func(any) error) error {
	type cfg Config
	if err := unmarshal((*cfg)(c)); err != nil {
		return err
	}

	// init config
	if err := setDefaults(c); err != nil {
		return err
	}
	if err := validate(c); err != nil {
		return err
	}
	return nil
}

func setDefaults(config *Config) error {
	// process 'default' tags
	if err := defaults.Set(config); err != nil {
		return fmt.Errorf("failed to set default configuration: %w", err)
	}
	return nil
}

func validate(config *Config) error {
	// process 'validate' tags
	v := validator.New()
	err := v.Struct(config)
	if err != nil {
		var ive *validator.InvalidValidationError
		if ok := errors.Is(err, ive); ok {
			return fmt.Errorf("failed to validate config: %w", err)
		}
		var errMessages []string
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			for _, valErr := range valErrs {
				errMessages = append(errMessages, valErr.Error()+"\n")
			}
		}
		return fmt.Errorf("invalid config provided:\n%v", errMessages)
	}
	return nil
}

type Collection struct {
	// Collection ID
	ID string `yaml:"id" json:"id" validate:"required"`

	// One or more feature tables backing this collection.
	Tables []FeatureTable `yaml:"tables" json:"tables" validate:"required"`

	// Fields that make up the display name and/or suggestions. These fields can be used as variables in the DisplayNameTemplate and SuggestTemplates.
	Fields []string `yaml:"fields,omitempty" json:"fields,omitempty" validate:"required"`

	// Template that indicates how a search record is displayed. Uses Go text/template syntax to reference fields.
	DisplayNameTemplate string `yaml:"displayNameTemplate,omitempty" json:"displayNameTemplate,omitempty" validate:"required"`

	// Version of this collection exposed through the API e.g., q=foo&thiscollection[version]=1&othercollection[version]=2
	Version int `yaml:"version,omitempty" json:"version,omitempty" default:"1"`

	// One or more templates that make up the autosuggestions. Uses Go text/template syntax to reference fields.
	SuggestTemplates []string `yaml:"suggestTemplates" json:"suggestTemplates" validate:"required,min=1"`

	// SQLite WHERE clause to filter features when importing/ETL-ing
	// (Without the WHERE keyword, only the clause)
	// +optional
	Filter string `yaml:"filter,omitempty" json:"filter,omitempty"`

	// Optional configuration for generation of external_fid
	// +optional
	ExternalFid *ExternalFid `yaml:"externalFid,omitempty" json:"externalFid,omitempty"`
}

type FeatureTable struct {
	// Name of the feature table
	Table string `yaml:"table" json:"table" validate:"required"`

	// Name of the feature ID column
	// +optional
	FID string `yaml:"fid,omitempty" json:"fid,omitempty" default:"fid" validate:"required"`

	// Name of the geometry column
	// +optional
	Geom string `yaml:"geom,omitempty" json:"geom,omitempty" default:"geom" validate:"required"`
}

type ExternalFid struct {
	// Namespace (UUID5) used to generate external_fid, defaults to uuid.NameSpaceURL
	// +kubebuilder:default="6ba7b811-9dad-11d1-80b4-00c04fd430c8"
	UUIDNamespace uuid.UUID `yaml:"uuidNamespace,omitempty" json:"uuidNamespace,omitempty" default:"6ba7b811-9dad-11d1-80b4-00c04fd430c8" validate:"required"`

	// Fields used to generate external_fid in the target OGC Features Collection(s).
	// Field names should match those in the source datasource.
	Fields []string `yaml:"fields" json:"fields" validate:"required"`
}
