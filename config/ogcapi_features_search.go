package config

import (
	"net/url"
	"slices"

	"github.com/PDOK/gokoala/internal/engine/types"
	"github.com/PDOK/gokoala/internal/engine/util"
)

// +kubebuilder:object:generate=true
type OgcAPIFeaturesSearch struct {
	// Basemap to use in embedded viewer on the HTML pages.
	// +kubebuilder:default="OSM"
	// +kubebuilder:validation:Enum=OSM;BRT
	// +optional
	Basemap string `yaml:"basemap,omitempty" json:"basemap,omitempty" default:"OSM" validate:"oneof=OSM BRT"`

	// Collections available for search through this API
	Collections FeaturesSearchCollections `yaml:"collections" json:"collections" validate:"required,dive"`

	// One or more datasources to get the features from (geopackages, postgres, etc).
	// Optional since you can also define datasources at the collection level
	// +optional
	Datasources *Datasources `yaml:"datasources,omitempty" json:"datasources,omitempty"`

	// Whether GeoJSON/JSON-FG responses will be validated against the OpenAPI spec
	// since it has a significant performance impact when dealing with large JSON payloads.
	//
	// +kubebuilder:default=true
	// +optional
	ValidateResponses *bool `yaml:"validateResponses,omitempty" json:"validateResponses,omitempty" default:"true"` // ptr due to https://github.com/creasty/defaults/issues/49

	// Maximum number of decimals allowed in geometry coordinates. When not specified (default value of 0) no limit is enforced.
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxDecimals int `yaml:"maxDecimals,omitempty" json:"maxDecimals,omitempty" default:"0"`

	// Force timestamps in features to the UTC timezone.
	//
	// +kubebuilder:default=false
	// +optional
	ForceUTC bool `yaml:"forceUtc,omitempty" json:"forceUtc,omitempty"`

	// Settings related to the search API/index.
	// +optional
	SearchSettings SearchSettings `yaml:"searchSettings" json:"searchSettings"`
}

func (fs *OgcAPIFeaturesSearch) CollectionsSRS() []string {
	return fs.CollectionSRS("")
}

func (fs *OgcAPIFeaturesSearch) CollectionSRS(_ string) []string {
	uniqueSRSs := make(map[string]struct{})
	if fs.Datasources != nil {
		for _, d := range fs.Datasources.OnTheFly {
			for _, srs := range d.SupportedSrs {
				uniqueSRSs[srs.Srs] = struct{}{}
			}
		}
		for _, d := range fs.Datasources.Additional {
			uniqueSRSs[d.Srs] = struct{}{}
		}
	}
	result := util.Keys(uniqueSRSs)
	slices.Sort(result)

	return result
}

type FeaturesSearchCollections []FeaturesSearchCollection

// ContainsID check if a given collection - by ID - exists.
func (csfs FeaturesSearchCollections) ContainsID(id string) bool {
	for _, coll := range csfs {
		if coll.ID == id {
			return true
		}
	}
	return false
}

// +kubebuilder:object:generate=true
//
//nolint:recvcheck
type FeaturesSearchCollection struct {
	// Unique ID of the collection
	// +kubebuilder:validation:Pattern=`^[a-z0-9"]([a-z0-9_-]*[a-z0-9"]+|)$`
	ID string `yaml:"id" validate:"required,lowercase_id" json:"id"`

	// Metadata describing the collection contents
	// +optional
	Metadata *GeoSpatialCollectionMetadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// Links pertaining to this collection (e.g., downloads, documentation)
	// +optional
	Links *CollectionLinks `yaml:"links,omitempty" json:"links,omitempty"`

	// Fields that make up the display name and/or suggestions. These fields can be used as variables in the DisplayNameTemplate.
	Fields []string `yaml:"fields,omitempty" json:"fields,omitempty"`

	// Template that indicates how a search record is displayed. Uses Go text/template syntax to reference fields.
	DisplayNameTemplate string `yaml:"displayNameTemplate,omitempty" json:"displayNameTemplate,omitempty"`

	// Version of the collection exposed through the API.
	// +kubebuilder:default=1
	Version int `yaml:"version,omitempty" json:"version,omitempty" default:"1"`

	// Links to the individual OGC API (feature) collections that are searchable in this collection.
	CollectionRefs []RelatedOGCAPIFeaturesCollection `yaml:"collectionRefs,omitempty" json:"collectionRefs,omitempty"`
}

func (cfs FeaturesSearchCollection) GetID() string {
	return cfs.ID
}

func (cfs FeaturesSearchCollection) GetMetadata() *GeoSpatialCollectionMetadata {
	return cfs.Metadata
}

func (cfs FeaturesSearchCollection) GetLinks() *CollectionLinks {
	return cfs.Links
}

func (cfs FeaturesSearchCollection) Merge(other GeoSpatialCollection) GeoSpatialCollection {
	cfs.Metadata = mergeMetadata(cfs, other)
	cfs.Links = mergeLinks(cfs, other)
	return cfs
}

// IsRemoteFeatureCollection true when the given collection ID is defined as a feature collection outside this config.
// In other words: it references a remote feature collection and doesn't point to a local one in this dataset.
func (cfs FeaturesSearchCollection) IsRemoteFeatureCollection(collID string) bool {
	if len(cfs.CollectionRefs) == 1 {
		collRef := cfs.CollectionRefs[0]
		return collRef.CollectionID != collID || collRef.APIBaseURL.URL != nil
	}
	return true
}

