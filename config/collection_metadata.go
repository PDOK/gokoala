package config

// +kubebuilder:object:generate=true
type GeoSpatialCollectionMetadata struct {
	// Human-friendly title of this collection. When no title is specified the collection ID is used.
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
type Extent struct {
	// Projection (SRS/CRS) to be used. When none is provided WGS84 (http://www.opengis.net/def/crs/OGC/1.3/CRS84) is used.
	// +optional
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	Srs string `yaml:"srs,omitempty" json:"srs,omitempty" validate:"omitempty,startswith=EPSG:"`

	// Geospatial extent
	Bbox []string `yaml:"bbox" json:"bbox"`

	// Temporal extent
	// +optional
	// +kubebuilder:validation:MinItems=2
	// +kubebuilder:validation:MaxItems=2
	Interval []string `yaml:"interval,omitempty" json:"interval,omitempty" validate:"omitempty,len=2"`
}

// +kubebuilder:object:generate=true
type CollectionLinks struct {
	// Links to downloads of an entire collection. These will be rendered as rel=enclosure links
	// +optional
	Downloads []DownloadLink `yaml:"downloads,omitempty" json:"downloads,omitempty" validate:"dive"`

	// Links to documentation describing the collection. These will be rendered as rel=describedby links
	// <placeholder>
}

// +kubebuilder:object:generate=true
type DownloadLink struct {
	// Name of the provided download
	Name string `yaml:"name" json:"name" validate:"required"`

	// Full URL to the file to be downloaded
	AssetURL *URL `yaml:"assetUrl" json:"assetUrl" validate:"required"`

	// Approximate size of the file to be downloaded
	// +optional
	Size string `yaml:"size,omitempty" json:"size,omitempty"`

	// Media type of the file to be downloaded
	MediaType MediaType `yaml:"mediaType" json:"mediaType" validate:"required"`
}
