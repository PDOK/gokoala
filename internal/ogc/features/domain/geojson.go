package domain

import (
	geojson2 "github.com/twpayne/go-geom/encoding/geojson"
)

// featureCollectionType allows the GeoJSON type to be automatically set during json marshalling
type featureCollectionType struct{}

func (fc *featureCollectionType) MarshalJSON() ([]byte, error) {
	return []byte(`"FeatureCollection"`), nil
}

// featureType allows the type for Feature to be automatically set during json Marshalling
type featureType struct{}

func (ft *featureType) MarshalJSON() ([]byte, error) {
	return []byte(`"Feature"`), nil
}

// FeatureCollection is a GeoJSON FeatureCollection with extras such as links
// Note: fields in this struct are sorted for optimal memory usage (field alignment)
type FeatureCollection struct {
	Type           featureCollectionType `json:"type"`
	Timestamp      string                `json:"timeStamp,omitempty"`
	Links          []Link                `json:"links,omitempty"`
	Features       []*Feature            `json:"features"`
	NumberReturned int                   `json:"numberReturned"`
}

// Feature is a GeoJSON Feature with extras such as links
// Note: fields in this struct are sorted for optimal memory usage (field alignment)
type Feature struct {
	Type       featureType        `json:"type"`
	Properties FeatureProperties  `json:"properties"`
	Geometry   *geojson2.Geometry `json:"geometry"`
	// We expect feature ids to be auto-incrementing integers (which is the default in geopackages)
	// since we use it for cursor-based pagination.
	ID    string `json:"id"`
	Links []Link `json:"links,omitempty"`
}

// Keys of the Feature properties.
func (f *Feature) Keys() []string {
	return f.Properties.Keys()
}

// Link according to RFC 8288, https://datatracker.ietf.org/doc/html/rfc8288
// Note: fields in this struct are sorted for optimal memory usage (field alignment)
type Link struct {
	Rel       string `json:"rel"`
	Title     string `json:"title,omitempty"`
	Type      string `json:"type,omitempty"`
	Href      string `json:"href"`
	Hreflang  string `json:"hreflang,omitempty"`
	Length    int64  `json:"length,omitempty"`
	Templated bool   `json:"templated,omitempty"`
}
