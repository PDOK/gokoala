//go:generate ../hack/generate-deepcopy.sh
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/creasty/defaults"
	"github.com/docker/go-units"
	"github.com/go-playground/validator/v10"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

const (
	CookieMaxAge = 60 * 60 * 24
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

	// custom validations
	if config.OgcAPI.Features != nil {
		return validateFeatureCollections(config.OgcAPI.Features.Collections)
	}
	return nil
}

func validateFeatureCollections(collections GeoSpatialCollections) error {
	var errMessages []string
	for _, collection := range collections {
		if collection.Metadata != nil && collection.Metadata.TemporalProperties != nil &&
			(collection.Metadata.Extent == nil || collection.Metadata.Extent.Interval == nil) {
			errMessages = append(errMessages, fmt.Sprintf("validation failed for collection '%s'; "+
				"field 'Extent.Interval' is required with field 'TemporalProperties'\n", collection.ID))
		}
		if collection.Features != nil && collection.Features.Filters.Properties != nil {
			for _, pf := range collection.Features.Filters.Properties {
				if pf.AllowedValues != nil && *pf.DeriveAllowedValuesFromDatasource {
					errMessages = append(errMessages, fmt.Sprintf("validation failed for property filter '%s'; "+
						"field 'AllowedValues' and field 'DeriveAllowedValuesFromDatasource' are mutually exclusive\n", pf.Name))
				}
			}
		}
	}
	if len(errMessages) > 0 {
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
	if config.OgcAPI.Styles != nil && !isExistingLocalDir(config.OgcAPI.Styles.StylesDir) {
		return errors.New("Config.OgcAPI.Styles.StylesDir should be an existing directory: " + config.OgcAPI.Styles.StylesDir)
	}
	return nil
}

func isExistingLocalDir(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && fileInfo.IsDir()
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

	// Key/value pairs to add extra information to the landing page
	// +optional
	DatasetDetails []DatasetDetail `yaml:"datasetDetails,omitempty" json:"datasetDetails,omitempty"`

	// Location where resources (e.g. thumbnails) specific to the given dataset are hosted
	// +optional
	Resources *Resources `yaml:"resources,omitempty" json:"resources,omitempty"`
}

func (c *Config) CookieMaxAge() int {
	return CookieMaxAge
}

func (c *Config) HasCollections() bool {
	return c.AllCollections() != nil
}

func (c *Config) AllCollections() GeoSpatialCollections {
	var result GeoSpatialCollections
	if c.OgcAPI.GeoVolumes != nil {
		result = append(result, c.OgcAPI.GeoVolumes.Collections...)
	}
	if c.OgcAPI.Tiles != nil {
		result = append(result, c.OgcAPI.Tiles.Collections...)
	}
	if c.OgcAPI.Features != nil {
		result = append(result, c.OgcAPI.Features.Collections...)
	}
	return result
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

	// // Location where resources (e.g. thumbnails) specific to the given dataset are hosted. This is optional if URL is set
	// +optional
	Directory *string `yaml:"directory,omitempty" json:"directory,omitempty" validate:"required_without=URL,omitempty,dirpath|filepath"`
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
type GeoSpatialCollection struct {
	// Unique ID of the collection
	ID string `yaml:"id" validate:"required" json:"id"`

	// Metadata describing the collection contents
	// +optional
	Metadata *GeoSpatialCollectionMetadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// Links pertaining to this collection (e.g., downloads, documentation)
	// +optional
	Links *CollectionLinks `yaml:"links,omitempty" json:"links,omitempty"`

	// 3D GeoVolumes specific to this collection
	// +optional
	GeoVolumes *CollectionEntry3dGeoVolumes `yaml:",inline" json:",inline"`

	// Tiles specific to this collection
	// +optional
	Tiles *CollectionEntryTiles `yaml:",inline" json:",inline"`

	// Features specific to this collection
	// +optional
	Features *CollectionEntryFeatures `yaml:",inline" json:",inline"`
}

type GeoSpatialCollectionJSON struct {
	ID                           string                        `json:"id"`
	Metadata                     *GeoSpatialCollectionMetadata `json:"metadata,omitempty"`
	*CollectionEntry3dGeoVolumes `json:",inline"`
	*CollectionEntryTiles        `json:",inline"`
	*CollectionEntryFeatures     `json:",inline"`
}

// MarshalJSON custom because inlining only works on embedded structs.
// Value instead of pointer receiver because only that way it can be used for both.
func (c GeoSpatialCollection) MarshalJSON() ([]byte, error) {
	return json.Marshal(GeoSpatialCollectionJSON{
		ID:                          c.ID,
		Metadata:                    c.Metadata,
		CollectionEntry3dGeoVolumes: c.GeoVolumes,
		CollectionEntryTiles:        c.Tiles,
		CollectionEntryFeatures:     c.Features,
	})
}

// UnmarshalJSON parses a string to GeoSpatialCollection
func (c *GeoSpatialCollection) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, c)
}

