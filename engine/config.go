package engine

import (
	"log"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func ReadConfigFile(configFile string) *Config {
	yamlData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config file %v", err)
	}

	// expand environment variables
	yamlData = []byte(os.ExpandEnv(string(yamlData)))

	var result *Config
	err = yaml.Unmarshal(yamlData, &result)
	if err != nil {
		log.Fatalf("failed to unmarshal config file %v", err)
	}
	return result
}

type Config struct {
	Title          string          `yaml:"title"`
	Abstract       string          `yaml:"abstract"`
	Thumbnail      *string         `yaml:"thumbnail"`
	Keywords       []string        `yaml:"keywords"`
	LastUpdated    *string         `yaml:"lastUpdated"`
	License        string          `yaml:"license"`
	Support        *string         `yaml:"support"`
	DatasetDetails []DatasetDetail `yaml:"datasetDetails"`
	BaseURL        YAMLURL         `yaml:"baseUrl"`
	OgcAPI         OgcAPI          `yaml:"ogcApi"`
	ResourcesDir   string
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

type DatasetDetail struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type OgcAPI struct {
	GeoVolumes *OgcAPI3dGeoVolumes `yaml:"3dgeovolumes"`
	Tiles      *OgcAPITiles        `yaml:"tiles"`
	Styles     *OgcAPIStyles       `yaml:"styles"`
	Features   *OgcAPIFeatures     `yaml:"features"`
	Maps       *OgcAPIMaps         `yaml:"maps"`
}

type GeoSpatialCollections []GeoSpatialCollection

// Unique lists all unique GeoSpatialCollections (no duplicate IDs)
func (g GeoSpatialCollections) Unique() []GeoSpatialCollection {
	collectionsByID := g.toMap()
	flattened := make([]GeoSpatialCollection, 0, len(collectionsByID))
	for _, v := range collectionsByID {
		flattened = append(flattened, v)
	}
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
	ID       string                        `yaml:"id"`
	Metadata *GeoSpatialCollectionMetadata `yaml:"metadata"`

	GeoVolumes *CollectionEntry3dGeoVolumes `yaml:",inline"`
	Tiles      *CollectionEntryTiles        `yaml:",inline"`
	Styles     *CollectionEntryStyles       `yaml:",inline"`
	Features   *CollectionEntryFeatures     `yaml:",inline"`
	Maps       *CollectionEntryMaps         `yaml:",inline"`
}

type GeoSpatialCollectionMetadata struct {
	Description *string  `yaml:"description"`
	Thumbnail   *string  `yaml:"thumbnail"`
	Keywords    []string `yaml:"keywords"`
	LastUpdated *string  `yaml:"lastUpdated"`
	Extent      *Extent  `yaml:"extent"`
}

type CollectionEntry3dGeoVolumes struct {
	// Optional basepath to 3D tiles on the tileserver. Defaults to the collection ID.
	TileServerPath *string `yaml:"tileServerPath"`

	// Optional URI template for individual 3D tiles, defaults to "tiles/{level}/{x}/{y}.glb".
	URITemplate3dTiles *string `yaml:"uriTemplate3dTiles"`

	// Optional URI template for subtrees, only required when "implicit tiling" extension is used.
	URITemplateImplicitTilingSubtree *string `yaml:"uriTemplateImplicitTilingSubtree"`

	// Optional URL to 3D viewer to visualize the given collection of 3D Tiles.
	URL3DViewer *YAMLURL `yaml:"3dViewerUrl"`
}

type CollectionEntryTiles struct {
	// placeholder
}

type CollectionEntryStyles struct {
	// placeholder
}

type CollectionEntryFeatures struct {
	// placeholder
}

type CollectionEntryMaps struct {
	// placeholder
}

type OgcAPI3dGeoVolumes struct {
	TileServer  YAMLURL               `yaml:"tileServer"`
	Thumbnail   string                `yaml:"thumbnail"`
	Collections GeoSpatialCollections `yaml:"collections"`
}

type OgcAPITiles struct {
	Title        string                `yaml:"title"`
	Abstract     string                `yaml:"abstract"`
	TileServer   YAMLURL               `yaml:"tileServer"`
	Types        []string              `yaml:"types"`
	SupportedSrs []SupportedSrs        `yaml:"supportedSrs"`
	Collections  GeoSpatialCollections `yaml:"collections"`
}

type OgcAPIStyles struct {
	BaseURL         YAMLURL
	Title           string          `yaml:"title"`
	Abstract        string          `yaml:"abstract"`
	Default         string          `yaml:"default,omitempty"`
	SupportedStyles []StyleMetadata `yaml:"supportedStyles"`
}

type OgcAPIFeatures struct {
	Collections GeoSpatialCollections `yaml:"collections"`
}

type OgcAPIMaps struct {
	Collections GeoSpatialCollections `yaml:"collections"`
}

type SupportedSrs struct {
	Srs            SRS            `yaml:"srs"`
	ZoomLevelRange ZoomLevelRange `yaml:"zoomLevelRange"`
}

type ZoomLevelRange struct {
	Start int `yaml:"start"`
	End   int `yaml:"end"`
}

type YAMLURL struct {
	*url.URL
}

type SRS struct {
	EPSG string
}

type Extent struct {
	Srs  SRS      `yaml:"srs"`
	Bbox []string `yaml:"bbox"`
}

// StyleMetadata based on OGC API Styles Requirement 7B
type StyleMetadata struct {
	ID             string       `yaml:"id" json:"id"`
	Title          *string      `yaml:"title" json:"title,omitempty"`
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

// UnmarshalYAML parses a string to EPSG srs/crs
func (o *SRS) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(s, "EPSG:") {
		log.Fatalf("failed to parse SRS, should contain EPSG code and start with 'EPSG:'")
	}
	o.EPSG = s
	return err
}
