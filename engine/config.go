package engine

import (
	"errors"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

const (
	cookieMaxAge        = 60 * 60 * 24
	defaultQueryTimeout = 10 * time.Second
)

func readConfigFile(configFile string) *Config {
	yamlData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config file %v", err)
	}

	// expand environment variables
	yamlData = []byte(os.ExpandEnv(string(yamlData)))

	var config *Config
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		log.Fatalf("failed to unmarshal config file %v", err)
	}

	setDefaults(config)
	validate(config)
	return config
}

func setDefaults(config *Config) {
	// process 'default' tags
	if err := defaults.Set(config); err != nil {
		log.Fatalf("failed to set default configuration: %v", err)
	}

	config.CookieMaxAge = cookieMaxAge

	if len(config.AvailableLanguages) == 0 {
		config.AvailableLanguages = append(config.AvailableLanguages, language.Dutch) // default to Dutch only
	}
}

func validate(config *Config) {
	v := validator.New()
	err := v.Struct(config)
	if err != nil {
		var ive *validator.InvalidValidationError
		if ok := errors.Is(err, ive); ok {
			log.Fatalf("failed to validate config file: %v", err)
		}
		var errMessages []string
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			for _, valErr := range valErrs {
				errMessages = append(errMessages, valErr.Error()+"\n")
			}
		}
		log.Fatalf("invalid config file provided:\n %v", errMessages)
	}
}

type Config struct {
	Version            string          `yaml:"version" validate:"required,semver"`
	Title              string          `yaml:"title" validate:"required"`
	ServiceIdentifier  string          `yaml:"serviceIdentifier" validate:"required"`
	Abstract           string          `yaml:"abstract" validate:"required"`
	Thumbnail          *string         `yaml:"thumbnail"`
	Keywords           []string        `yaml:"keywords"`
	LastUpdated        *string         `yaml:"lastUpdated"`
	LastUpdatedBy      string          `yaml:"lastUpdatedBy"`
	License            License         `yaml:"license" validate:"required"`
	Support            *Support        `yaml:"support"`
	DatasetDetails     []DatasetDetail `yaml:"datasetDetails"`
	DatasetMetadata    DatasetMetadata `yaml:"datasetMetadata"`
	DatasetCatalogURL  YAMLURL         `yaml:"datasetCatalogUrl" validate:"url"`
	BaseURL            YAMLURL         `yaml:"baseUrl" validate:"required,url"`
	Resources          *Resources      `yaml:"resources"`
	AvailableLanguages []language.Tag  `yaml:"availableLanguages"`
	OgcAPI             OgcAPI          `yaml:"ogcApi" validate:"required"`
	CookieMaxAge       int
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
	if c.OgcAPI.Maps != nil {
		result = append(result, c.OgcAPI.Maps.Collections...)
	}
	return result
}

type Support struct {
	Name  string `yaml:"name" validate:"required"`
	Email string `yaml:"email" validate:"omitempty,email"`
	URL   string `yaml:"url" validate:"required,url"`
}

type DatasetDetail struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type DatasetMetadata struct {
	Source  string  `yaml:"source"`
	API     *string `yaml:"api" validate:"omitempty,url"`
	Dataset *string `yaml:"dataset" validate:"omitempty,url"`
}

type Resources struct {
	URL       YAMLURL `yaml:"url" validate:"required_without=Directory,omitempty,url"`
	Directory string  `yaml:"directory" validate:"required_without=URL,omitempty,dir"`
}

type OgcAPI struct {
	GeoVolumes *OgcAPI3dGeoVolumes `yaml:"3dgeovolumes"`
	Tiles      *OgcAPITiles        `yaml:"tiles" validate:"required_with=Styles"`
	Styles     *OgcAPIStyles       `yaml:"styles"`
	Features   *OgcAPIFeatures     `yaml:"features"`
	Maps       *OgcAPIMaps         `yaml:"maps"`
	Processes  *OgcAPIProcesses    `yaml:"processes"`
}

type GeoSpatialCollections []GeoSpatialCollection

