package config

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"github.com/PDOK/gokoala/internal/engine/util"
	"gopkg.in/yaml.v3"
)

var DefaultSrs = "EPSG:28992"

// +kubebuilder:object:generate=true
type OgcAPITiles struct {
	// Tiles for the entire dataset, these are hosted at the root of the API (/tiles endpoint).
	// +optional
	DatasetTiles *Tiles `yaml:",inline" json:",inline"`

	// Tiles per collection. When no collections are specified tiles should be hosted at the root of the API (/tiles endpoint).
	// +optional
	Collections CollectionsTiles `yaml:"collections,omitempty" json:"collections,omitempty"`
}

type CollectionsTiles []CollectionTiles

// ContainsID check if a given collection - by ID - exists.
func (cst CollectionsTiles) ContainsID(id string) bool {
	for _, coll := range cst {
		if coll.ID == id {
			return true
		}
	}
	return false
}

type OgcAPITilesJSON struct {
	*Tiles      `json:",inline"`
	Collections []CollectionTiles `json:"collections,omitempty"`
}

// MarshalJSON custom because inlining only works on embedded structs.
// Value instead of pointer receiver because only that way it can be used for both.
func (o OgcAPITiles) MarshalJSON() ([]byte, error) {
	return json.Marshal(OgcAPITilesJSON{
		Tiles:       o.DatasetTiles,
		Collections: o.Collections,
	})
}

// UnmarshalJSON parses a string to OgcAPITiles.
func (o *OgcAPITiles) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, o)
}

func (o *OgcAPITiles) Defaults() {
	if o.DatasetTiles != nil && o.DatasetTiles.HealthCheck.Srs == DefaultSrs &&
		o.DatasetTiles.HealthCheck.TilePath == nil && *o.DatasetTiles.HealthCheck.Enabled {
		o.DatasetTiles.deriveHealthCheckTilePath()
	} else if o.Collections != nil {
		for i := range o.Collections {
			if o.Collections[i].GeoDataTiles.HealthCheck.Srs == DefaultSrs &&
				o.Collections[i].GeoDataTiles.HealthCheck.TilePath == nil &&
				*o.Collections[i].GeoDataTiles.HealthCheck.Enabled {
				o.Collections[i].GeoDataTiles.deriveHealthCheckTilePath()
			}
		}
	}
}

// +kubebuilder:object:generate=true
//
//nolint:recvcheck
type CollectionTiles struct {
	// Unique ID of the collection
	// +kubebuilder:validation:Pattern=`^[a-z0-9"]([a-z0-9_-]*[a-z0-9"]+|)$`
	ID string `yaml:"id" validate:"required,lowercase_id" json:"id"`

	// Metadata describing the collection contents
	// +optional
	Metadata *GeoSpatialCollectionMetadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// Links pertaining to this collection (e.g., downloads, documentation)
	// +optional
	Links *CollectionLinks `yaml:"links,omitempty" json:"links,omitempty"`

	// Tiles specific to this collection. Called 'geodata tiles' in OGC spec.
	GeoDataTiles Tiles `yaml:",inline" json:",inline" validate:"required"`
}

type CollectionEntryTilesJSON struct {
	Tiles `json:",inline"`
}

// MarshalJSON custom because inlining only works on embedded structs.
// Value instead of pointer receiver because only that way it can be used for both.
func (ct CollectionTiles) MarshalJSON() ([]byte, error) {
	return json.Marshal(CollectionEntryTilesJSON{
		Tiles: ct.GeoDataTiles,
	})
}

// UnmarshalJSON parses a string to CollectionTiles.
func (ct CollectionTiles) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, ct)
}

func (ct CollectionTiles) GetID() string {
	return ct.ID
}

func (ct CollectionTiles) GetMetadata() *GeoSpatialCollectionMetadata {
	return ct.Metadata
}

func (ct CollectionTiles) GetLinks() *CollectionLinks {
	return ct.Links
}

func (ct CollectionTiles) Merge(other GeoSpatialCollection) GeoSpatialCollection {
	ct.Metadata = mergeMetadata(ct, other)
	ct.Links = mergeLinks(ct, other)
	return ct
}

// +kubebuilder:validation:Enum=raster;vector
type TilesType string

const (
	TilesTypeRaster TilesType = "raster"
	TilesTypeVector TilesType = "vector"
)

func (o *OgcAPITiles) HasType(t TilesType) bool {
	if o.DatasetTiles != nil && slices.Contains(o.DatasetTiles.Types, t) {
		return true
	}
	for _, coll := range o.Collections {
		if slices.Contains(coll.GeoDataTiles.Types, t) {
			return true
		}
	}

	return false
}

// AllTileProjections projections supported by GoKoala for serving (vector) tiles, regardless of the dataset.
// When adding a new projection also add corresponding HTML/JSON templates.
var AllTileProjections = map[string]string{
	"EPSG:28992": "NetherlandsRDNewQuad",
	"EPSG:3035":  "EuropeanETRS89_LAEAQuad",
	"EPSG:3857":  "WebMercatorQuad",
}

// HasProjection true when the given projection is supported for this dataset.
func (o *OgcAPITiles) HasProjection(srs string) bool {
	for _, projection := range o.GetProjections() {
		if projection.Srs == srs {
			return true
		}
	}

	return false
}

