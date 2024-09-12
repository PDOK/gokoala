package config

import (
	"slices"
	"sort"

	"github.com/PDOK/gokoala/internal/engine/util"
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

// +kubebuilder:object:generate=true
type CollectionEntryTiles struct {

	// Tiles specific to this collection. Called 'geodata tiles' in OGC spec.
	GeoDataTiles Tiles `yaml:",inline" json:",inline" validate:"required"`
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

func (o *OgcAPITiles) HasProjection(srs string) bool {
	for _, projection := range o.GetProjections() {
		if projection.Srs == srs {
			return true
		}
	}
	return false
}

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
