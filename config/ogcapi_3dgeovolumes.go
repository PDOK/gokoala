package config

// +kubebuilder:object:generate=true
type OgcAPI3dGeoVolumes struct {
	// Reference to the server (or object storage) hosting the 3D Tiles
	TileServer URL `yaml:"tileServer" json:"tileServer" validate:"required"`

	// Collections to be served as 3D GeoVolumes
	Collections GeoVolumesCollections `yaml:"collections" json:"collections"`

	// Whether JSON responses will be validated against the OpenAPI spec
	// since it has a significant performance impact when dealing with large JSON payloads.
	//
	// +kubebuilder:default=true
	// +optional
	ValidateResponses *bool `yaml:"validateResponses,omitempty" json:"validateResponses,omitempty" default:"true"` // ptr due to https://github.com/creasty/defaults/issues/49
}

type GeoVolumesCollections []GeoVolumesCollection

// ContainsID check if a given collection - by ID - exists.
func (csg GeoVolumesCollections) ContainsID(id string) bool {
	for _, coll := range csg {
		if coll.ID == id {
			return true
		}
	}
	return false
}

// +kubebuilder:object:generate=true
//
//nolint:recvcheck
type GeoVolumesCollection struct {
	// Unique ID of the collection
	// +kubebuilder:validation:Pattern=`^[a-z0-9"]([a-z0-9_-]*[a-z0-9"]+|)$`
	ID string `yaml:"id" validate:"required,lowercase_id" json:"id"`

	// Metadata describing the collection contents
	// +optional
	Metadata *GeoSpatialCollectionMetadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// Links pertaining to this collection (e.g., downloads, documentation)
	// +optional
	Links *CollectionLinks `yaml:"links,omitempty" json:"links,omitempty"`

	// Optional basepath to 3D tiles on the tileserver. Defaults to the collection ID.
	// +optional
	TileServerPath *string `yaml:"tileServerPath,omitempty" json:"tileServerPath,omitempty"`

	// Is a digital terrain model (DTM) in Quantized Mesh format, REQUIRED when you want to serve a DTM.
	// +kubebuilder:default=false
	// +optional
	IsDtm bool `yaml:"isDtm,omitempty" json:"isDtm,omitempty"`

	// Optional flag to indicate that the collection uses implicit tiling.
	// +optional
	IsImplicit bool `yaml:"isImplicit,omitempty" json:"isImplicit,omitempty"`

	// Optional URL to 3D viewer to visualize the given collection of 3D Tiles.
	// +optional
	URL3DViewer *URL `yaml:"3dViewerUrl,omitempty" json:"3dViewerUrl,omitempty"`
}

func (cgv GeoVolumesCollection) GetID() string {
	return cgv.ID
}

func (cgv GeoVolumesCollection) GetMetadata() *GeoSpatialCollectionMetadata {
	return cgv.Metadata
}

func (cgv GeoVolumesCollection) GetLinks() *CollectionLinks {
	return cgv.Links
}

func (cgv GeoVolumesCollection) Merge(other GeoSpatialCollection) GeoSpatialCollection {
	cgv.Metadata = mergeMetadata(cgv, other)
	cgv.Links = mergeLinks(cgv, other)
	return cgv
}