// Unique lists all unique GeoSpatialCollections (no duplicate IDs),
// return results in alphabetic order
func (g GeoSpatialCollections) Unique() []GeoSpatialCollection {
	collectionsByID := g.toMap()
	flattened := make([]GeoSpatialCollection, 0, len(collectionsByID))
	for _, v := range collectionsByID {
		flattened = append(flattened, v)
	}
	sort.Slice(flattened, func(i, j int) bool {
		icomp := flattened[i].ID
		jcomp := flattened[j].ID
		// prefer to sort by title when available, collection ID otherwise
		if flattened[i].Metadata != nil && flattened[i].Metadata.Title != nil {
			icomp = *flattened[i].Metadata.Title
		}
		if flattened[j].Metadata != nil && flattened[j].Metadata.Title != nil {
			jcomp = *flattened[j].Metadata.Title
		}
		return icomp < jcomp
	})
	return flattened
}

// ContainsID check if given collection - by ID - exists
func (g GeoSpatialCollections) ContainsID(id string) bool {
	_, ok := g.toMap()[id]
	return ok
}

func (g GeoSpatialCollections) toMap() map[string]GeoSpatialCollection {
	collectionsByID := make(map[string]GeoSpatialCollection)
	for _, v := range g {
		collectionsByID[v.ID] = v
	}
	return collectionsByID
}

type GeoSpatialCollection struct {
	ID       string                        `yaml:"id" validate:"required"`
	Metadata *GeoSpatialCollectionMetadata `yaml:"metadata"`

	GeoVolumes *CollectionEntry3dGeoVolumes `yaml:",inline"`
	Tiles      *CollectionEntryTiles        `yaml:",inline"`
	Features   *CollectionEntryFeatures     `yaml:",inline"`
	Maps       *CollectionEntryMaps         `yaml:",inline"`
}

type GeoSpatialCollectionMetadata struct {
	Title         *string  `yaml:"title"`
	Description   *string  `yaml:"description"`
	Thumbnail     *string  `yaml:"thumbnail"`
	Keywords      []string `yaml:"keywords"`
	LastUpdated   *string  `yaml:"lastUpdated"`
	LastUpdatedBy string   `yaml:"lastUpdatedBy"`
	Extent        *Extent  `yaml:"extent"`
}

type CollectionEntry3dGeoVolumes struct {
	// Optional basepath to 3D tiles on the tileserver. Defaults to the collection ID.
	TileServerPath *string `yaml:"tileServerPath"`

	// URI template for individual 3D tiles.
	URITemplate3dTiles *string `yaml:"uriTemplate3dTiles" validate:"required_without_all=URITemplateDTM"`

	// Optional URI template for subtrees, only required when "implicit tiling" extension is used.
	URITemplateImplicitTilingSubtree *string `yaml:"uriTemplateImplicitTilingSubtree"`

	// URI template for digital terrain model (DTM) in Quantized Mesh format, REQUIRED when you want to serve a DTM.
	URITemplateDTM *string `yaml:"uriTemplateDTM" validate:"required_without_all=URITemplate3dTiles"`

	// Optional URL to 3D viewer to visualize the given collection of 3D Tiles.
	URL3DViewer *YAMLURL `yaml:"3dViewerUrl" validate:"url"`
}

func (gv *CollectionEntry3dGeoVolumes) Has3DTiles() bool {
	return gv.URITemplate3dTiles != nil
}

func (gv *CollectionEntry3dGeoVolumes) HasDTM() bool {
	return gv.URITemplateDTM != nil
}

type CollectionEntryTiles struct {
	// placeholder
}

type CollectionEntryFeatures struct {
	// Optional way to explicitly map a collection ID to the underlying table in the datasource.
	TableName *string `yaml:"tableName"`

	// Optional collection specific datasources. Mutually exclusive with top-level defined datasources.
	Datasources *Datasources `yaml:"datasources" validate:"dive"`
}

type CollectionEntryMaps struct {
	// placeholder
}

type OgcAPI3dGeoVolumes struct {
	TileServer  YAMLURL               `yaml:"tileServer" validate:"required,url"`
	Collections GeoSpatialCollections `yaml:"collections"`
}

