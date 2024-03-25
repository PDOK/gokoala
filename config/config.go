//go:generate ../hack/generate-deepcopy.sh
package config

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/engine/util"
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
		return validateCollectionsTemporalConfig(config.OgcAPI.Features.Collections)
	}
	return nil
}

func validateCollectionsTemporalConfig(collections GeoSpatialCollections) error {
	var errMessages []string
	for _, collection := range collections {
		if collection.Metadata != nil && collection.Metadata.TemporalProperties != nil && collection.Metadata.Extent.Interval == nil {
			errMessages = append(errMessages, fmt.Sprintf("validation failed for collection '%s'; "+
				"field 'Extent.Interval' is required with field 'TemporalProperties'\n", collection.ID))
		}
	}
	if len(errMessages) > 0 {
		return fmt.Errorf("invalid config provided:\n%v", errMessages)
	}
	return nil
}

// +kubebuilder:object:generate=true
type Config struct {
	Version            string     `yaml:"version" json:"version" validate:"required,semver"`
	Title              string     `yaml:"title" json:"title"  validate:"required"`
	ServiceIdentifier  string     `yaml:"serviceIdentifier"  json:"serviceIdentifier" validate:"required"`
	Abstract           string     `yaml:"abstract" json:"abstract" validate:"required"`
	License            License    `yaml:"license" json:"license" validate:"required"`
	BaseURL            URL        `yaml:"baseUrl" json:"baseUrl" validate:"required,url"`
	DatasetCatalogURL  URL        `yaml:"datasetCatalogUrl" json:"datasetCatalogUrl" validate:"url"`
	AvailableLanguages []Language `yaml:"availableLanguages" json:"availableLanguages"`
	OgcAPI             OgcAPI     `yaml:"ogcApi" json:"ogcApi" validate:"required"`
	// +optional
	Thumbnail *string `yaml:"thumbnail" json:"thumbnail"`
	// +optional
	Keywords []string `yaml:"keywords" json:"keywords"`
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format="date-time"
	LastUpdated *string `yaml:"lastUpdated" json:"lastUpdated"`
	// +optional
	LastUpdatedBy string `yaml:"lastUpdatedBy" json:"lastUpdatedBy"`
	// +optional
	Support *Support `yaml:"support" json:"support"`
	// +optional
	DatasetDetails []DatasetDetail `yaml:"datasetDetails" json:"datasetDetails"`
	// +optional
	DatasetMetadata DatasetMetadata `yaml:"datasetMetadata" json:"datasetMetadata"`
	// +optional
	Resources *Resources `yaml:"resources" json:"resources"`
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
	Name string `yaml:"name" json:"name" validate:"required"`
	// +kubebuilder:validation:Type=string
	URL URL `yaml:"url" json:"url" validate:"required,url"`

	// +optional
	Email string `yaml:"email" json:"email" validate:"omitempty,email"`
}

// +kubebuilder:object:generate=true
type DatasetDetail struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
}

// +kubebuilder:object:generate=true
type DatasetMetadata struct {
	Source string `yaml:"source" json:"source"`

	// +optional
	API *string `yaml:"api" json:"api" validate:"omitempty,url"`
	// +optional
	Dataset *string `yaml:"dataset" json:"dataset" validate:"omitempty,url"`
}

// +kubebuilder:object:generate=true
type Resources struct {
	// This is optional if Directory is set
	// +optional
	URL URL `yaml:"url" json:"url" validate:"required_without=Directory,omitempty,url"`
	// This is optional if URL is set
	// +optional
	Directory string `yaml:"directory" json:"directory" validate:"required_without=URL,omitempty,dir"`
}

// +kubebuilder:object:generate=true
type OgcAPI struct {
	// +optional
	GeoVolumes *OgcAPI3dGeoVolumes `yaml:"3dgeovolumes" json:"3dgeovolumes"`
	// +optional
	Tiles *OgcAPITiles `yaml:"tiles" json:"tiles" validate:"required_with=Styles"`
	// +optional
	Styles *OgcAPIStyles `yaml:"styles" json:"styles"`
	// +optional
	Features *OgcAPIFeatures `yaml:"features" json:"features"`
	// +optional
	Processes *OgcAPIProcesses `yaml:"processes" json:"processes"`
}

// +kubebuilder:object:generate=true
type GeoSpatialCollection struct {
	ID string `yaml:"id" json:"id" validate:"required"`

	// +optional
	Metadata *GeoSpatialCollectionMetadata `yaml:"metadata" json:"metadata"`

	// +optional
	GeoVolumes *CollectionEntry3dGeoVolumes `yaml:",inline" json:",inline"`
	// +optional
	Tiles *CollectionEntryTiles `yaml:",inline" json:",inline"`
	// +optional
	Features *CollectionEntryFeatures `yaml:",inline" json:",inline"`
}

