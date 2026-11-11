package config

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// +kubebuilder:object:generate=true
type OgcAPIMaps struct {
	// Basemap to use in embedded viewer on the HTML pages.
	// +kubebuilder:default="OSM"
	// +kubebuilder:validation:Enum=OSM;BRT
	// +optional
	Basemap string `yaml:"basemap,omitempty" json:"basemap,omitempty" default:"OSM" validate:"oneof=OSM BRT"`

	// Base URL of the MapServer WMS service.
	WMSServer *URL `yaml:"wmsServer" json:"wmsServer" validate:"required"`

	// Maximum map width in pixels.
	// +optional
	// +kubebuilder:default=4000
	// +kubebuilder:validation:Minimum=1
	MaxWidth int `yaml:"maxWidth,omitempty" json:"maxWidth,omitempty" default:"4000" validate:"gt=0"`

	// Maximum map height in pixels.
	// +optional
	// +kubebuilder:default=4000
	// +kubebuilder:validation:Minimum=1
	MaxHeight int `yaml:"maxHeight,omitempty" json:"maxHeight,omitempty" default:"4000" validate:"gt=0"`

	// Default map view used when request parameters are omitted.
	DefaultMap DefaultMap `yaml:"defaultMap" json:"defaultMap" validate:"required"`

	// Collections to be served as maps through this API.
	Collections MapsCollections `yaml:"collections" json:"collections" validate:"required,min=1,dive"`
}

type MapsCollections []MapsCollection

// ContainsID check if a given collection - by ID - exists.
func (csm MapsCollections) ContainsID(id string) bool {
	for _, coll := range csm {
		if coll.ID == id {
			return true
		}
	}
	return false
}

func (o *OgcAPIMaps) Defaults() {
	if o.DefaultMap.MMPerPixel == "" {
		o.DefaultMap.MMPerPixel = "0.28"
	}
}

func (o *OgcAPIMaps) MaxPixels() int64 {
	return int64(o.MaxWidth) * int64(o.MaxHeight)
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
	Width int `yaml:"width,omitempty" json:"width,omitempty" default:"1024" validate:"gt=0"`

	// Height of the map image in pixels. Defaults to 768.
	// +optional
	// +kubebuilder:default=768
	// +kubebuilder:validation:Minimum=1
	Height int `yaml:"height,omitempty" json:"height,omitempty" default:"768" validate:"gt=0"`

	// Bounding box of the default map view.
	// +kubebuilder:validation:MinItems=4
	// +kubebuilder:validation:MaxItems=4
	Bbox []string `yaml:"bbox" json:"bbox" validate:"required,len=4,dive,numeric"`

	// Size of a rendering pixel in millimeters. Defaults to 0.28.
	// OGC API Maps Display Resolution conformance requires 0.28 when
	// mm-per-pixel is omitted. Overriding this default breaks that conformance.
	// +optional
	// +kubebuilder:default="0.28"
	// +kubebuilder:validation:Pattern=`^\d*\.?\d+$`
	MMPerPixel string `yaml:"mmPerPixel,omitempty" json:"mmPerPixel,omitempty" validate:"numeric"`
}

