package config

import "github.com/google/uuid"

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

	// Optional configuration for generation of external_fid
	// +optional
	ExternalFid *ExternalFid `yaml:"externalFid,omitempty" json:"externalFid,omitempty"`
}

type ExternalFid struct {
	// Namespace (UUID5) used to generate external_fid, defaults to uuid.NameSpaceURL
	// +kubebuilder:default="6ba7b811-9dad-11d1-80b4-00c04fd430c8"
	UUIDNamespace uuid.UUID `yaml:"uuidNamespace,omitempty" json:"uuidNamespace,omitempty" default:"6ba7b811-9dad-11d1-80b4-00c04fd430c8" validate:"required"`

	// Fields used to generate external_fid in the target OGC Features Collection(s).
	// Field names should match those in the source datasource.
	Fields []string `yaml:"fields" json:"fields" validate:"required"`
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
