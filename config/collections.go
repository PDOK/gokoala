package config

import (
	"encoding/json"
	"log"
	"sort"

	"dario.cat/mergo"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

// GeoSpatialCollections All collections configured for this OGC API. Can contain a mix of tiles/features/etc.
type GeoSpatialCollections []GeoSpatialCollection

// +kubebuilder:object:generate=true
type GeoSpatialCollection struct {
	// Unique ID of the collection
	// +kubebuilder:validation:Pattern=`^[a-z0-9"]([a-z0-9_-]*[a-z0-9"]+|)$`
	ID string `yaml:"id" validate:"required,lowercase_id" json:"id"`

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

	// Features search (geocoding) specific to this collection
	// +optional
	FeaturesSearch *CollectionEntryFeaturesSearch `yaml:",inline" json:",inline"`
}

type GeoSpatialCollectionJSON struct {
	// Keep this in sync with the GeoSpatialCollection struct!
	ID                             string                        `json:"id"`
	Metadata                       *GeoSpatialCollectionMetadata `json:"metadata,omitempty"`
	Links                          *CollectionLinks              `json:"links,omitempty"`
	*CollectionEntry3dGeoVolumes   `json:",inline"`
	*CollectionEntryTiles          `json:",inline"`
	*CollectionEntryFeatures       `json:",inline"`
	*CollectionEntryFeaturesSearch `json:",inline"`
}

// MarshalJSON custom because inlining only works on embedded structs.
// Value instead of pointer receiver because only that way it can be used for both.
func (c GeoSpatialCollection) MarshalJSON() ([]byte, error) {
	return json.Marshal(GeoSpatialCollectionJSON{
		ID:                            c.ID,
		Metadata:                      c.Metadata,
		Links:                         c.Links,
		CollectionEntry3dGeoVolumes:   c.GeoVolumes,
		CollectionEntryTiles:          c.Tiles,
		CollectionEntryFeatures:       c.Features,
		CollectionEntryFeaturesSearch: c.FeaturesSearch,
	})
}

// UnmarshalJSON parses a string to GeoSpatialCollection.
func (c *GeoSpatialCollection) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, c)
}

// HasDateTime true when collection has temporal support, false otherwise.
func (c *GeoSpatialCollection) HasDateTime() bool {
	return c.Metadata != nil && c.Metadata.TemporalProperties != nil
}

// HasTableName true when collection uses the given table, false otherwise.
func (c *GeoSpatialCollection) HasTableName(table string) bool {
	return c.Features != nil && c.Features.TableName != nil &&
		table == *c.Features.TableName
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

// HasCollections does this API offer collections with for example features, tiles, 3d tiles, etc.
func (c *Config) HasCollections() bool {
	return c.AllCollections() != nil
}

// AllCollections get all collections - with  for example features, tiles, 3d tiles - offered through this OGC API.
// Results are returned in alphabetic or literal order.
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
	if c.OgcAPI.FeaturesSearch != nil {
		result = append(result, c.OgcAPI.FeaturesSearch.Collections...)
	}

	// sort
	if len(c.OgcAPICollectionOrder) > 0 {
		sortByLiteralOrder(result, c.OgcAPICollectionOrder)
	} else {
		sortByAlphabet(result)
	}

	return result
}

// FeaturePropertiesByID returns a map of collection IDs to their corresponding FeatureProperties.
// Skips collections that do not have features defined.
func (g GeoSpatialCollections) FeaturePropertiesByID() map[string]*FeatureProperties {
	result := make(map[string]*FeatureProperties)
	for _, collection := range g {
		if collection.Features == nil {
			continue
		}
		result[collection.ID] = collection.Features.FeatureProperties
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

// ContainsID check if given collection - by ID - exists.
// Don't use in the hot path (creates a map on every invocation).
func (g GeoSpatialCollections) ContainsID(id string) bool {
	collectionsByID := g.toMap()
	_, ok := collectionsByID.Get(id)

	return ok
}

func (g GeoSpatialCollections) toMap() orderedmap.OrderedMap[string, GeoSpatialCollection] {
	collectionsByID := orderedmap.New[string, GeoSpatialCollection]()
	for _, current := range g {
		existing, ok := collectionsByID.Get(current.ID)
		if ok {
			err := mergo.Merge(&existing, current)
			if err != nil {
				log.Fatalf("failed to merge 2 collections with the same name '%s': %v", current.ID, err)
			}
			collectionsByID.Set(current.ID, existing)
		} else {
			collectionsByID.Set(current.ID, current)
		}
	}

	return *collectionsByID
}

func sortByAlphabet(collection []GeoSpatialCollection) {
	sort.Slice(collection, func(i, j int) bool {
		iName := collection[i].ID
		jName := collection[j].ID
		// prefer to sort by title when available, collection ID otherwise
		if collection[i].Metadata != nil && collection[i].Metadata.Title != nil {
			iName = *collection[i].Metadata.Title
		}
		if collection[j].Metadata != nil && collection[j].Metadata.Title != nil {
			jName = *collection[j].Metadata.Title
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
		// sort according to the explicit/literal order specified in OgcAPICollectionOrder
		return collectionOrderIndex[collections[i].ID] < collectionOrderIndex[collections[j].ID]
	})
}