// +kubebuilder:object:generate=true
type GeoSpatialCollectionMetadata struct {
	// +optional
	Title       *string `yaml:"title" json:"title"`
	Description *string `yaml:"description" json:"description" validate:"required"`
	// +optional
	Thumbnail *string `yaml:"thumbnail" json:"thumbnail"`
	// +optional
	Keywords []string `yaml:"keywords" json:"keywords"`
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format="date-time"
	LastUpdated *string `yaml:"lastUpdated" json:"lastUpdated"`
	// +optional
	LastUpdatedBy string `yaml:"lastUpdatedBy" json:"lastUpdatedBy"`
	// +optional
	TemporalProperties *TemporalProperties `yaml:"temporalProperties" json:"temporalProperties" validate:"omitempty,required_with=Extent.Interval"`
	// +optional
	Extent *Extent `yaml:"extent" json:"extent"`
}

// +kubebuilder:object:generate=true
type CollectionEntry3dGeoVolumes struct {
	// Optional basepath to 3D tiles on the tileserver. Defaults to the collection ID.
	// +optional
	TileServerPath *string `yaml:"tileServerPath" json:"tileServerPath"`

	// URI template for individual 3D tiles.
	// +optional
	URITemplate3dTiles *string `yaml:"uriTemplate3dTiles" json:"uriTemplate3dTiles" validate:"required_without_all=URITemplateDTM"`

	// Optional URI template for subtrees, only required when "implicit tiling" extension is used.
	// +optional
	URITemplateImplicitTilingSubtree *string `yaml:"uriTemplateImplicitTilingSubtree" json:"uriTemplateImplicitTilingSubtree"`

	// URI template for digital terrain model (DTM) in Quantized Mesh format, REQUIRED when you want to serve a DTM.
	// +optional
	URITemplateDTM *string `yaml:"uriTemplateDTM" json:"uriTemplateDTM" validate:"required_without_all=URITemplate3dTiles"`

	// Optional URL to 3D viewer to visualize the given collection of 3D Tiles.
	// +optional
	URL3DViewer *URL `yaml:"3dViewerUrl" json:"3dViewerUrl" validate:"url"`
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
	TableName *string `yaml:"tableName" json:"tableName"`

	// Optional collection specific datasources. Mutually exclusive with top-level defined datasources.
	// +optional
	Datasources *Datasources `yaml:"datasources" json:"datasources"`

	// +optional
	Filters FeatureFilters `yaml:"filters" json:"filters"`
}

// +kubebuilder:object:generate=true
type FeatureFilters struct {
	// OAF Part 1: filter on feature properties
	// https://docs.ogc.org/is/17-069r4/17-069r4.html#_parameters_for_filtering_on_feature_properties
	// +optional
	Properties []PropertyFilter `yaml:"properties" json:"properties" validate:"dive"`

	// OAF Part 3: add config for complex/CQL filters here
	// <placeholder>
}

// +kubebuilder:object:generate=true
type OgcAPI3dGeoVolumes struct {
	TileServer  URL                   `yaml:"tileServer" json:"tileServer" validate:"required,url"`
	Collections GeoSpatialCollections `yaml:"collections" json:"collections"`
}

// +kubebuilder:object:generate=true
type OgcAPITiles struct {
	TileServer   URL            `yaml:"tileServer" json:"tileServer" validate:"required,url"`
	Types        []string       `yaml:"types" json:"types" validate:"required"`
	SupportedSrs []SupportedSrs `yaml:"supportedSrs" json:"supportedSrs" validate:"required,dive"`
	// Optional template to the vector tiles on the tileserver. Defaults to {tms}/{z}/{x}/{y}.pbf.
	// +optional
	URITemplateTiles *string `yaml:"uriTemplateTiles" json:"uriTemplateTiles"`
	// +optional
	Collections GeoSpatialCollections `yaml:"collections" json:"collections"`
}

// +kubebuilder:object:generate=true
type OgcAPIStyles struct {
	Default         string          `yaml:"default" json:"default" validate:"required"`
	StylesDir       string          `yaml:"stylesDir" json:"stylesDir" validate:"required,dir"`
	SupportedStyles []StyleMetadata `yaml:"supportedStyles" json:"supportedStyles" validate:"required,dive"`
}

