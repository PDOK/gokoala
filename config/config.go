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
	DefaultSrs   = "EPSG:28992"
)

// NewConfig read YAML config file, required to start GoKoala
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

// +kubebuilder:object:generate=true
type Config struct {
	// Version of the API. When releasing a new version which contains backwards-incompatible changes, a new major version must be released.
	Version string `yaml:"version" json:"version" validate:"required,semver"`

	// Human friendly title of the API. Don't include "OGC API" in the title, this is added automatically.
	Title string `yaml:"title" json:"title"  validate:"required"`

	// Shorted title / abbreviation describing the API.
	ServiceIdentifier string `yaml:"serviceIdentifier"  json:"serviceIdentifier" validate:"required"`

	// Human friendly description of the API and dataset.
	Abstract string `yaml:"abstract" json:"abstract" validate:"required"`

	// Licensing term that apply to this API and dataset
	License License `yaml:"license" json:"license" validate:"required"`

	// The base URL - that's the part until the OGC API landing page - under which this API is served
	BaseURL URL `yaml:"baseUrl" json:"baseUrl" validate:"required"`

	// Optional reference to a catalog/portal/registry that lists all datasets, not just this one
	// +optional
	DatasetCatalogURL URL `yaml:"datasetCatalogUrl,omitempty" json:"datasetCatalogUrl,omitempty"`

	// The languages/translations to offer, valid options are Dutch (nl) and English (en). Dutch is the default.
	// +optional
	AvailableLanguages []Language `yaml:"availableLanguages,omitempty" json:"availableLanguages,omitempty"`

	// Define which OGC API building blocks this API supports
	OgcAPI OgcAPI `yaml:"ogcApi" json:"ogcApi" validate:"required"`

	// Order in which collections (containing features, tiles, 3d tiles, etc.) should be returned.
	// When not specified collections are returned in alphabetic order.
	// +optional
	OgcAPICollectionOrder []string `yaml:"collectionOrder,omitempty" json:"collectionOrder,omitempty"`

	// Reference to a PNG image to use a thumbnail on the landing page.
	// The full path is constructed by appending Resources + Thumbnail.
	// +optional
	Thumbnail *string `yaml:"thumbnail,omitempty" json:"thumbnail,omitempty"`

	// Keywords to make this API beter discoverable
	// +optional
	Keywords []string `yaml:"keywords,omitempty" json:"keywords,omitempty"`

	// Moment in time when the dataset was last updated
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format="date-time"
	LastUpdated *string `yaml:"lastUpdated,omitempty" json:"lastUpdated,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z"`

	// Who updated the dataset
	// +optional
	LastUpdatedBy string `yaml:"lastUpdatedBy,omitempty" json:"lastUpdatedBy,omitempty"`

	// Available support channels
	// +optional
	Support *Support `yaml:"support,omitempty" json:"support,omitempty"`

	// Metadata links
	// +optional
	MetadataLinks []MetadataLink `yaml:"metadataLinks,omitempty" json:"metadataLinks,omitempty"`

	// Key/value pairs to add extra information to the landing page
	// +optional
	DatasetDetails []DatasetDetail `yaml:"datasetDetails,omitempty" json:"datasetDetails,omitempty"`

	// Location where resources (e.g. thumbnails) specific to the given dataset are hosted
	// +optional
	Resources *Resources `yaml:"resources,omitempty" json:"resources,omitempty"`
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

func (c *Config) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, c)
}

func (c *Config) CookieMaxAge() int {
	return CookieMaxAge
}

