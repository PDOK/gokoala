package config

// +kubebuilder:object:generate=true
type OgcAPIMaps struct {
	// Base URL of the MapServer WMS service.
	WMSServer *URL `yaml:"wmsServer" json:"wmsServer" validate:"required"`

	// Default map view used when request parameters are omitted.
	DefaultMap DefaultMap `yaml:"defaultMap" json:"defaultMap" validate:"required"`

	// Collections to be served as maps through this API.
	// +kubebuilder:validation:MinItems=1
	Collections []MapsCollection `yaml:"collections" json:"collections" validate:"required,min=1,dive"`
}

// +kubebuilder:object:generate=true
type DefaultMap struct {
	// CRS of the default map view, supported by the WMS service.
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	CRS string `yaml:"crs" json:"crs" validate:"required,startswith=EPSG:"`

	// Width of the map image in pixels.
	// +kubebuilder:validation:Minimum=1
	Width int `yaml:"width" json:"width" validate:"gt=0"`

	// Height of the map image in pixels.
	// +kubebuilder:validation:Minimum=1
	Height int `yaml:"height" json:"height" validate:"gt=0"`

	// Bounding box of the default map view.
	// +kubebuilder:validation:MinItems=4
	// +kubebuilder:validation:MaxItems=4
	Bbox []string `yaml:"bbox" json:"bbox" validate:"required,len=4,dive,numeric"`
}

// +kubebuilder:object:generate=true
type MapsCollection struct {
	// Unique ID of the collection.
	// +kubebuilder:validation:Pattern=`^[a-z0-9"]([a-z0-9_-]*[a-z0-9"]+|)$`
	ID string `yaml:"id" json:"id" validate:"required,lowercase_id"`

	// Metadata describing the collection contents.
	// +optional
	Metadata *GeoSpatialCollectionMetadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// WMS layers included in this collection.
	// +kubebuilder:validation:MinItems=1
	Layers []MapLayer `yaml:"layers" json:"layers" validate:"required,min=1,dive"`

	// Default CRS for this collection, supported by its WMS layers.
	// +optional
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	DefaultCRS string `yaml:"defaultCrs,omitempty" json:"defaultCrs,omitempty" validate:"omitempty,startswith=EPSG:"`

	// Bounding box of this collection.
	// +optional
	// +kubebuilder:validation:MinItems=4
	// +kubebuilder:validation:MaxItems=4
	Bbox []string `yaml:"bbox,omitempty" json:"bbox,omitempty" validate:"omitempty,len=4,dive,numeric"`
}

// +kubebuilder:object:generate=true
type MapLayer struct {
	// Unique ID of the layer within this collection.
	// +kubebuilder:validation:Pattern=`^[a-z0-9"]([a-z0-9_-]*[a-z0-9"]+|)$`
	ID string `yaml:"id" json:"id" validate:"required,lowercase_id"`

	// Layer name advertised by the WMS service.
	Name string `yaml:"name" json:"name" validate:"required"`

	// Human-friendly title of the layer.
	// +optional
	Title string `yaml:"title,omitempty" json:"title,omitempty"`

	// Styles advertised by the WMS service for this layer.
	// +optional
	Styles []MapStyle `yaml:"styles,omitempty" json:"styles,omitempty" validate:"dive"`
}

// +kubebuilder:object:generate=true
type MapStyle struct {
	// Unique ID of the style within this layer.
	// +kubebuilder:validation:Pattern=`^[a-z0-9"]([a-z0-9_-]*[a-z0-9"]+|)$`
	ID string `yaml:"id" json:"id" validate:"required,lowercase_id"`

	// Style name advertised by the WMS service.
	WMSStyle string `yaml:"wmsStyle" json:"wmsStyle" validate:"required"`
}
