package config

import (
	"net/url"

	"github.com/PDOK/gokoala/internal/engine/types"
	"gopkg.in/yaml.v3"
)

// +kubebuilder:object:generate=true
type OgcAPIFeaturesSearch struct {
	// Builds on top of the OGC API Features configuration.
	OgcAPIFeatures `yaml:",inline" json:",inline"`

	// Collections available for search through this API
	Collections CollectionsFeaturesSearch `yaml:"collections" json:"collections" validate:"required,dive"`

	// Settings related to the search API/index.
	// +optional
	SearchSettings SearchSettings `yaml:"searchSettings" json:"searchSettings"`
}

// UnmarshalYAML Handles YAML unmarshalling conflict with the "collections" field
// present in both OgcAPIFeaturesSearch and embedded OgcAPIFeatures.
func (c *OgcAPIFeaturesSearch) UnmarshalYAML(value *yaml.Node) error {
	type base OgcAPIFeatures // empty struct/copy to avoid a possible infinite loop
	if err := value.Decode((*base)(&c.OgcAPIFeatures)); err != nil {
		return err
	}
	// Favor the 'collections' field from OgcAPIFeaturesSearch
	pairSize := 2
	for i := 0; i < len(value.Content); i += pairSize {
		if value.Content[i].Value == "collections" {
			return value.Content[i+1].Decode(&c.Collections)
		}
	}
	return nil
}

type CollectionsFeaturesSearch []CollectionFeaturesSearch

// ContainsID check if a given collection - by ID - exists.
func (csfs CollectionsFeaturesSearch) ContainsID(id string) bool {
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
type CollectionFeaturesSearch struct {
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

func (cfs CollectionFeaturesSearch) GetID() string {
	return cfs.ID
}

func (cfs CollectionFeaturesSearch) GetMetadata() *GeoSpatialCollectionMetadata {
	return cfs.Metadata
}

func (cfs CollectionFeaturesSearch) GetLinks() *CollectionLinks {
	return cfs.Links
}

func (cfs CollectionFeaturesSearch) Merge(other GeoSpatialCollection) GeoSpatialCollection {
	cfs.Metadata = mergeMetadata(cfs, other)
	cfs.Links = mergeLinks(cfs, other)
	return cfs
}

// IsRemoteFeatureCollection true when the given collection ID is defined as a feature collection outside this config.
// In other words: it references a remote feature collection and doesn't point to a local one in this dataset.
func (cfs CollectionFeaturesSearch) IsRemoteFeatureCollection(collID string) bool {
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

type FeaturesAndSearchConfig struct {
	Features *OgcAPIFeatures
	Search   *OgcAPIFeaturesSearch
}

func (fas FeaturesAndSearchConfig) Datasources() *Datasources {
	if fas.Search != nil {
		return fas.Search.Datasources
	}
	return fas.Features.Datasources
}

func (fas FeaturesAndSearchConfig) Collections() GeoSpatialCollections {
	if fas.Search != nil {
		result := types.ToInterfaceSlice[CollectionFeaturesSearch, GeoSpatialCollection](fas.Search.Collections)
		return result
	}
	result := types.ToInterfaceSlice[CollectionFeatures, GeoSpatialCollection](fas.Features.Collections)
	return result
}

func (fas FeaturesAndSearchConfig) FeatureCollections() CollectionsFeatures {
	if fas.Features != nil {
		return fas.Features.Collections
	}
	return nil
}

func (fas FeaturesAndSearchConfig) MaxDecimals() int {
	if fas.Search != nil {
		return fas.Search.MaxDecimals
	}
	return fas.Features.MaxDecimals
}

func (fas FeaturesAndSearchConfig) ForceUTC() bool {
	if fas.Search != nil {
		return fas.Search.ForceUTC
	}
	return fas.Features.ForceUTC
}