// +kubebuilder:object:generate=true
type GeoSpatialCollectionMetadata struct {
	// Human friendly title of this collection. When no title is specified the collection ID is used.
	// +optional
	Title *string `yaml:"title,omitempty" json:"title,omitempty"`

	// Describes the content of this collection
	Description *string `yaml:"description" json:"description" validate:"required"`

	// Reference to a PNG image to use a thumbnail on the collections.
	// The full path is constructed by appending Resources + Thumbnail.
	// +optional
	Thumbnail *string `yaml:"thumbnail,omitempty" json:"thumbnail,omitempty"`

	// Keywords to make this collection beter discoverable
	// +optional
	Keywords []string `yaml:"keywords,omitempty" json:"keywords,omitempty"`

	// Moment in time when the collection was last updated
	//
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format="date-time"
	LastUpdated *string `yaml:"lastUpdated,omitempty" json:"lastUpdated,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z"`

	// Who updated this collection
	// +optional
	LastUpdatedBy string `yaml:"lastUpdatedBy,omitempty" json:"lastUpdatedBy,omitempty"`

	// Fields in the datasource to be used in temporal queries
	// +optional
	TemporalProperties *TemporalProperties `yaml:"temporalProperties,omitempty" json:"temporalProperties,omitempty" validate:"omitempty,required_with=Extent.Interval"`

	// Extent of the collection, both geospatial and/or temporal
	// +optional
	Extent *Extent `yaml:"extent,omitempty" json:"extent,omitempty"`

	// The CRS identifier which the features are originally stored, meaning no CRS transformations are applied when features are retrieved in this CRS.
	// WGS84 is the default storage CRS.
	//
	// +kubebuilder:default="http://www.opengis.net/def/crs/OGC/1.3/CRS84"
	// +kubebuilder:validation:Pattern=`^http:\/\/www\.opengis\.net\/def\/crs\/.*$`
	// +optional
	StorageCrs *string `yaml:"storageCrs,omitempty" json:"storageCrs,omitempty" default:"http://www.opengis.net/def/crs/OGC/1.3/CRS84" validate:"startswith=http://www.opengis.net/def/crs"`
}

// +kubebuilder:object:generate=true
type CollectionEntry3dGeoVolumes struct {
	// Optional basepath to 3D tiles on the tileserver. Defaults to the collection ID.
	// +optional
	TileServerPath *string `yaml:"tileServerPath,omitempty" json:"tileServerPath,omitempty"`

	// URI template for individual 3D tiles.
	// +optional
	URITemplate3dTiles *string `yaml:"uriTemplate3dTiles,omitempty" json:"uriTemplate3dTiles,omitempty" validate:"required_without_all=URITemplateDTM"`

	// Optional URI template for subtrees, only required when "implicit tiling" extension is used.
	// +optional
	URITemplateImplicitTilingSubtree *string `yaml:"uriTemplateImplicitTilingSubtree,omitempty" json:"uriTemplateImplicitTilingSubtree,omitempty"`

	// URI template for digital terrain model (DTM) in Quantized Mesh format, REQUIRED when you want to serve a DTM.
	// +optional
	URITemplateDTM *string `yaml:"uriTemplateDTM,omitempty" json:"uriTemplateDTM,omitempty" validate:"required_without_all=URITemplate3dTiles"`

	// Optional URL to 3D viewer to visualize the given collection of 3D Tiles.
	// +optional
	URL3DViewer *URL `yaml:"3dViewerUrl,omitempty" json:"3dViewerUrl,omitempty"`
}