// +kubebuilder:object:generate=true
type MapsCollection struct { //nolint:recvcheck
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

func (cm MapsCollection) GetID() string {
	return cm.ID
}

func (cm MapsCollection) GetMetadata() *GeoSpatialCollectionMetadata {
	return cm.Metadata
}

func (cm MapsCollection) GetLinks() *CollectionLinks {
	return cm.Links
}

func (cm MapsCollection) Merge(other GeoSpatialCollection) GeoSpatialCollection {
	cm.Metadata = mergeMetadata(cm, other)
	cm.Links = mergeLinks(cm, other)
	return cm
}

func (cm MapsCollection) GetStyle(styleID string) *MapStyle {
	for _, layer := range cm.Layers {
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

func validateMapCollections(collections []MapsCollection, defaultCRS string) error {
	var errMessages []string
	defaultCode, defaultErr := parseMapEPSGCode(defaultCRS)
	collectionIDs := make(map[string]struct{})
	for _, collection := range collections {
		if _, exists := collectionIDs[collection.ID]; exists {
			errMessages = append(errMessages, fmt.Sprintf("duplicate maps collection ID '%s'", collection.ID))
		}
		collectionIDs[collection.ID] = struct{}{}

		supported := false
		for _, srs := range collection.SupportedSrs {
			code, err := parseMapEPSGCode(srs.Srs)
			if err != nil {
				errMessages = append(errMessages, fmt.Sprintf("invalid supported CRS for maps collection %q: %v", collection.ID, err))
				continue
			}
			if defaultErr == nil && code == defaultCode {
				supported = true
			}
		}
		if defaultErr == nil && !supported {
			errMessages = append(errMessages, fmt.Sprintf("default map CRS %q must be supported by maps collection %q", defaultCRS, collection.ID))
		}

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

func validateMapDefaults(dm DefaultMap) error {
	return errors.Join(
		validateMMPerPixel(dm.MMPerPixel),
		validateMapCRS(dm.CRS),
		validateMapBbox(dm.Bbox),
	)
}

func validateMapBbox(bbox []string) error {
	if len(bbox) != 4 {
		return errors.New("default map bbox must contain four coordinates")
	}
	var coordinates [4]float64
	for index, coordinate := range bbox {
		value, err := strconv.ParseFloat(coordinate, 64)
		if err != nil {
			return fmt.Errorf("invalid default map bbox coordinate %q: %w", coordinate, err)
		}
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("default map bbox coordinate %q must be finite", coordinate)
		}
		coordinates[index] = value
	}
	width := coordinates[2] - coordinates[0]
	height := coordinates[3] - coordinates[1]
	if width <= 0 || height <= 0 || math.IsInf(width, 0) || math.IsInf(height, 0) {
		return errors.New("default map bbox must have finite positive extents")
	}
	ratio := width / height
	if ratio <= 0 || math.IsInf(ratio, 0) {
		return errors.New("default map bbox must have a finite positive aspect ratio")
	}
	return nil
}

func validateMapCRS(crs string) error {
	_, err := parseMapEPSGCode(crs)
	if err != nil {
		return fmt.Errorf("invalid default map CRS: %w", err)
	}
	return nil
}

func parseMapEPSGCode(crs string) (int, error) {
	code, found := strings.CutPrefix(crs, "EPSG:")
	if !found || code == "" || strings.Trim(code, "0123456789") != "" {
		return 0, fmt.Errorf("expected EPSG CRS with a positive numeric code, got %q", crs)
	}
	value, err := strconv.Atoi(code)
	if err != nil {
		return 0, fmt.Errorf("invalid EPSG code in %q: %w", crs, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("EPSG code must be positive, got %q", crs)
	}
	return value, nil
}

func validateMapLimits(o *OgcAPIMaps) error {
	var errMessages []string
	largestSafeWidth := int64(math.MaxInt64) / int64(o.MaxHeight)
	if int64(o.MaxWidth) > largestSafeWidth {
		errMessages = append(errMessages, "maps maxWidth times maxHeight exceeds the maximum supported pixel count")
	}
	if o.DefaultMap.Width > o.MaxWidth {
		errMessages = append(errMessages, fmt.Sprintf("maps defaultMap.width must not exceed maxWidth (%d)", o.MaxWidth))
	}
	if o.DefaultMap.Height > o.MaxHeight {
		errMessages = append(errMessages, fmt.Sprintf("maps defaultMap.height must not exceed maxHeight (%d)", o.MaxHeight))
	}
	if len(errMessages) > 0 {
		return fmt.Errorf("invalid config provided:\n%v", errMessages)
	}
	return nil
}

func validateMMPerPixel(mmPerPixel string) error {
	value, err := strconv.ParseFloat(mmPerPixel, 64)
	if err != nil {
		return fmt.Errorf("invalid mmPerPixel '%s' in default map: %w", mmPerPixel, err)
	}

	if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
		return fmt.Errorf("mmPerPixel in default map must be finite and positive, got '%s'", mmPerPixel)
	}

	return nil
}