// GetProjections projections supported for this dataset.
func (o *OgcAPITiles) GetProjections() []SupportedSrs {
	supportedSrsSet := map[SupportedSrs]struct{}{}
	if o.DatasetTiles != nil {
		for _, supportedSrs := range o.DatasetTiles.SupportedSrs {
			supportedSrsSet[supportedSrs] = struct{}{}
		}
	}
	for _, coll := range o.Collections {
		for _, supportedSrs := range coll.GeoDataTiles.SupportedSrs {
			supportedSrsSet[supportedSrs] = struct{}{}
		}
	}
	result := util.Keys(supportedSrsSet)
	sort.Slice(result, func(i, j int) bool {
		return len(result[i].Srs) > len(result[j].Srs)
	})

	return result
}

// +kubebuilder:object:generate=true
type Tiles struct {
	// Reference to the server (or object storage) hosting the tiles.
	// Note: Only marked as optional in CRD to support top-level OR collection-level tiles
	// +optional
	TileServer URL `yaml:"tileServer" json:"tileServer" validate:"required"`

	// Could be 'vector' and/or 'raster' to indicate the types of tiles offered
	// Note: Only marked as optional in CRD to support top-level OR collection-level tiles
	// +optional
	Types []TilesType `yaml:"types" json:"types" validate:"required"`

	// Specifies in what projections (SRS/CRS) the tiles are offered
	// Note: Only marked as optional in CRD to support top-level OR collection-level tiles
	// +optional
	SupportedSrs []SupportedSrs `yaml:"supportedSrs" json:"supportedSrs" validate:"required,dive"`

	// Optional template to the vector tiles on the tileserver. Defaults to {tms}/{z}/{x}/{y}.pbf.
	// +optional
	URITemplateTiles *string `yaml:"uriTemplateTiles,omitempty" json:"uriTemplateTiles,omitempty"`

	// Optional health check configuration
	// +optional
	HealthCheck HealthCheck `yaml:"healthCheck" json:"healthCheck"`
}

func (t *Tiles) deriveHealthCheckTilePath() {
	var deepestZoomLevel int
	for _, srs := range t.SupportedSrs {
		if srs.Srs == DefaultSrs {
			deepestZoomLevel = srs.ZoomLevelRange.End
		}
	}
	defaultTile := HealthCheckDefaultTiles[deepestZoomLevel]
	tileMatrixSet := AllTileProjections[DefaultSrs]
	tilePath := fmt.Sprintf("/%s/%d/%d/%d.pbf", tileMatrixSet, deepestZoomLevel, defaultTile.x, defaultTile.y)
	t.HealthCheck.TilePath = &tilePath
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

type TileCoordinates struct {
	x int
	y int
}

// default tiles for EPSG:28992 - location centered just outside a village in the province of Friesland.
var HealthCheckDefaultTiles = map[int]TileCoordinates{
	0:  {x: 0, y: 0},
	1:  {x: 1, y: 0},
	2:  {x: 2, y: 1},
	3:  {x: 4, y: 2},
	4:  {x: 8, y: 5},
	5:  {x: 17, y: 11},
	6:  {x: 35, y: 22},
	7:  {x: 71, y: 45},
	8:  {x: 143, y: 91},
	9:  {x: 286, y: 182},
	10: {x: 572, y: 365},
	11: {x: 1144, y: 731},
	12: {x: 2288, y: 1462},
	13: {x: 4576, y: 2925},
	14: {x: 9152, y: 5851},
	15: {x: 18304, y: 11702},
	16: {x: 36608, y: 23404},
}

// +kubebuilder:object:generate=true
type HealthCheck struct {
	// Enable/disable healthcheck on tiles. Defaults to true.
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `yaml:"enabled" json:"enabled" default:"true"`

	// Projection (SRS/CRS) used for tile healthcheck
	// +kubebuilder:default="EPSG:28992"
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	// +optional
	Srs string `yaml:"srs" json:"srs" default:"EPSG:28992" validate:"required,startswith=EPSG:"`

	// Path to specific tile used for healthcheck
	// +optional
	TilePath *string `yaml:"tilePath,omitempty" json:"tilePath,omitempty" validate:"required_unless=Srs EPSG:28992"`
}

func validateTileProjections(tiles *OgcAPITiles) error {
	var errMessages []string
	if tiles.DatasetTiles != nil {
		for _, srs := range tiles.DatasetTiles.SupportedSrs {
			if _, ok := AllTileProjections[srs.Srs]; !ok {
				errMessages = append(errMessages, fmt.Sprintf("validation failed for srs '%s'; srs is not supported", srs.Srs))
			}
		}
	}
	for _, collection := range tiles.Collections {
		for _, srs := range collection.GeoDataTiles.SupportedSrs {
			if _, ok := AllTileProjections[srs.Srs]; !ok {
				errMessages = append(errMessages, fmt.Sprintf("validation failed for srs '%s'; srs is not supported", srs.Srs))
			}
		}
	}
	if len(errMessages) > 0 {
		return fmt.Errorf("invalid config provided:\n%v", errMessages)
	}

	return nil
}