func (gv *CollectionEntry3dGeoVolumes) Has3DTiles() bool {
	return gv.URITemplate3dTiles != nil
}

func (gv *CollectionEntry3dGeoVolumes) HasDTM() bool {
	return gv.URITemplateDTM != nil
}

// +kubebuilder:object:generate=true
type CollectionEntryTiles struct {
	// placeholder
}

// +kubebuilder:object:generate=true
type CollectionEntryFeatures struct {
	// Optional way to explicitly map a collection ID to the underlying table in the datasource.
	// +optional
	TableName *string `yaml:"tableName,omitempty" json:"tableName,omitempty"`

	// Optional collection specific datasources. Mutually exclusive with top-level defined datasources.
	// +optional
	Datasources *Datasources `yaml:"datasources,omitempty" json:"datasources,omitempty"`

	// Filters available for this collection
	// +optional
	Filters FeatureFilters `yaml:"filters,omitempty" json:"filters,omitempty"`

	// Downloads available for this collection through map sheets. Note that 'map sheets' refer to a map
	// divided in rectangle areas that can be downloaded individually.
	// +optional
	MapSheetDownloads *MapSheetDownloads `yaml:"mapSheetDownloads,omitempty" json:"mapSheetDownloads,omitempty"`

	// Configuration specifically related to HTML/Web representation
	// +optional
	Web *WebConfig `yaml:"web,omitempty" json:"web,omitempty"`
}

// +kubebuilder:object:generate=true
type FeatureFilters struct {
	// OAF Part 1: filter on feature properties
	// https://docs.ogc.org/is/17-069r4/17-069r4.html#_parameters_for_filtering_on_feature_properties
	//
	// +optional
	Properties []PropertyFilter `yaml:"properties,omitempty" json:"properties,omitempty" validate:"dive"`

	// OAF Part 3: add config for complex/CQL filters here
	// <placeholder>
}

// +kubebuilder:object:generate=true
type CollectionLinks struct {
	// Links to downloads of entire collection. These will be rendered as rel=enclosure links
	// +optional
	Downloads []DownloadLink `yaml:"downloads,omitempty" json:"downloads,omitempty" validate:"dive"`

	// Links to documentation describing the collection. These will be rendered as rel=describedby links
	// <placeholder>
}

// +kubebuilder:object:generate=true
type DownloadLink struct {
	// Name of the provided download
	Name string `yaml:"name" json:"name" validate:"required"`

	// Full URL to the file to be downloaded
	AssetURL *URL `yaml:"assetUrl" json:"assetUrl" validate:"required"`

	// Approximate size of the file to be downloaded
	// +optional
	Size string `yaml:"size,omitempty" json:"size,omitempty"`

	// Media type of the file to be downloaded
	MediaType MediaType `yaml:"mediaType" json:"mediaType" validate:"required"`
}

// +kubebuilder:object:generate=true
type MapSheetDownloads struct {
	// Properties that provide the download details per map sheet. Note that 'map sheets' refer to a map
	// divided in rectangle areas that can be downloaded individually.
	Properties MapSheetDownloadProperties `yaml:"properties" json:"properties" validate:"required"`
}

// +kubebuilder:object:generate=true
type MapSheetDownloadProperties struct {
	// Property/column containing file download URL
	AssetURL string `yaml:"assetUrl" json:"assetUrl" validate:"required"`

	// Property/column containing file size
	Size string `yaml:"size" json:"size" validate:"required"`

	// The actual media type (not a property/column) of the download, like application/zip.
	MediaType MediaType `yaml:"mediaType" json:"mediaType" validate:"required"`

	// Property/column containing the map sheet identifier
	MapSheetID string `yaml:"mapSheetId" json:"mapSheetId" validate:"required"`
}

// +kubebuilder:object:generate=true
type WebConfig struct {
	// Viewer config for displaying multiple features on a map
	// +optional
	FeaturesViewer *FeaturesViewer `yaml:"featuresViewer,omitempty" json:"featuresViewer,omitempty"`

	// Viewer config for displaying a single feature on a map
	// +optional
	FeatureViewer *FeaturesViewer `yaml:"featureViewer,omitempty" json:"featureViewer,omitempty"`
}

