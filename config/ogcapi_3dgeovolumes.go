package config

// +kubebuilder:object:generate=true
type OgcAPI3dGeoVolumes struct {
	// Reference to the server (or object storage) hosting the 3D Tiles
	TileServer URL `yaml:"tileServer" json:"tileServer" validate:"required"`

	// Collections to be served as 3D GeoVolumes
	Collections GeoSpatialCollections `yaml:"collections" json:"collections"`

	// Whether JSON responses will be validated against the OpenAPI spec
	// since it has a significant performance impact when dealing with large JSON payloads.
	//
	// +kubebuilder:default=true
	// +optional
	ValidateResponses *bool `yaml:"validateResponses,omitempty" json:"validateResponses,omitempty" default:"true"` // ptr due to https://github.com/creasty/defaults/issues/49
}

// +kubebuilder:object:generate=true
type CollectionEntry3dGeoVolumes struct {
	// Optional basepath to 3D tiles on the tileserver. Defaults to the collection ID.
	// +optional
	TileServerPath *string `yaml:"tileServerPath,omitempty" json:"tileServerPath,omitempty"`

	// Is a digital terrain model (DTM) in Quantized Mesh format, REQUIRED when you want to serve a DTM.
	// +kubebuilder:default=false
	// +optional
	IsDTM bool `yaml:"isDTM,omitempty" json:"isDTM,omitempty" default:"false"`

	// Optional flag to indicate that the collection is implicit.
	// +optional
	IsImplicit bool `yaml:"isImplicit,omitempty" json:"isImplicit,omitempty"`

	// Optional URL to 3D viewer to visualize the given collection of 3D Tiles.
	// +optional
	URL3DViewer *URL `yaml:"3dViewerUrl,omitempty" json:"3dViewerUrl,omitempty"`
}
