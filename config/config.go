//go:generate ../hack/generate-deepcopy.sh
package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

const (
	CookieMaxAge = 60 * 60 * 24
)

// NewConfig read YAML config file, required to start Gomagpie
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
	err = validateLocalPaths(config)
	if err != nil {
		return nil, fmt.Errorf("validation error in config file, error: %w", err)
	}
	return config, nil
}

// UnmarshalYAML hooks into unmarshalling to set defaults and validate config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

type Config struct {
	// Version of the API. When releasing a new version which contains backwards-incompatible changes, a new major version must be released.
	Version string `yaml:"version" json:"version" validate:"required,semver" default:"1.0.0"`

	// Human friendly title of the API.
	Title string `yaml:"title" json:"title"  validate:"required" default:"Location API"`

	// Shorted title / abbreviation describing the API.
	ServiceIdentifier string `yaml:"serviceIdentifier"  json:"serviceIdentifier" validate:"required" default:"Location API"`

	// Human friendly description of the API and dataset.
	Abstract string `yaml:"abstract" json:"abstract" validate:"required" default:"Location search & geocoding API"`

	// Licensing term that apply to this API and dataset
	License License `yaml:"license" json:"license" validate:"required"`

	// The base URL - that's the part until the OGC API landing page - under which this API is served
	BaseURL URL `yaml:"baseUrl" json:"baseUrl" validate:"required"`

	// The languages/translations to offer, valid options are Dutch (nl) and English (en). Dutch is the default.
	AvailableLanguages []Language `yaml:"availableLanguages,omitempty" json:"availableLanguages,omitempty"`

	// Reference to a PNG image to use a thumbnail on the landing page.
	// The full path is constructed by appending Resources + Thumbnail.
	// +optional
	Thumbnail *string `yaml:"thumbnail,omitempty" json:"thumbnail,omitempty"`

	// Moment in time when the dataset was last updated
	LastUpdated *string `yaml:"lastUpdated,omitempty" json:"lastUpdated,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z"`

	// Who updated the dataset
	LastUpdatedBy string `yaml:"lastUpdatedBy,omitempty" json:"lastUpdatedBy,omitempty"`

	// Available support channels
	Support *Support `yaml:"support,omitempty" json:"support,omitempty"`

	// Location where resources (e.g. thumbnails) specific to the given dataset are hosted
	Resources *Resources `yaml:"resources,omitempty" json:"resources,omitempty"`

	// Database to run the queries against
	Database Database `yaml:"database" json:"database" validate:"required"`

	// Order in which collections should be returned.
	// When not specified collections are returned in alphabetic order.
	CollectionOrder []string `yaml:"collectionOrder,omitempty" json:"collectionOrder,omitempty"`

	// Collections offered through this API
	Collections GeoSpatialCollections `yaml:"collections,omitempty" json:"collections,omitempty" validate:"required,dive"`
}

type Database struct {
	// ConnectionString to connect with backing database
	ConnectionString string `yaml:"connectionString" json:"connectionString" validate:"required"`
}

type License struct {
	// Name of the license, e.g. MIT, CC0, etc
	Name string `yaml:"name" json:"name" validate:"required" default:"CC0"`

	// URL to license text on the web
	URL URL `yaml:"url" json:"url" validate:"required" default:"https://creativecommons.org/publicdomain/zero/1.0/deed"`
}

type Support struct {
	// Name of the support organization
	Name string `yaml:"name" json:"name" validate:"required"`

	// URL to external support webpage
	URL URL `yaml:"url" json:"url" validate:"required"`

	// Email for support questions
	Email string `yaml:"email,omitempty" json:"email,omitempty" validate:"omitempty,email"`
}

type Resources struct {
	// Location where resources (e.g. thumbnails) specific to the given dataset are hosted. This is optional if Directory is set
	URL *URL `yaml:"url,omitempty" json:"url,omitempty" validate:"required_without=Directory,omitempty"`

	// Location where resources (e.g. thumbnails) specific to the given dataset are hosted. This is optional if URL is set
	Directory *string `yaml:"directory,omitempty" json:"directory,omitempty" validate:"required_without=URL,omitempty,dirpath|filepath"`
}

func (c *Config) CookieMaxAge() int {
	return CookieMaxAge
}

func setDefaults(config *Config) error {
	// process 'default' tags
	if err := defaults.Set(config); err != nil {
		return fmt.Errorf("failed to set default configuration: %w", err)
	}

	// custom default logic
	if len(config.AvailableLanguages) == 0 {
		config.AvailableLanguages = append(config.AvailableLanguages, Language{language.Dutch}) // default to Dutch only
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

// validateLocalPaths validates the existence of local paths.
// Not suitable for general validation while unmarshalling.
// Because that could happen on another machine.
func validateLocalPaths(config *Config) error {
	// Could use a deep dive and reflection.
	// But the settings with a path are not recursive and relatively limited in numbers.
	// GeoPackageCloudCache.Path is not verified. It will be created anyway in cloud_sqlite_vfs.createCacheDir during startup time.
	if config.Resources != nil && config.Resources.Directory != nil && *config.Resources.Directory != "" &&
		!isExistingLocalDir(*config.Resources.Directory) {
		return errors.New("Config.Resources.Directory should be an existing directory: " + *config.Resources.Directory)
	}
	return nil
}

func isExistingLocalDir(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && fileInfo.IsDir()
}