// +kubebuilder:object:generate=true
type FeaturesViewer struct {
	// Maximum initial zoom level of the viewer when rendering features, specified by scale denominator.
	// Defaults to 1000 (= scale 1:1000).
	// +optional
	MinScale int `yaml:"minScale,omitempty" json:"minScale,omitempty" validate:"gt=0" default:"1000"`

	// Minimal initial zoom level of the viewer when rendering features, specified by scale denominator
	// (not set by default).
	// +optional
	MaxScale *int `yaml:"maxScale,omitempty" json:"maxScale,omitempty" validate:"omitempty,gt=0,gtefield=MinScale"`
}

// +kubebuilder:object:generate=true
type OgcAPI3dGeoVolumes struct {
	// Reference to the server (or object storage) hosting the 3D Tiles
	TileServer URL `yaml:"tileServer" json:"tileServer" validate:"required"`

	// Collections to be served as 3D GeoVolumes
	Collections GeoSpatialCollections `yaml:"collections" json:"collections"`

	// Whether JSON responses will be validated against the OpenAPI spec
	// since it has significant performance impact when dealing with large JSON payloads.
	//
	// +kubebuilder:default=true
	// +optional
	ValidateResponses *bool `yaml:"validateResponses,omitempty" json:"validateResponses,omitempty" default:"true"` // ptr due to https://github.com/creasty/defaults/issues/49
}

// +kubebuilder:validation:Enum=raster;vector
type TilesType string

const (
	TilesTypeRaster TilesType = "raster"
	TilesTypeVector TilesType = "vector"
)

// +kubebuilder:object:generate=true
type OgcAPITiles struct {
	// Reference to the server (or object storage) hosting the tiles
	TileServer URL `yaml:"tileServer" json:"tileServer" validate:"required"`

	// Could be 'vector' and/or 'raster' to indicate the types of tiles offered
	Types []TilesType `yaml:"types" json:"types" validate:"required"`

	// Specifies in what projections (SRS/CRS) the tiles are offered
	SupportedSrs []SupportedSrs `yaml:"supportedSrs" json:"supportedSrs" validate:"required,dive"`

	// Optional template to the vector tiles on the tileserver. Defaults to {tms}/{z}/{x}/{y}.pbf.
	// +optional
	URITemplateTiles *string `yaml:"uriTemplateTiles,omitempty" json:"uriTemplateTiles,omitempty"`

	// The collections to offer as tiles. When no collection is specified the tiles are hosted at the root of the API (/tiles endpoint).
	// +optional
	Collections GeoSpatialCollections `yaml:"collections,omitempty" json:"collections,omitempty"`
}

// +kubebuilder:object:generate=true
type OgcAPIStyles struct {
	// ID of the style to use a default
	Default string `yaml:"default" json:"default" validate:"required"`

	// Location on disk where the styles are hosted
	StylesDir string `yaml:"stylesDir" json:"stylesDir" validate:"required,dirpath|filepath"`

	// Styles exposed though this API
	SupportedStyles []Style `yaml:"supportedStyles" json:"supportedStyles" validate:"required,dive"`
}

// +kubebuilder:object:generate=true
type OgcAPIFeatures struct {
	// Basemap to use in embedded viewer on the HTML pages.
	// +kubebuilder:default="OSM"
	// +kubebuilder:validation:Enum=OSM;BRT
	// +optional
	Basemap string `yaml:"basemap,omitempty" json:"basemap,omitempty" default:"OSM" validate:"oneof=OSM BRT"`

	// Collections to be served as features through this API
	Collections GeoSpatialCollections `yaml:"collections" json:"collections" validate:"required,dive"`

	// Limits the amount of features to retrieve with a single call
	// +optional
	Limit Limit `yaml:"limit,omitempty" json:"limit,omitempty"`

	// One or more datasources to get the features from (geopackages, postgis, etc).
	// Optional since you can also define datasources at the collection level
	// +optional
	Datasources *Datasources `yaml:"datasources,omitempty" json:"datasources,omitempty"`

	// Whether GeoJSON/JSON-FG responses will be validated against the OpenAPI spec
	// since it has significant performance impact when dealing with large JSON payloads.
	//
	// +kubebuilder:default=true
	// +optional
	ValidateResponses *bool `yaml:"validateResponses,omitempty" json:"validateResponses,omitempty" default:"true"` // ptr due to https://github.com/creasty/defaults/issues/49
}