type OgcAPITiles struct {
	TileServer YAMLURL `yaml:"tileServer" validate:"required,url"`
	// Optional template to the vector tiles on the tileserver. Defaults to {tms}/{z}/{x}/{y}.pbf.
	URITemplateTiles *string               `yaml:"uriTemplateTiles"`
	Types            []string              `yaml:"types" validate:"required"`
	SupportedSrs     []SupportedSrs        `yaml:"supportedSrs" validate:"required,dive"`
	Collections      GeoSpatialCollections `yaml:"collections"`
}

type OgcAPIStyles struct {
	Default          string          `yaml:"default" validate:"required"`
	MapboxStylesPath string          `yaml:"mapboxStylesPath" validate:"required,dir"`
	SupportedStyles  []StyleMetadata `yaml:"supportedStyles" validate:"required"`
}

type OgcAPIFeatures struct {
	Limit       Limit                 `yaml:"limit"`
	Datasources *Datasources          `yaml:"datasources"`
	Collections GeoSpatialCollections `yaml:"collections" validate:"required"`
}

type OgcAPIMaps struct {
	Collections GeoSpatialCollections `yaml:"collections"`
}

type OgcAPIProcesses struct {
	SupportsDismiss  bool    `yaml:"supportsDismiss"`
	SupportsCallback bool    `yaml:"supportsCallback"`
	ProcessesServer  YAMLURL `yaml:"processesServer" validate:"url"`
}

type Limit struct {
	Default int `yaml:"default" validate:"gt=1" default:"10"`
	Max     int `yaml:"max" validate:"gt=1" default:"1000"`
}

type Datasources struct {
	DefaultWGS84 Datasource             `yaml:"defaultWGS84" validate:"required"`
	Additional   []AdditionalDatasource `yaml:"additional" validate:"dive"`
}

type Datasource struct {
	GeoPackage *GeoPackage `yaml:"geopackage" validate:"required_without_all=PostGIS"`
	PostGIS    *PostGIS    `yaml:"postgis" validate:"required_without_all=GeoPackage"`
	// Add more datasources here such as Mongo, Elastic, etc
}

type AdditionalDatasource struct {
	Srs        string `yaml:"srs" validate:"required,startswith=EPSG:"`
	Datasource `yaml:",inline"`
}

type PostGIS struct {
	// placeholder
}

type GeoPackage struct {
	Local *GeoPackageLocal `yaml:"local" validate:"required_without_all=Cloud"`
	Cloud *GeoPackageCloud `yaml:"cloud" validate:"required_without_all=Local"`
}

// GeoPackageCommon shared config between local and cloud GeoPackage
type GeoPackageCommon struct {
	// feature id column name
	Fid string `yaml:"fid" validate:"required"`

	// optional timeout after which queries are canceled (default is 10s, see constant)
	QueryTimeout *time.Duration `yaml:"queryTimeout"`
}

func (gc *GeoPackageCommon) GetQueryTimeout() time.Duration {
	if gc.QueryTimeout != nil {
		return *gc.QueryTimeout
	}
	return defaultQueryTimeout
}

// GeoPackageLocal settings to read a GeoPackage from local disk
type GeoPackageLocal struct {
	GeoPackageCommon `yaml:",inline"`

	// location of GeoPackage on disk
	File string `yaml:"file" validate:"file"`
}

// GeoPackageCloud settings to read a GeoPackage as a Cloud-Backed SQLite database
type GeoPackageCloud struct {
	GeoPackageCommon `yaml:",inline"`

	// reference to the cloud storage (either azure or google at the moment), e.g:
	// - azure?emulator=127.0.0.1:10000&sas=0
	// - google
	Connection string `yaml:"connection" validate:"required"`

	// username of the storage account, e.g: devstoreaccount1 when using Azurite
	User string `yaml:"user" validate:"required"`

	// some kind of credential like a password or key to authenticate with the storage backend, e.g:
	// 'Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==' when using Azurite
	Auth string `yaml:"auth" validate:"required"`

	// container/bucket on the storage account
	Container string `yaml:"container" validate:"required"`

	// filename of the GeoPackage
	File string `yaml:"file" validate:"required"`

	// local cache of fetched blocks from cloud storage
	Cache *string `yaml:"cache" validate:"omitempty,dir"`
}

