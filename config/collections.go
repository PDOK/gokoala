package config

import (
	"log"
	"sort"

	"dario.cat/mergo"
	"github.com/PDOK/gokoala/internal/engine/types"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// GeoSpatialCollections All collections configured for this OGC API. Can contain a mix of tiles/features/etc.
// +kubebuilder:object:generate:false
type GeoSpatialCollections []GeoSpatialCollection

// GeoSpatialCollection Configuration for a collection of geospatial data.
//
// Interface/abstraction for common collection properties regardless of the specific
// type (e.g., tiles, features, 3dgeovolumes, etc.).
//
// +kubebuilder:object:generate:false
type GeoSpatialCollection interface {

	// GetID Unique ID of the collection
	GetID() string

	// GetMetadata Metadata describing the collection contents
	GetMetadata() *GeoSpatialCollectionMetadata

	// GetLinks Links pertaining to this collection (e.g., downloads, documentation)
	GetLinks() *CollectionLinks

	// HasDateTime true when collection has temporal support, false otherwise.
	HasDateTime() bool

	// HasTableName true when collection uses the given table, false otherwise.
	HasTableName(table string) bool

	// Merge the (metadata and links) of the given collection with this collection. Return the merged collection.
	Merge(collection GeoSpatialCollection) GeoSpatialCollection
}

// +kubebuilder:object:generate=true
type GeoSpatialCollectionMetadata struct {
	// Human-friendly title of this collection. When no title is specified the collection ID is used.
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
type CollectionLinks struct {
	// Links to downloads of an entire collection. These will be rendered as rel=enclosure links
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

// HasCollections does this API offer collections, for example, with features, tiles, 3d tiles, etc.
func (c *Config) HasCollections() bool {
	return c.AllCollections() != nil
}

// AllCollections get all collections - for example, with features, tiles, 3d tiles - offered through this OGC API.
// Results are returned in alphabetic or literal order.
func (c *Config) AllCollections() GeoSpatialCollections {
	var result []GeoSpatialCollection
	if c.OgcAPI.GeoVolumes != nil {
		geoVolumes := types.ToInterfaceSlice[Collection3dGeoVolumes, GeoSpatialCollection](c.OgcAPI.GeoVolumes.Collections)
		result = append(result, geoVolumes...)
	}
	if c.OgcAPI.Tiles != nil {
		tiles := types.ToInterfaceSlice[CollectionTiles, GeoSpatialCollection](c.OgcAPI.Tiles.Collections)
		result = append(result, tiles...)
	}
	if c.OgcAPI.Features != nil {
		features := types.ToInterfaceSlice[CollectionFeatures, GeoSpatialCollection](c.OgcAPI.Features.Collections)
		result = append(result, features...)
	}
	if c.OgcAPI.FeaturesSearch != nil {
		featuresSearch := types.ToInterfaceSlice[CollectionFeaturesSearch, GeoSpatialCollection](c.OgcAPI.FeaturesSearch.Collections)
		result = append(result, featuresSearch...)
	}

	// sort
	if len(c.OgcAPICollectionOrder) > 0 {
		sortByLiteralOrder(result, c.OgcAPICollectionOrder)
	} else {
		sortByAlphabet(result)
	}

	return result
}

// Unique lists all unique GeoSpatialCollections (no duplicate IDs).
// Don't use in the hot path (creates a map on every invocation).
func (g GeoSpatialCollections) Unique() []GeoSpatialCollection {
	collectionsByID := g.toMap()
	result := make([]GeoSpatialCollection, 0, collectionsByID.Len())
	for pair := collectionsByID.Oldest(); pair != nil; pair = pair.Next() {
		result = append(result, pair.Value)
	}

	return result
}

// ContainsID check if a given collection - by ID - exists.
// Don't use in the hot path (creates a map on every invocation).
func (g GeoSpatialCollections) ContainsID(id string) bool {
	collectionsByID := g.toMap()
	_, ok := collectionsByID.Get(id)

	return ok
}

func (g GeoSpatialCollections) toMap() orderedmap.OrderedMap[string, GeoSpatialCollection] {
	collectionsByID := orderedmap.New[string, GeoSpatialCollection]()
	for _, current := range g {
		existing, ok := collectionsByID.Get(current.GetID())
		if ok {
			existing = existing.Merge(current)
			collectionsByID.Set(current.GetID(), existing)
		} else {
			collectionsByID.Set(current.GetID(), current)
		}
	}

	return *collectionsByID
}

func sortByAlphabet(collection []GeoSpatialCollection) {
	sort.Slice(collection, func(i, j int) bool {
		iName := collection[i].GetID()
		jName := collection[j].GetID()
		// prefer to sort by title when available, collection ID otherwise
		if collection[i].GetMetadata() != nil && collection[i].GetMetadata().Title != nil {
			iName = *collection[i].GetMetadata().Title
		}
		if collection[j].GetMetadata() != nil && collection[j].GetMetadata().Title != nil {
			jName = *collection[j].GetMetadata().Title
		}

		return iName < jName
	})
}

func sortByLiteralOrder(collections []GeoSpatialCollection, literalOrder []string) {
	collectionOrderIndex := make(map[string]int)
	for i, id := range literalOrder {
		collectionOrderIndex[id] = i
	}
	sort.Slice(collections, func(i, j int) bool {
		// sort, according to the explicit/literal order specified in OgcAPICollectionOrder
		return collectionOrderIndex[collections[i].GetID()] < collectionOrderIndex[collections[j].GetID()]
	})
}

func mergeMetadata(this GeoSpatialCollection, other GeoSpatialCollection) *GeoSpatialCollectionMetadata {
	return mergeField(this.GetID(), this.GetMetadata(), other.GetMetadata(), false)
}

func mergeLinks(this GeoSpatialCollection, other GeoSpatialCollection) *CollectionLinks {
	return mergeField(this.GetID(), this.GetLinks(), other.GetLinks(), true)
}

func mergeField[T any](id string, this *T, other *T, shouldAppend bool) *T {
	switch {
	case this == nil && other == nil:
		return nil
	case this == nil:
		return other
	case other == nil:
		return this
	}

	existing := *this
	var err error
	if shouldAppend {
		err = mergo.Merge(&existing, other, mergo.WithAppendSlice)
	} else {
		err = mergo.Merge(&existing, other)
	}
	if err != nil {
		log.Fatalf("failed to merge fields from 2 collections "+
			"with the same name '%s': %v", id, err)
		return nil
	}
	return &existing
}

func getGeoSpatialCollectionType(collection GeoSpatialCollection) string {
	switch collection.(type) {
	case Collection3dGeoVolumes:
		return "3dgeovolumes"
	case CollectionFeatures:
		return "features"
	case CollectionFeaturesSearch:
		return "featuressearch"
	case CollectionTiles:
		return "tiles"
	}
	log.Println("unknown collection type")
	return ""
}