// +kubebuilder:object:generate=true
type OgcAPI struct {
	// Enable when this API should offer OGC API 3D GeoVolumes. This includes OGC 3D Tiles.
	// +optional
	GeoVolumes *OgcAPI3dGeoVolumes `yaml:"3dgeovolumes,omitempty" json:"3dgeovolumes,omitempty"`

	// Enable when this API should offer OGC API Tiles. This also requires OGC API Styles.
	// +optional
	Tiles *OgcAPITiles `yaml:"tiles,omitempty" json:"tiles,omitempty" validate:"required_with=Styles"`

	// Enable when this API should offer OGC API Styles.
	// +optional
	Styles *OgcAPIStyles `yaml:"styles,omitempty" json:"styles,omitempty"`

	// Enable when this API should offer OGC API Features.
	// +optional
	Features *OgcAPIFeatures `yaml:"features,omitempty" json:"features,omitempty"`

	// Enable when this API should offer OGC API Processes.
	// +optional
	Processes *OgcAPIProcesses `yaml:"processes,omitempty" json:"processes,omitempty"`
}

// +kubebuilder:object:generate=true
type Support struct {
	// Name of the support organization
	Name string `yaml:"name" json:"name" validate:"required"`

	// URL to external support webpage
	// +kubebuilder:validation:Type=string
	URL URL `yaml:"url" json:"url" validate:"required"`

	// Email for support questions
	// +optional
	Email string `yaml:"email,omitempty" json:"email,omitempty" validate:"omitempty,email"`
}

// +kubebuilder:object:generate=true
type MetadataLink struct {
	// Name of the metadata collection/site/organization
	Name string `yaml:"name" json:"name" validate:"required"`

	// Which category of the API this metadata concerns. E.g. dataset (in general), tiles or features
	// +kubebuilder:default="dataset"
	Category string `yaml:"category" json:"category" validate:"required" default:"dataset"`

	// URL to external metadata detail page
	// +kubebuilder:validation:Type=string
	URL URL `yaml:"url" json:"url" validate:"required"`
}

// +kubebuilder:object:generate=true
type DatasetDetail struct {
	// Arbitrary name to add extra information to the landing page
	Name string `yaml:"name" json:"name"`

	// Arbitrary value associated with the given name
	Value string `yaml:"value" json:"value"`
}

// +kubebuilder:object:generate=true
type Resources struct {
	// Location where resources (e.g. thumbnails) specific to the given dataset are hosted. This is optional if Directory is set
	// +optional
	URL *URL `yaml:"url,omitempty" json:"url,omitempty" validate:"required_without=Directory,omitempty"`

	// Location where resources (e.g. thumbnails) specific to the given dataset are hosted. This is optional if URL is set
	// +optional
	Directory *string `yaml:"directory,omitempty" json:"directory,omitempty" validate:"required_without=URL,omitempty,dirpath|filepath"`
}

// +kubebuilder:object:generate=true
type License struct {
	// Name of the license, e.g. MIT, CC0, etc
	Name string `yaml:"name" json:"name" validate:"required"`

	// URL to license text on the web
	URL URL `yaml:"url" json:"url" validate:"required"`
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
	if config.OgcAPI.Tiles != nil {
		config.OgcAPI.Tiles.Defaults()
	}
	return nil
}

func validate(config *Config) error {
	// process 'validate' tags
	v := validator.New()
	err := v.RegisterValidation(lowercaseID, LowercaseID)
	if err != nil {
		return err
	}
	err = v.Struct(config)
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

	// custom validations
	var errs []error
	if config.OgcAPI.Features != nil {
		errs = append(errs, validateFeatureCollections(config.OgcAPI.Features.Collections))
	}
	if config.OgcAPI.Tiles != nil {
		errs = append(errs, validateTileProjections(config.OgcAPI.Tiles))
	}
	err = errors.Join(errs...)
	if err != nil {
		return err
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
	if config.OgcAPI.Styles != nil && !isExistingLocalDir(config.OgcAPI.Styles.StylesDir) {
		return errors.New("Config.OgcAPI.Styles.StylesDir should be an existing directory: " + config.OgcAPI.Styles.StylesDir)
	}
	return nil
}

func isExistingLocalDir(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && fileInfo.IsDir()
}