type SupportedSrs struct {
	Srs            string         `yaml:"srs" validate:"required,startswith=EPSG:"`
	ZoomLevelRange ZoomLevelRange `yaml:"zoomLevelRange" validate:"required"`
}

type ZoomLevelRange struct {
	Start int `yaml:"start" validate:"gte=0,ltefield=End"`
	End   int `yaml:"end" validate:"required,gtefield=Start"`
}

type Extent struct {
	Srs  string   `yaml:"srs" validate:"required,startswith=EPSG:"`
	Bbox []string `yaml:"bbox"`
}

type License struct {
	Name string `yaml:"name" validate:"required"`
	URL  string `yaml:"url" validate:"required,url"`
}

// StyleMetadata based on OGC API Styles Requirement 7B
type StyleMetadata struct {
	ID             string       `yaml:"id" json:"id"`
	Title          string       `yaml:"title" json:"title,omitempty"`
	Description    *string      `yaml:"description" json:"description,omitempty"`
	Keywords       []string     `yaml:"keywords" json:"keywords,omitempty"`
	PointOfContact *string      `yaml:"pointOfContact" json:"pointOfContact,omitempty"`
	License        *string      `yaml:"license" json:"license,omitempty"`
	Created        *string      `yaml:"created" json:"created,omitempty"`
	Updated        *string      `yaml:"updated" json:"updated,omitempty"`
	Scope          *string      `yaml:"scope" json:"scope,omitempty"`
	Version        *string      `yaml:"version" json:"version,omitempty"`
	Stylesheets    []StyleSheet `yaml:"stylesheets" json:"stylesheets,omitempty"`
	Layers         []struct {
		ID           string  `yaml:"id" json:"id"`
		GeometryType *string `yaml:"type" json:"geometryType,omitempty"`
		SampleData   Link    `yaml:"sampleData" json:"sampleData,omitempty"`
		// TODO: the Properties schema is a stub and can be an implementation of: https://raw.githubusercontent.com/OAI/OpenAPI-Specification/master/schemas/v3.0/schema.json#/definitions/Schema
		PropertiesSchema *PropertiesSchema `yaml:"propertiesSchema" json:"propertiesSchema,omitempty"`
	} `yaml:"layers" json:"layers,omitempty"`
	Links []Link `yaml:"links" json:"links,omitempty"`
}

// StyleSheet based on OGC API Styles Requirement 7B
type StyleSheet struct {
	Title         *string `yaml:"title" json:"title,omitempty"`
	Version       *string `yaml:"version" json:"version,omitempty"`
	Specification *string `yaml:"specification" json:"specification,omitempty"`
	Native        *bool   `yaml:"native" json:"native,omitempty"`
	Link          Link    `yaml:"link" json:"link"`
}

// Link based on OGC API Features - http://schemas.opengis.net/ogcapi/features/part1/1.0/openapi/schemas/link.yaml - as referenced by OGC API Styles Requirements 3B and 7B
type Link struct {
	AssetFilename *string `yaml:"assetFilename" json:"-"`
	Href          *string `yaml:"href" json:"href"`
	Rel           string  `yaml:"rel" json:"rel,omitempty"` // This is allowed to be empty according to the spec, but we leverage this
	Type          *string `yaml:"type" json:"type,omitempty"`
	Format        *string `yaml:"format"`
	Title         *string `yaml:"title" json:"title,omitempty"`
	Hreflang      *string `yaml:"hreflang" json:"hreflang,omitempty"`
	Length        *int    `yaml:"length" json:"length,omitempty"`
}

type PropertiesSchema struct {
	// placeholder
}

type YAMLURL struct {
	*url.URL
}

// UnmarshalYAML parses a string to URL and also removes trailing slash if present,
// so we can easily append a longer path without having to worry about double slashes
func (o *YAMLURL) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	parsedURL, err := url.ParseRequestURI(strings.TrimSuffix(s, "/"))
	o.URL = parsedURL
	return err
}