func (oaf *OgcAPIFeatures) ProjectionsForCollections() []string {
	return oaf.ProjectionsForCollection("")
}

func (oaf *OgcAPIFeatures) ProjectionsForCollection(collectionID string) []string {
	uniqueSRSs := make(map[string]struct{})
	if oaf.Datasources != nil {
		for _, a := range oaf.Datasources.Additional {
			uniqueSRSs[a.Srs] = struct{}{}
		}
	}
	for _, coll := range oaf.Collections {
		if (coll.ID == collectionID || collectionID == "") && coll.Features != nil && coll.Features.Datasources != nil {
			for _, a := range coll.Features.Datasources.Additional {
				uniqueSRSs[a.Srs] = struct{}{}
			}
			break
		}
	}
	result := util.Keys(uniqueSRSs)
	slices.Sort(result)
	return result
}

// +kubebuilder:object:generate=true
type OgcAPIProcesses struct {
	// Enable to advertise dismiss operations on the conformance page
	SupportsDismiss bool `yaml:"supportsDismiss" json:"supportsDismiss"`

	// Enable to advertise callback operations on the conformance page
	SupportsCallback bool `yaml:"supportsCallback" json:"supportsCallback"`

	// Reference to an external service implementing the process API. GoKoala acts only as a proxy for OGC API Processes.
	ProcessesServer URL `yaml:"processesServer" json:"processesServer" validate:"required"`
}

// +kubebuilder:object:generate=true
type Limit struct {
	// Number of features to return by default.
	// +kubebuilder:default=10
	// +kubebuilder:validation:Minimum=2
	// +optional
	Default int `yaml:"default,omitempty" json:"default,omitempty" validate:"gt=1" default:"10"`

	// Max number of features to return. Should be larger than 100 since the HTML interface always offers a 100 limit option.
	// +kubebuilder:default=1000
	// +kubebuilder:validation:Minimum=100
	// +optional
	Max int `yaml:"max,omitempty" json:"max,omitempty" validate:"gte=100" default:"1000"`
}

// +kubebuilder:object:generate=true
type Datasources struct {
	// Features should always be available in WGS84 (according to spec).
	// This specifies the datasource to be used for features in the WGS84 projection
	DefaultWGS84 Datasource `yaml:"defaultWGS84" json:"defaultWGS84" validate:"required"`

	// One or more additional datasources for features in other projections. GoKoala doesn't do
	// any on-the-fly reprojection so additional datasources need to be reprojected ahead of time.
	// +optional
	Additional []AdditionalDatasource `yaml:"additional" json:"additional" validate:"dive"`
}

// +kubebuilder:object:generate=true
type Datasource struct {
	// GeoPackage to get the features from.
	// +optional
	GeoPackage *GeoPackage `yaml:"geopackage,omitempty" json:"geopackage,omitempty" validate:"required_without_all=PostGIS"`

	// PostGIS database to get the features from (not implemented yet).
	// +optional
	PostGIS *PostGIS `yaml:"postgis,omitempty" json:"postgis,omitempty" validate:"required_without_all=GeoPackage"`

	// Add more datasources here such as Mongo, Elastic, etc
}

// +kubebuilder:object:generate=true
type AdditionalDatasource struct {
	// Projection (SRS/CRS) used for the features in this datasource
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	Srs string `yaml:"srs" json:"srs" validate:"required,startswith=EPSG:"`

	// The additional datasource
	Datasource `yaml:",inline" json:",inline"`
}

// +kubebuilder:object:generate=true
type PostGIS struct {
	// placeholder
}

// +kubebuilder:object:generate=true
type GeoPackage struct {
	// Settings to read a GeoPackage from local disk
	// +optional
	Local *GeoPackageLocal `yaml:"local,omitempty" json:"local,omitempty" validate:"required_without_all=Cloud"`

	// Settings to read a GeoPackage as a Cloud-Backed SQLite database
	// +optional
	Cloud *GeoPackageCloud `yaml:"cloud,omitempty" json:"cloud,omitempty" validate:"required_without_all=Local"`
}

