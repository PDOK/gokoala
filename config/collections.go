package config

import (
	"encoding/json"
	"log"
	"sort"

	"dario.cat/mergo"
	"gopkg.in/yaml.v3"
)

type GeoSpatialCollections []GeoSpatialCollection

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
	// Keep this in sync with the GeoSpatialCollection struct!
	ID                           string                        `json:"id"`
	Metadata                     *GeoSpatialCollectionMetadata `json:"metadata,omitempty"`
	Links                        *CollectionLinks              `json:"links,omitempty"`
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
		Links:                       c.Links,
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
type CollectionLinks struct {
	// Links to downloads of entire collection. These will be rendered as rel=enclosure links
	// +optional
	Downloads []DownloadLink `yaml:"downloads,omitempty" json:"downloads,omitempty" validate:"dive"`

	// Links to documentation describing the collection. These will be rendered as rel=describedby links
	// <placeholder>
}

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
	for _, current := range g {
		existing, ok := collectionsByID[current.ID]
		if ok {
			err := mergo.Merge(&existing, current)
			if err != nil {
				log.Fatalf("failed to merge 2 collections with the same name '%s': %v", current.ID, err)
			}
			collectionsByID[current.ID] = existing
		} else {
			collectionsByID[current.ID] = current
		}
	}
	return collectionsByID
}
