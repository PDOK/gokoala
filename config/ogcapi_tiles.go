package config

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"github.com/PDOK/gokoala/internal/engine/util"
	"gopkg.in/yaml.v3"
)

// +kubebuilder:object:generate=true
type OgcAPITiles struct {
	// Tiles for the entire dataset, these are hosted at the root of the API (/tiles endpoint).
	// +optional
	DatasetTiles *Tiles `yaml:",inline" json:",inline"`

	// Tiles per collection. When no collections are specified tiles should be hosted at the root of the API (/tiles endpoint).
	// +optional
	Collections GeoSpatialCollections `yaml:"collections,omitempty" json:"collections,omitempty"`
}

type OgcAPITilesJSON struct {
	*Tiles      `json:",inline"`
	Collections GeoSpatialCollections `json:"collections,omitempty"`
}

// MarshalJSON custom because inlining only works on embedded structs.
// Value instead of pointer receiver because only that way it can be used for both.
func (o OgcAPITiles) MarshalJSON() ([]byte, error) {
	return json.Marshal(OgcAPITilesJSON{
		Tiles:       o.DatasetTiles,
		Collections: o.Collections,
	})
}

// UnmarshalJSON parses a string to OgcAPITiles
func (o *OgcAPITiles) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, o)
}

func (o *OgcAPITiles) Defaults() {
	if o.DatasetTiles != nil && o.DatasetTiles.HealthCheck.Srs == DefaultSrs &&
		o.DatasetTiles.HealthCheck.TilePath == nil {
		o.DatasetTiles.deriveHealthCheckTilePath()
	} else if o.Collections != nil {
		for _, coll := range o.Collections {
			if coll.Tiles != nil && coll.Tiles.GeoDataTiles.HealthCheck.Srs == DefaultSrs && coll.Tiles.GeoDataTiles.HealthCheck.TilePath == nil {
				coll.Tiles.GeoDataTiles.deriveHealthCheckTilePath()
			}
		}
	}
}

// +kubebuilder:object:generate=true
type CollectionEntryTiles struct {

	// Tiles specific to this collection. Called 'geodata tiles' in OGC spec.
	GeoDataTiles Tiles `yaml:",inline" json:",inline" validate:"required"`
}

type CollectionEntryTilesJSON struct {
	Tiles `json:",inline"`
}

// MarshalJSON custom because inlining only works on embedded structs.
// Value instead of pointer receiver because only that way it can be used for both.
func (c CollectionEntryTiles) MarshalJSON() ([]byte, error) {
	return json.Marshal(CollectionEntryTilesJSON{
		Tiles: c.GeoDataTiles,
	})
}

// UnmarshalJSON parses a string to OgcAPITiles
func (c *CollectionEntryTiles) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, c)
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
		if coll.Tiles != nil && slices.Contains(coll.Tiles.GeoDataTiles.Types, t) {
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

// HasProjection true when the given projection is supported for this dataset
func (o *OgcAPITiles) HasProjection(srs string) bool {
	for _, projection := range o.GetProjections() {
		if projection.Srs == srs {
			return true
		}
	}
	return false
}

// GetProjections projections supported for this dataset
func (o *OgcAPITiles) GetProjections() []SupportedSrs {
	supportedSrsSet := map[SupportedSrs]struct{}{}
	if o.DatasetTiles != nil {
		for _, supportedSrs := range o.DatasetTiles.SupportedSrs {
			supportedSrsSet[supportedSrs] = struct{}{}
		}
	}
	for _, coll := range o.Collections {
		if coll.Tiles == nil {
			continue
		}
		for _, supportedSrs := range coll.Tiles.GeoDataTiles.SupportedSrs {
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

// default tiles for EPSG:28992 - location centered just outside a village in the province of Friesland
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
	// Projection (SRS/CRS) used for tile healthcheck
	// +kubebuilder:default="EPSG:28992"
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	// +optional
	Srs string `yaml:"srs" json:"srs" default:"EPSG:28992" validate:"required,startswith=EPSG:"`

	// Path to specific tile used for healthcheck
	// +optional
	TilePath *string `yaml:"tilePath,omitempty" json:"tilePath,omitempty" validate:"required_unless=Srs EPSG:28992"`
}
