package config

// +kubebuilder:object:generate=true
type OgcAPI3dGeoVolumes struct {
	// Reference to the server (or object storage) hosting the 3D Tiles
	TileServer URL `yaml:"tileServer" json:"tileServer" validate:"required"`

	// Collections to be served as 3D GeoVolumes
	Collections Collections3dGeoVolumes `yaml:"collections" json:"collections"`

	// Whether JSON responses will be validated against the OpenAPI spec
	// since it has a significant performance impact when dealing with large JSON payloads.
	//
	// +kubebuilder:default=true
	// +optional
	ValidateResponses *bool `yaml:"validateResponses,omitempty" json:"validateResponses,omitempty" default:"true"` // ptr due to https://github.com/creasty/defaults/issues/49
}

type Collections3dGeoVolumes []Collection3dGeoVolumes

// ContainsID check if a given collection - by ID - exists.
func (csg Collections3dGeoVolumes) ContainsID(id string) bool {
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
type Collection3dGeoVolumes struct {
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

	// Optional flag to indicate that the collection is implicit.
	// +optional
	IsImplicit bool `yaml:"isImplicit,omitempty" json:"isImplicit,omitempty"`

	// Optional URL to 3D viewer to visualize the given collection of 3D Tiles.
	// +optional
	URL3DViewer *URL `yaml:"3dViewerUrl,omitempty" json:"3dViewerUrl,omitempty"`
}

func (cgv Collection3dGeoVolumes) GetType() string {
	return getGeoSpatialCollectionType(cgv)
}

func (cgv Collection3dGeoVolumes) GetID() string {
	return cgv.ID
}

func (cgv Collection3dGeoVolumes) GetMetadata() *GeoSpatialCollectionMetadata {
	return cgv.Metadata
}

func (cgv Collection3dGeoVolumes) GetLinks() *CollectionLinks {
	return cgv.Links
}

func (cgv Collection3dGeoVolumes) HasDateTime() bool {
	return cgv.Metadata != nil && cgv.Metadata.TemporalProperties != nil
}

func (cgv Collection3dGeoVolumes) HasTableName(_ string) bool {
	return false
}

func (cgv Collection3dGeoVolumes) Merge(other GeoSpatialCollection) GeoSpatialCollection {
	cgv.Metadata = mergeMetadata(cgv, other)
	cgv.Links = mergeLinks(cgv, other)
	return cgv
}