// +kubebuilder:object:generate=true
type OgcAPIFeatures struct {
	// +kubebuilder:default="OSM"
	// +kubebuilder:validation:Enum=OSM;BRT
	Basemap     string                `yaml:"basemap" json:"basemap" default:"OSM" validate:"oneof=OSM BRT"`
	Collections GeoSpatialCollections `yaml:"collections" json:"collections" validate:"required,dive"`
	// +optional
	Limit Limit `yaml:"limit" json:"limit"`
	// +optional
	Datasources *Datasources `yaml:"datasources" json:"datasources"` // optional since you can also define datasources at the collection level

	// Whether GeoJSON/JSON-FG responses will be validated against the OpenAPI spec
	// since it has significant performance impact when dealing with large JSON payloads.
	// +kubebuilder:default=true
	ValidateResponses *bool `yaml:"validateResponses" json:"validateResponses" default:"true"` // ptr due to https://github.com/creasty/defaults/issues/49
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

func (oaf *OgcAPIFeatures) PropertyFiltersForCollection(collectionID string) []PropertyFilter {
	for _, coll := range oaf.Collections {
		if coll.ID == collectionID && coll.Features != nil && coll.Features.Filters.Properties != nil {
			return coll.Features.Filters.Properties
		}
	}
	return []PropertyFilter{}
}

// +kubebuilder:object:generate=true
type OgcAPIProcesses struct {
	SupportsDismiss  bool `yaml:"supportsDismiss" json:"supportsDismiss"`
	SupportsCallback bool `yaml:"supportsCallback" json:"supportsCallback"`
	ProcessesServer  URL  `yaml:"processesServer" json:"processesServer" validate:"required,url"`
}

// +kubebuilder:object:generate=true
type Limit struct {
	// +kubebuilder:default=10
	// +kubebuilder:validation:Minimum=2
	Default int `yaml:"default" json:"default" validate:"gt=1" default:"10"`
	// +kubebuilder:default=1000
	// +kubebuilder:validation:Minimum=100
	Max int `yaml:"max" json:"max" validate:"gte=100" default:"1000"`
}

// +kubebuilder:object:generate=true
type Datasources struct {
	DefaultWGS84 Datasource             `yaml:"defaultWGS84" json:"defaultWGS84" validate:"required"`
	Additional   []AdditionalDatasource `yaml:"additional" json:"additional" validate:"dive"`
}

// +kubebuilder:object:generate=true
type Datasource struct {
	// +optional
	GeoPackage *GeoPackage `yaml:"geopackage" json:"geopackage" validate:"required_without_all=PostGIS"`
	// +optional
	PostGIS *PostGIS `yaml:"postgis" json:"postgis" validate:"required_without_all=GeoPackage"`
	// Add more datasources here such as Mongo, Elastic, etc
}

// +kubebuilder:object:generate=true
type AdditionalDatasource struct {
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	Srs        string `yaml:"srs" json:"srs" validate:"required,startswith=EPSG:"`
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
	Local *GeoPackageLocal `yaml:"local" json:"local" validate:"required_without_all=Cloud"`
	// Settings to read a GeoPackage as a Cloud-Backed SQLite database
	// +optional
	Cloud *GeoPackageCloud `yaml:"cloud" json:"cloud" validate:"required_without_all=Local"`
}

// +kubebuilder:object:generate=true
type GeoPackageCommon struct {
	// feature id column name
	// +kubebuilder:default="fid"
	Fid string `yaml:"fid" json:"fid" validate:"required" default:"fid"`

	// optional timeout after which queries are canceled
	// +kubebuilder:default="15s"
	QueryTimeout Duration `yaml:"queryTimeout" json:"queryTimeout" validate:"required" default:"15s"`

	// when the number of features in a bbox stay within the given value use an RTree index, otherwise use a BTree index
	// +kubebuilder:default=30000
	MaxBBoxSizeToUseWithRTree int `yaml:"maxBBoxSizeToUseWithRTree" json:"maxBBoxSizeToUseWithRTree" validate:"required" default:"30000"`
}

// +kubebuilder:object:generate=true
type GeoPackageLocal struct {
	// GeoPackageCommon shared config between local and cloud GeoPackage
	GeoPackageCommon `yaml:",inline" json:",inline"`

	// location of GeoPackage on disk
	File string `yaml:"file" json:"file" validate:"file"`
}

// +kubebuilder:object:generate=true
type GeoPackageCloud struct {
	// GeoPackageCommon shared config between local and cloud GeoPackage
	GeoPackageCommon `yaml:",inline" json:",inline"`

	// reference to the cloud storage (either azure or google at the moment), e.g:
	// - azure?emulator=127.0.0.1:10000&sas=0
	// - google
	Connection string `yaml:"connection" json:"connection" validate:"required"`

	// username of the storage account, e.g: devstoreaccount1 when using Azurite
	User string `yaml:"user" json:"user" validate:"required"`

	// some kind of credential like a password or key to authenticate with the storage backend, e.g:
	// 'Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==' when using Azurite
	Auth string `yaml:"auth" json:"auth" validate:"required"`

	// container/bucket on the storage account
	Container string `yaml:"container" json:"container" validate:"required"`

	// filename of the GeoPackage
	File string `yaml:"file" json:"file" validate:"required"`

	// local cache of fetched blocks from cloud storage
	// +optional
	Cache GeoPackageCloudCache `yaml:"cache" json:"cache"`

	// only for debug purposes! When true all HTTP requests executed by sqlite to cloud object storage are logged to stdout
	// +kubebuilder:default=false
	LogHTTPRequests bool `yaml:"logHttpRequests" json:"logHttpRequests" default:"false"`
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
	// optional path to directory for caching cloud-backed GeoPackage blocks, when omitted a temp dir will be used.
	// +optional
	Path *string `yaml:"path" json:"path" validate:"omitempty,dir"`

	// max size of the local cache. Accepts human-readable size such as 100Mb, 4Gb, 1Tb, etc. When omitted 1Gb is used.
	// +kubebuilder:default="1Gb"
	MaxSize string `yaml:"maxSize" json:"maxSize" default:"1Gb"`

	// when true a warm-up query is executed on startup which aims to fill the local cache. Does increase startup time.
	// +kubebuilder:default=false
	WarmUp bool `yaml:"warmUp" json:"warmUp" default:"false"`
}

func (cache *GeoPackageCloudCache) MaxSizeAsBytes() (int64, error) {
	return units.FromHumanSize(cache.MaxSize)
}

// +kubebuilder:object:generate=true
type PropertyFilter struct {
	// needs to match with a column name in the feature table (in the configured datasource)
	Name string `yaml:"name" json:"name" validate:"required"`
	// +kubebuilder:default="Filter features by this property"
	Description string `yaml:"description" json:"description" default:"Filter features by this property"`
}

// +kubebuilder:object:generate=true
type SupportedSrs struct {
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	Srs            string         `yaml:"srs" json:"srs" validate:"required,startswith=EPSG:"`
	ZoomLevelRange ZoomLevelRange `yaml:"zoomLevelRange" json:"zoomLevelRange" validate:"required"`
}

// +kubebuilder:object:generate=true
type ZoomLevelRange struct {
	// +kubebuilder:validation:Minimum=0
	Start int `yaml:"start" json:"start" validate:"gte=0,ltefield=End"`
	End   int `yaml:"end" json:"end" validate:"required,gtefield=Start"`
}

// +kubebuilder:object:generate=true
type Extent struct {
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	Srs  string   `yaml:"srs" json:"srs" validate:"required,startswith=EPSG:"`
	Bbox []string `yaml:"bbox" json:"bbox"`
	// +optional
	// +kubebuilder:validation:MinItems=2
	// +kubebuilder:validation:MaxItems=2
	Interval []string `yaml:"interval" json:"interval" validate:"omitempty,len=2"`
}

// +kubebuilder:object:generate=true
type TemporalProperties struct {
	StartDate string `yaml:"startDate" json:"startDate" validate:"required"`
	EndDate   string `yaml:"endDate" json:"endDate" validate:"required"`
}

// +kubebuilder:object:generate=true
type License struct {
	Name string `yaml:"name" json:"name" validate:"required"`
	URL  URL    `yaml:"url" json:"url" validate:"required,url"`
}

// +kubebuilder:object:generate=true
type StyleMetadata struct {
	ID    string `yaml:"id" json:"id" validate:"required"`
	Title string `yaml:"title" json:"title" validate:"required"`
	// +optional
	Description *string `yaml:"description" json:"description"`
	// +optional
	Keywords []string `yaml:"keywords" json:"keywords"`
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format="date-time"
	LastUpdated *string `yaml:"lastUpdated" json:"lastUpdated" validate:"omitempty,datetime=2006-01-02T15:04:05Z"`
	// +optional
	Version *string `yaml:"version" json:"version"`
	// +optional
	Thumbnail   *string      `yaml:"thumbnail" json:"thumbnail"`
	Stylesheets []StyleSheet `yaml:"stylesheets" json:"stylesheets" validate:"required,dive"`
}

// +kubebuilder:object:generate=true
type StyleSheet struct {
	// +kubebuilder:default="mapbox"
	Format string `yaml:"format" json:"format" default:"mapbox" validate:"required,oneof=mapbox sld10"`
}
