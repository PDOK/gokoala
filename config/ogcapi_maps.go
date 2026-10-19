package config

import "fmt"

// +kubebuilder:object:generate=true
type OgcAPIMaps struct {
	// Base URL of the MapServer WMS service.
	WMSServer *URL `yaml:"wmsServer" json:"wmsServer" validate:"required"`

	// Default map view used when request parameters are omitted.
	DefaultMap DefaultMap `yaml:"defaultMap" json:"defaultMap" validate:"required"`

	// Collections to be served as maps through this API.
	// +kubebuilder:validation:MinItems=1
	Collections MapsCollections `yaml:"collections" json:"collections" validate:"required,min=1,dive"`
}

type MapsCollections []MapsCollection

// ContainsID check if a given collection - by ID - exists.
func (mc MapsCollections) ContainsID(id string) bool {
	for _, coll := range mc {
		if coll.ID == id {
			return true
		}
	}
	return false
}

func (o *OgcAPIMaps) Defaults() {
	if o.DefaultMap.Width == 0 {
		o.DefaultMap.Width = 1024
	}
	if o.DefaultMap.Height == 0 {
		o.DefaultMap.Height = 768
	}
}

// +kubebuilder:object:generate=true
type DefaultMap struct {
	// CRS of the default map view.
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	CRS string `yaml:"crs" json:"crs" validate:"required,startswith=EPSG:"`

	// Width of the map image in pixels. Defaults to 1024.
	// +optional
	// +kubebuilder:default=1024
	// +kubebuilder:validation:Minimum=1
	Width int `yaml:"width,omitempty" json:"width,omitempty" validate:"gt=0"`

	// Height of the map image in pixels. Defaults to 768.
	// +optional
	// +kubebuilder:default=768
	// +kubebuilder:validation:Minimum=1
	Height int `yaml:"height,omitempty" json:"height,omitempty" validate:"gt=0"`

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

	// Links from/to this collection
	// +optional
	Links *CollectionLinks `yaml:"links,omitempty" json:"links,omitempty"`

	// WMS layers included in this collection.
	// +kubebuilder:validation:MinItems=1
	Layers []MapLayer `yaml:"layers" json:"layers" validate:"required,min=1,dive"`

	// List of supported CRS for this collection.
	// +kubebuilder:validation:MinItems=1
	SupportedSrs []MapsSupportedSrs `yaml:"supportedSrs" json:"supportedSrs" validate:"required,min=1,dive"`

	// Bounding box of this collection.
	// +optional
	// +kubebuilder:validation:MinItems=4
	// +kubebuilder:validation:MaxItems=4
	Bbox []string `yaml:"bbox,omitempty" json:"bbox,omitempty" validate:"omitempty,len=4,dive,numeric"`
}

func (mf MapsCollection) GetID() string {
	return mf.ID
}

func (mf MapsCollection) GetMetadata() *GeoSpatialCollectionMetadata {
	return mf.Metadata
}

func (mf MapsCollection) GetLinks() *CollectionLinks {
	return mf.Links
}

func (mf MapsCollection) Merge(other GeoSpatialCollection) GeoSpatialCollection {
	mf.Metadata = mergeMetadata(mf, other)
	mf.Links = mergeLinks(mf, other)
	return mf
}

func (mf MapsCollection) GetStyle(styleID string) *MapStyle {
	for _, layer := range mf.Layers {
		for _, style := range layer.Styles {
			if style.ID == styleID {
				return &style
			}
		}
	}

	return nil
}

// +kubebuilder:object:generate=true
type MapsSupportedSrs struct {
	// Supported coordinate reference system.
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	Srs string `yaml:"srs" json:"srs" validate:"required,startswith=EPSG:"`
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

func validateMapCollections(collections []MapsCollection) error {
	var errMessages []string
	collectionIDs := make(map[string]struct{})
	for _, collection := range collections {
		if _, exists := collectionIDs[collection.ID]; exists {
			errMessages = append(errMessages, fmt.Sprintf("duplicate maps collection ID '%s'", collection.ID))
		}
		collectionIDs[collection.ID] = struct{}{}

		layerIDs := make(map[string]struct{})
		for _, layer := range collection.Layers {
			if _, exists := layerIDs[layer.ID]; exists {
				errMessages = append(errMessages, fmt.Sprintf("duplicate layer ID '%s' in maps collection '%s'", layer.ID, collection.ID))
			}
			layerIDs[layer.ID] = struct{}{}

			styleIDs := make(map[string]struct{})
			for _, style := range layer.Styles {
				if _, exists := styleIDs[style.ID]; exists {
					errMessages = append(errMessages, fmt.Sprintf("duplicate style ID '%s' in layer '%s' of maps collection '%s'", style.ID, layer.ID, collection.ID))
				}
				styleIDs[style.ID] = struct{}{}
			}
		}
	}

	if len(errMessages) > 0 {
		return fmt.Errorf("invalid config provided:\n%v", errMessages)
	}

	return nil
}