// +kubebuilder:object:generate=true
type GeoPackageCommon struct {
	// Feature id column name
	// +kubebuilder:default="fid"
	// +optional
	Fid string `yaml:"fid,omitempty" json:"fid,omitempty" validate:"required" default:"fid"`

	// External feature id column name. When specified this ID column will be exposed to clients instead of the regular FID column.
	// It allows one to offer a more stable ID to clients instead of an auto-generated FID. External FID column should contain UUIDs.
	// +optional
	ExternalFid string `yaml:"externalFid" json:"externalFid"`

	// Optional timeout after which queries are canceled
	// +kubebuilder:default="15s"
	// +optional
	QueryTimeout Duration `yaml:"queryTimeout,omitempty" json:"queryTimeout,omitempty" validate:"required" default:"15s"`

	// ADVANCED SETTING. When the number of features in a bbox stay within the given value use an RTree index, otherwise use a BTree index
	// +kubebuilder:default=30000
	// +optional
	MaxBBoxSizeToUseWithRTree int `yaml:"maxBBoxSizeToUseWithRTree,omitempty" json:"maxBBoxSizeToUseWithRTree,omitempty" validate:"required" default:"30000"`

	// ADVANCED SETTING. Sets the SQLite "cache_size" pragma which determines how many pages are cached in-memory.
	// See https://sqlite.org/pragma.html#pragma_cache_size for details.
	// Default in SQLite is 2000 pages, which equates to 2000KiB (2048000 bytes). Which is denoted as -2000.
	// +kubebuilder:default=-2000
	// +optional
	InMemoryCacheSize int `yaml:"inMemoryCacheSize,omitempty" json:"inMemoryCacheSize,omitempty" validate:"required" default:"-2000"`
}

// +kubebuilder:object:generate=true
type GeoPackageLocal struct {
	// GeoPackageCommon shared config between local and cloud GeoPackage
	GeoPackageCommon `yaml:",inline" json:",inline"`

	// Location of GeoPackage on disk.
	// You can place the GeoPackage here manually (out-of-band) or you can specify Download
	// and let the application download the GeoPackage for you and store it at this location.
	File string `yaml:"file" json:"file" validate:"required,omitempty,filepath"`

	// Optional initialization task to download a GeoPackage during startup. GeoPackage will be
	// downloaded to local disk and stored at the location specified in File.
	// +optional
	Download *GeoPackageDownload `yaml:"download,omitempty" json:"download,omitempty"`
}

// +kubebuilder:object:generate=true
type GeoPackageDownload struct {
	// Location of GeoPackage on remote HTTP(S) URL. GeoPackage will be downloaded to local disk
	// during startup and stored at the location specified in "file".
	From URL `yaml:"from" json:"from" validate:"required"`

	// ADVANCED SETTING. Determines how many workers (goroutines) in parallel will download the specified GeoPackage.
	// Setting this to 1 will disable concurrent downloads.
	// +kubebuilder:default=4
	// +kubebuilder:validation:Minimum=1
	// +optional
	Parallelism int `yaml:"parallelism,omitempty" json:"parallelism,omitempty" validate:"required,gte=1" default:"4"`

	// ADVANCED SETTING. When true TLS certs are NOT validated, false otherwise. Only use true for your own self-signed certificates!
	// +kubebuilder:default=false
	// +optional
	TLSSkipVerify bool `yaml:"tlsSkipVerify,omitempty" json:"tlsSkipVerify,omitempty" default:"false"`

	// ADVANCED SETTING. HTTP request timeout when downloading (part of) GeoPackage.
	// +kubebuilder:default="2m"
	// +optional
	Timeout Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required" default:"2m"`

	// ADVANCED SETTING. Minimum delay to use when retrying HTTP request to download (part of) GeoPackage.
	// +kubebuilder:default="1s"
	// +optional
	RetryDelay Duration `yaml:"retryDelay,omitempty" json:"retryDelay,omitempty" validate:"required" default:"1s"`

	// ADVANCED SETTING. Maximum overall delay of the exponential backoff while retrying HTTP requests to download (part of) GeoPackage.
	// +kubebuilder:default="30s"
	// +optional
	RetryMaxDelay Duration `yaml:"retryMaxDelay,omitempty" json:"retryMaxDelay,omitempty" validate:"required" default:"30s"`

	// ADVANCED SETTING. Maximum number of retries when retrying HTTP requests to download (part of) GeoPackage.
	// +kubebuilder:default=5
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxRetries int `yaml:"maxRetries,omitempty" json:"maxRetries,omitempty" validate:"required,gte=1" default:"5"`
}