// +kubebuilder:object:generate=true
type SearchSettings struct {
	// Name of the search index in the data store.
	// +kubebuilder:default=search_index
	IndexName string `yaml:"indexName" json:"indexName" default:"search_index" validate:"required"`

	// ADVANCED SETTING. Normalization specifies whether and how a document's length should impact its rank.
	// Possible values are 0, 1, 2, 4, 8, 16 and 32. For more information see https://www.postgresql.org/docs/current/textsearch-controls.html
	// +kubebuilder:default=1
	RankNormalization int `yaml:"rankNormalization,omitempty" json:"rankNormalization,omitempty" default:"1" validate:"gt=0"`

	// ADVANCED SETTING. Multiply the exact match rank to boost it above the wildcard matches.
	// +kubebuilder:validation:Pattern=`^-?\d+(\.\d+)?$`
	// +kubebuilder:default="3.0"
	ExactMatchMultiplier string `yaml:"exactMatchMultiplier,omitempty" json:"exactMatchMultiplier,omitempty" default:"3.0" validate:"numeric,gt=0"`

	// ADVANCED SETTING. The primary suggest is equal to the display name. With this multiplier you can boost it above other suggests.
	// +kubebuilder:validation:Pattern=`^-?\d+(\.\d+)?$`
	// +kubebuilder:default="1.01"
	PrimarySuggestMultiplier string `yaml:"primarySuggestMultiplier,omitempty" json:"primarySuggestMultiplier,omitempty" default:"1.01" validate:"numeric,gt=0"`

	// ADVANCED SETTING. The threshold above which results are pre-ranked instead ranked exactly.
	// +kubebuilder:default=40000
	RankThreshold int `yaml:"rankThreshold,omitempty" json:"rankThreshold,omitempty" default:"40000" validate:"gt=0"`

	// ADVANCED SETTING. The number of results which are pre-ranked when the rank threshold is hit.
	// +kubebuilder:default=10
	PreRankLimitMultiplier int `yaml:"preRankLimitMultiplier,omitempty" json:"preRankLimitMultiplier,omitempty" default:"10" validate:"gt=0"`

	// ADVANCED SETTING. Pre-ranking is based on word count. Results with a word count above this cutoff are not eligible for pre-ranking.
	// +kubebuilder:default=3
	PreRankWordCountCutoff int `yaml:"preRankWordCountCutoff,omitempty" json:"preRankWordCountCutoff,omitempty" default:"3" validate:"gt=0"`

	// ADVANCED SETTING. When true synonyms are taken into account during exact match calculation.
	// +kubebuilder:default=false
	SynonymsExactMatch bool `yaml:"synonymsExactMatch,omitempty" json:"synonymsExactMatch,omitempty" default:"false"`
}

// +kubebuilder:object:generate=true
type RelatedOGCAPIFeaturesCollection struct {
	// Base URL/Href to the OGC Features API.
	//
	// Only required when the given collection is hosted on a different server than the search API
	// (in a separate deployment).Otherwise, the base URL of this server is used.
	//
	// +kubebuilder:validation:Type=string
	APIBaseURL URL `yaml:"api,omitempty" json:"api,omitempty"`

	// Geometry type of the features in the related collection.
	// A collection in an OGC Features API has a single geometry type.
	// But a searchable collection has no geometry type distinction and thus
	// could be assembled of multiple OGC Feature API collections (with the same feature type).
	//
	// +kubebuilder:validation:Enum=point;multipoint;linestring;multilinestring;polygon;multipolygon
	GeometryType string `yaml:"geometryType" json:"geometryType" validate:"required"`

	// Collection ID in the OGC Features API. This can be a collection on this
	// server (listed under ogcApi>Features>Collections) or a remote collection on another server.
	CollectionID string `yaml:"collection" json:"collection" validate:"required,lowercase_id"`
}

func (rel *RelatedOGCAPIFeaturesCollection) CollectionURL(baseURL URL) string {
	if rel.APIBaseURL.URL != nil && rel.APIBaseURL.String() != "" {
		baseURL = rel.APIBaseURL
	}
	result, err := url.JoinPath(baseURL.String(), "collections", rel.CollectionID, "items")
	if err != nil {
		return ""
	}
	return result
}

// FeaturesAndSearchConfig Convince wrapper for OGC API Features and/or Features Search
type FeaturesAndSearchConfig struct {
	features *OgcAPIFeatures
	search   *OgcAPIFeaturesSearch
}

func NewFeaturesConfig(features *OgcAPIFeatures) FeaturesAndSearchConfig {
	return FeaturesAndSearchConfig{features, nil}
}

func NewSearchConfig(search *OgcAPIFeaturesSearch) FeaturesAndSearchConfig {
	return FeaturesAndSearchConfig{nil, search}
}

func (fas FeaturesAndSearchConfig) Datasources() *Datasources {
	if fas.search != nil {
		return fas.search.Datasources
	}
	return fas.features.Datasources
}

func (fas FeaturesAndSearchConfig) Collections() GeoSpatialCollections {
	if fas.search != nil {
		result := types.ToInterfaceSlice[FeaturesSearchCollection, GeoSpatialCollection](fas.search.Collections)
		return result
	}
	result := types.ToInterfaceSlice[FeaturesCollection, GeoSpatialCollection](fas.features.Collections)
	return result
}

func (fas FeaturesAndSearchConfig) FeatureCollections() FeaturesCollections {
	if fas.features != nil {
		return fas.features.Collections
	}
	return nil
}

func (fas FeaturesAndSearchConfig) MaxDecimals() int {
	if fas.search != nil {
		return fas.search.MaxDecimals
	}
	return fas.features.MaxDecimals
}

func (fas FeaturesAndSearchConfig) ForceUTC() bool {
	if fas.search != nil {
		return fas.search.ForceUTC
	}
	return fas.features.ForceUTC
}
