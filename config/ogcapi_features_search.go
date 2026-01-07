package config

// +kubebuilder:object:generate=true
type OgcAPIFeaturesSearch struct {
	// Builds on top of the OGC API Features configuration.
	OgcAPIFeatures `yaml:",inline" json:",inline"`

	// Settings related to the search API/index.
	SearchSettings SearchSettings `yaml:"searchSettings" json:"searchSettings" validate:"required"`
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
	// +kubebuilder:default=3.0
	ExactMatchMultiplier string `yaml:"exactMatchMultiplier,omitempty" json:"exactMatchMultiplier,omitempty" default:"3.0" validate:"numeric,gt=0"`

	// ADVANCED SETTING. The primary suggest is equal to the display name. With this multiplier you can boost it above other suggests.
	// +kubebuilder:validation:Pattern=`^-?\d+(\.\d+)?$`
	// +kubebuilder:default=1.01
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
type CollectionEntryFeaturesSearch struct {
	// Fields that make up the display name and/or suggestions. These fields can be used as variables in the DisplayNameTemplate.
	// +kubebuilder:validation:MinItems=1
	Fields []string `yaml:"fields,omitempty" json:"fields,omitempty" validate:"required,unique"`

	// Template that indicates how a search record is displayed. Uses Go text/template syntax to reference fields.
	DisplayNameTemplate string `yaml:"displayNameTemplate,omitempty" json:"displayNameTemplate,omitempty" validate:"required"`

	// Version of the collection exposed through the API.
	// +kubebuilder:default=1
	Version int `yaml:"version,omitempty" json:"version,omitempty" default:"1"`

	// Links to the individual OGC API (feature) collections that are searchable in this collection.
	// +kubebuilder:validation:MinItems=1
	CollectionRefs []RelatedOGCAPIFeaturesCollection `yaml:"collectionRefs" json:"collectionRefs" validate:"required,min=1"`
}

// +kubebuilder:object:generate=true
type RelatedOGCAPIFeaturesCollection struct {
	// Base URL/Href to the OGC Features API
	// +kubebuilder:validation:Type=string
	APIBaseURL URL `yaml:"api" json:"api" validate:"required"`

	// Geometry type of the features in the related collection.
	// A collection in an OGC Features API has a single geometry type.
	// But a searchable collection has no geometry type distinction and thus
	// could be assembled of multiple OGC Feature API collections (with the same feature type).
	GeometryType string `yaml:"geometryType" json:"geometryType" validate:"required"`

	// Collection ID in the OGC Features API
	CollectionID string `yaml:"collection" json:"collection" validate:"required,lowercase_id"`

	// `datetime` query parameter for the OGC Features API. In case it's temporal.
	// E.g.: "{now()-1h}"
	// +optional
	Datetime *string `yaml:"datetime,omitempty" json:"datetime,omitempty"`
}