// +kubebuilder:object:generate=true
type GeoPackageCloud struct {
	// GeoPackageCommon shared config between local and cloud GeoPackage
	GeoPackageCommon `yaml:",inline" json:",inline"`

	// Reference to the cloud storage (either azure or google at the moment).
	// For example 'azure?emulator=127.0.0.1:10000&sas=0' or 'google'
	Connection string `yaml:"connection" json:"connection" validate:"required"`

	// Username of the storage account, like devstoreaccount1 when using Azurite
	User string `yaml:"user" json:"user" validate:"required"`

	// Some kind of credential like a password or key to authenticate with the storage backend, e.g:
	// 'Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==' when using Azurite
	Auth string `yaml:"auth" json:"auth" validate:"required"`

	// Container/bucket on the storage account
	Container string `yaml:"container" json:"container" validate:"required"`

	// Filename of the GeoPackage
	File string `yaml:"file" json:"file" validate:"required"`

	// Local cache of fetched blocks from cloud storage
	// +optional
	Cache GeoPackageCloudCache `yaml:"cache,omitempty" json:"cache,omitempty"`

	// ADVANCED SETTING. Only for debug purposes! When true all HTTP requests executed by sqlite to cloud object storage are logged to stdout
	// +kubebuilder:default=false
	// +optional
	LogHTTPRequests bool `yaml:"logHttpRequests,omitempty" json:"logHttpRequests,omitempty" default:"false"`
}

func (gc *GeoPackageCloud) CacheDir() (string, error) {
	fileNameWithoutExt := strings.TrimSuffix(gc.File, filepath.Ext(gc.File))
	if gc.Cache.Path != nil {
		randomSuffix := strconv.Itoa(rand.Intn(99999)) //nolint:gosec // random isn't used for security purposes
		return filepath.Join(*gc.Cache.Path, fileNameWithoutExt+"-"+randomSuffix), nil
	}
	cacheDir, err := os.MkdirTemp("", fileNameWithoutExt)
	if err != nil {
		return "", fmt.Errorf("failed to create tempdir to cache %s, error %w", fileNameWithoutExt, err)
	}
	return cacheDir, nil
}

// +kubebuilder:object:generate=true
type GeoPackageCloudCache struct {
	// Optional path to directory for caching cloud-backed GeoPackage blocks, when omitted a temp dir will be used.
	// +optional
	Path *string `yaml:"path,omitempty" json:"path,omitempty" validate:"omitempty,dirpath|filepath"`

	// Max size of the local cache. Accepts human-readable size such as 100Mb, 4Gb, 1Tb, etc. When omitted 1Gb is used.
	// +kubebuilder:default="1Gb"
	// +optional
	MaxSize string `yaml:"maxSize,omitempty" json:"maxSize,omitempty" default:"1Gb"`

	// When true a warm-up query is executed on startup which aims to fill the local cache. Does increase startup time.
	// +kubebuilder:default=false
	// +optional
	WarmUp bool `yaml:"warmUp,omitempty" json:"warmUp,omitempty" default:"false"`
}

func (cache *GeoPackageCloudCache) MaxSizeAsBytes() (int64, error) {
	return units.FromHumanSize(cache.MaxSize)
}

