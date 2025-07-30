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

	// URI template for individual 3D tiles.
	// +optional
	URITemplate3dTiles *string `yaml:"uriTemplate3dTiles,omitempty" json:"uriTemplate3dTiles,omitempty" validate:"required_without_all=URITemplateDTM"`

	// Optional URI template for subtrees, only required when "implicit tiling" extension is used.
	// +optional
	URITemplateImplicitTilingSubtree *string `yaml:"uriTemplateImplicitTilingSubtree,omitempty" json:"uriTemplateImplicitTilingSubtree,omitempty"`

	// URI template for digital terrain model (DTM) in Quantized Mesh format, REQUIRED when you want to serve a DTM.
	// +optional
	URITemplateDTM *string `yaml:"uriTemplateDTM,omitempty" json:"uriTemplateDTM,omitempty" validate:"required_without_all=URITemplate3dTiles"` //nolint:tagliatelle // grandfathered

	// Optional URL to 3D viewer to visualize the given collection of 3D Tiles.
	// +optional
	URL3DViewer *URL `yaml:"3dViewerUrl,omitempty" json:"3dViewerUrl,omitempty"`
}

func (gv *CollectionEntry3dGeoVolumes) Has3DTiles() bool {
	return gv.URITemplate3dTiles != nil
}

func (gv *CollectionEntry3dGeoVolumes) HasDTM() bool {
	return gv.URITemplateDTM != nil
}
