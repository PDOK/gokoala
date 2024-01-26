package domain

import (
	"github.com/go-spatial/geom/encoding/geojson"
)

// featureCollectionType allows the GeoJSON type to be automatically set during json marshalling
type featureCollectionType struct{}

func (fc *featureCollectionType) MarshalJSON() ([]byte, error) {
	return []byte(`"FeatureCollection"`), nil
}
func (fc *featureCollectionType) UnmarshalJSON([]byte) error { return nil }

// FeatureCollection is a GeoJSON FeatureCollection with extras such as links
type FeatureCollection struct {
	Type      featureCollectionType `json:"type"`
	Timestamp string                `json:"timeStamp,omitempty"`
	Links     []Link                `json:"links,omitempty"`

	Features []*Feature `json:"features"`

	NumberReturned int `json:"numberReturned"`
}

// Feature is a GeoJSON Feature with extras such as links
type Feature struct {
	geojson.Feature
	Links []Link `json:"links,omitempty"`

	// we overwrite ID since we want to make it a required attribute. We also expect feature ids to be
	// auto-incrementing integers (which is the default in geopackages) since we use it for cursor-based pagination.
	ID int64 `json:"id"`
}

// Link according to RFC 8288, https://datatracker.ietf.org/doc/html/rfc8288
type Link struct {
	Rel       string `json:"rel"`
	Title     string `json:"title,omitempty"`
	Type      string `json:"type,omitempty"`
	Href      string `json:"href"`
	Hreflang  string `json:"hreflang,omitempty"`
	Length    int64  `json:"length,omitempty"`
	Templated bool   `json:"templated,omitempty"`
}
