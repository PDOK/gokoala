package config

import (
	"log"
	"sort"

	"dario.cat/mergo"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type GeoSpatialCollections []GeoSpatialCollection

type GeoSpatialCollection struct {
	// Unique ID of the collection
	ID string `yaml:"id" json:"id" validate:"required"`

	// Metadata describing the collection contents
	Metadata *GeoSpatialCollectionMetadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// Links pertaining to this collection (e.g., downloads, documentation)
	Links *CollectionLinks `yaml:"links,omitempty" json:"links,omitempty"`

	// Search config related to location search/suggest
	Search *Search `yaml:"search,omitempty" json:"search,omitempty"`
}

type GeoSpatialCollectionMetadata struct {
	// Human friendly title of this collection. When no title is specified the collection ID is used.
	Title *string `yaml:"title,omitempty" json:"title,omitempty"`

	// Describes the content of this collection
	Description *string `yaml:"description" json:"description" validate:"required"`

	// Reference to a PNG image to use a thumbnail on the collections.
	// The full path is constructed by appending Resources + Thumbnail.
	// +optional
	Thumbnail *string `yaml:"thumbnail,omitempty" json:"thumbnail,omitempty"`

	// Keywords to make this collection beter discoverable
	Keywords []string `yaml:"keywords,omitempty" json:"keywords,omitempty"`

	// Moment in time when the collection was last updated
	LastUpdated *string `yaml:"lastUpdated,omitempty" json:"lastUpdated,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z"`

	// Who updated this collection
	LastUpdatedBy string `yaml:"lastUpdatedBy,omitempty" json:"lastUpdatedBy,omitempty"`

	// Extent of the collection, both geospatial and/or temporal
	Extent *Extent `yaml:"extent,omitempty" json:"extent,omitempty"`

	// The CRS identifier which the features are originally stored, meaning no CRS transformations are applied when features are retrieved in this CRS.
	// WGS84 is the default storage CRS.
	StorageCrs *string `yaml:"storageCrs,omitempty" json:"storageCrs,omitempty" default:"http://www.opengis.net/def/crs/OGC/1.3/CRS84" validate:"startswith=http://www.opengis.net/def/crs"`
}

type Extent struct {
	// Projection (SRS/CRS) to be used. When none is provided WGS84 (http://www.opengis.net/def/crs/OGC/1.3/CRS84) is used.
	Srs string `yaml:"srs,omitempty" json:"srs,omitempty" validate:"omitempty,startswith=EPSG:"`

	// Geospatial extent
	Bbox []string `yaml:"bbox" json:"bbox"`

	// Temporal extent
	Interval []string `yaml:"interval,omitempty" json:"interval,omitempty" validate:"omitempty,len=2"`
}

type Search struct {
	// Fields that make up the display name and/or suggestions. These fields can be used as variables in the DisplayNameTemplate and SuggestTemplates.
	Fields []string `yaml:"fields,omitempty" json:"fields,omitempty" validate:"required"`

	// Template that indicates how a search record is displayed. Uses Go text/template syntax to reference fields.
	DisplayNameTemplate string `yaml:"displayNameTemplate,omitempty" json:"displayNameTemplate,omitempty" validate:"required"`

	// Version of the collection used to link to search results
	Version int `yaml:"version,omitempty" json:"version,omitempty" default:"1"`

	// (Links to) the individual OGC API (feature) collections that are searchable in this collection.
	// +kubebuilder:validation:MinItems=1
	OGCCollections []RelatedOGCAPIFeaturesCollection `yaml:"ogcCollections" json:"ogcCollections" validate:"required,min=1"`

	ETL SearchETL `yaml:"etl" json:"etl" validate:"required"`
}

type SearchETL struct {
	// One or more templates that make up the autosuggestions. Uses Go text/template syntax to reference fields.
	SuggestTemplates []string `yaml:"suggestTemplates" json:"suggestTemplates" validate:"required,min=1"`

	// SQLite WHERE clause to filter features when importing/ETL-ing
	// (Without the WHERE keyword, only the clause)
	// +Optional
	Filter string `yaml:"filter,omitempty" json:"filter,omitempty"`
}

type RelatedOGCAPIFeaturesCollection struct {
	// Base URL/Href to the OGC Features API
	APIBaseURL URL `yaml:"api" json:"api" validate:"required"`

	// Geometry type of the features in the related collection.
	// A collections in an OGC Features API has a single geometry type.
	// But a searchable collection has no geometry type distinction and thus
	// could be assembled of multiple OGC Feature API collections (with the same feature type).
	GeometryType string `yaml:"geometryType" json:"geometryType" validate:"required"`

	// Collection ID in the OGC Features API
	CollectionID string `yaml:"collection" json:"collection" validate:"required"`

	// `datetime` query parameter for the OGC Features API. In case it's temporal.
	// E.g.: "{now()-1h}"
	// +optional
	Datetime *string `yaml:"datetime,omitempty" json:"datetime,omitempty"`
}

type CollectionLinks struct {
	// Links to downloads of entire collection. These will be rendered as rel=enclosure links
	// <placeholder>

	// Links to documentation describing the collection. These will be rendered as rel=describedby links
	// <placeholder>
}

// HasCollections does this API offer collections with for example features, tiles, 3d tiles, etc
func (c *Config) HasCollections() bool {
	return c.AllCollections() != nil
}

// AllCollections get all collections - with  for example features, tiles, 3d tiles - offered through this OGC API.
// Results are returned in alphabetic or literal order.
func (c *Config) AllCollections() GeoSpatialCollections {
	if len(c.CollectionOrder) > 0 {
		sortByLiteralOrder(c.Collections, c.CollectionOrder)
	} else {
		sortByAlphabet(c.Collections)
	}
	return c.Collections
}

// Unique lists all unique GeoSpatialCollections (no duplicate IDs).
// Don't use in hot path (creates a map on every invocation).
func (g GeoSpatialCollections) Unique() []GeoSpatialCollection {
	collectionsByID := g.toMap()
	result := make([]GeoSpatialCollection, 0, collectionsByID.Len())
	for pair := collectionsByID.Oldest(); pair != nil; pair = pair.Next() {
		result = append(result, pair.Value)
	}
	return result
}

// ContainsID check if given collection - by ID - exists.
// Don't use in hot path (creates a map on every invocation).
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