// +kubebuilder:object:generate=true
type PropertyFilter struct {
	// Needs to match with a column name in the feature table (in the configured datasource)
	Name string `yaml:"name" json:"name" validate:"required"`

	// Explains this property filter
	// +kubebuilder:default="Filter features by this property"
	// +optional
	Description string `yaml:"description,omitempty" json:"description,omitempty" default:"Filter features by this property"`

	// When true the property/column in the feature table needs to be indexed. Initialization will fail
	// when no index is present, when false the index check is skipped. For large tables an index is recommended!
	//
	// +kubebuilder:default=true
	// +optional
	IndexRequired *bool `yaml:"indexRequired,omitempty" json:"indexRequired,omitempty" default:"true"` // ptr due to https://github.com/creasty/defaults/issues/49

	// Static list of allowed values to be used as input for this property filter. Will be enforced by OpenAPI spec.
	// +optional
	AllowedValues []string `yaml:"allowedValues,omitempty" json:"allowedValues,omitempty"`

	// Derive list of allowed values for this property filter from the corresponding column in the datastore.
	// Use with caution since it can increase startup time when used on large tables. Make sure an index in present.
	//
	// +kubebuilder:default=false
	// +optional
	DeriveAllowedValuesFromDatasource *bool `yaml:"deriveAllowedValuesFromDatasource,omitempty" json:"deriveAllowedValuesFromDatasource,omitempty" default:"false"`
}

// +kubebuilder:object:generate=true
type SupportedSrs struct {
	// Projection (SRS/CRS) used
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	Srs string `yaml:"srs" json:"srs" validate:"required,startswith=EPSG:"`

	// Available zoom levels
	ZoomLevelRange ZoomLevelRange `yaml:"zoomLevelRange" json:"zoomLevelRange" validate:"required"`
}

// +kubebuilder:object:generate=true
type ZoomLevelRange struct {
	// Start zoom level
	// +kubebuilder:validation:Minimum=0
	Start int `yaml:"start" json:"start" validate:"gte=0,ltefield=End"`

	// End zoom level
	End int `yaml:"end" json:"end" validate:"required,gtefield=Start"`
}

// +kubebuilder:object:generate=true
type Extent struct {
	// Projection (SRS/CRS) to be used. When none is provided WGS84 (http://www.opengis.net/def/crs/OGC/1.3/CRS84) is used.
	// +optional
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	Srs string `yaml:"srs,omitempty" json:"srs,omitempty" validate:"omitempty,startswith=EPSG:"`

	// Geospatial extent
	Bbox []string `yaml:"bbox" json:"bbox"`

	// Temporal extent
	// +optional
	// +kubebuilder:validation:MinItems=2
	// +kubebuilder:validation:MaxItems=2
	Interval []string `yaml:"interval,omitempty" json:"interval,omitempty" validate:"omitempty,len=2"`
}

// +kubebuilder:object:generate=true
type TemporalProperties struct {
	// Name of field in datasource to be used in temporal queries as the start date
	StartDate string `yaml:"startDate" json:"startDate" validate:"required"`

	// Name of field in datasource to be used in temporal queries as the end date
	EndDate string `yaml:"endDate" json:"endDate" validate:"required"`
}

// +kubebuilder:object:generate=true
type License struct {
	// Name of the license, e.g. MIT, CC0, etc
	Name string `yaml:"name" json:"name" validate:"required"`

	// URL to license text on the web
	URL URL `yaml:"url" json:"url" validate:"required"`
}

// +kubebuilder:object:generate=true
type Style struct {
	// Unique ID of this style
	ID string `yaml:"id" json:"id" validate:"required"`

	// Human-friendly name of this style
	Title string `yaml:"title" json:"title" validate:"required"`

	// Explains what is visualized by this style
	// +optional
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`

	// Keywords to make this style better discoverable
	// +optional
	Keywords []string `yaml:"keywords,omitempty" json:"keywords,omitempty"`

	// Moment in time when the style was last updated
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format="date-time"
	LastUpdated *string `yaml:"lastUpdated,omitempty" json:"lastUpdated,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z"`

	// Optional version of this style
	// +optional
	Version *string `yaml:"version,omitempty" json:"version,omitempty"`

	// Reference to a PNG image to use a thumbnail on the style metadata page.
	// The full path is constructed by appending Resources + Thumbnail.
	// +optional
	Thumbnail *string `yaml:"thumbnail,omitempty" json:"thumbnail,omitempty"`

	// This style is offered in the following formats
	Formats []StyleFormat `yaml:"formats" json:"formats" validate:"required,dive"`
}

// +kubebuilder:object:generate=true
type StyleFormat struct {
	// Name of the format
	// +kubebuilder:default="mapbox"
	// +optional
	Format string `yaml:"format,omitempty" json:"format,omitempty" default:"mapbox" validate:"required,oneof=mapbox sld10"`
}
