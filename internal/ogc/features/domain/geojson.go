package domain

import (
	"github.com/go-spatial/geom/encoding/geojson"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// featureCollectionType allows the GeoJSON type to be automatically set during json marshalling
type featureCollectionType struct{}

func (fc *featureCollectionType) MarshalJSON() ([]byte, error) {
	return []byte(`"FeatureCollection"`), nil
}

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

	// we overwrite ID since we want to make it a required attribute. We also expect feature ids to be
	// auto-incrementing integers (which is the default in geopackages) since we use it for cursor-based pagination.
	ID    string `json:"id"`
	Links []Link `json:"links,omitempty"`

	Properties orderedmap.OrderedMap[string, any] `json:"properties"`
}

// Keys of the Feature properties.
//
// Note: In the future we might replace this with Go 1.23 iterators (range-over-func) however at the moment this
// isn't supported in Go templates: https://github.com/golang/go/pull/68329
func (f *Feature) Keys() []string {
	result := make([]string, 0, f.Properties.Len())
	for pair := f.Properties.Oldest(); pair != nil; pair = pair.Next() {
		result = append(result, pair.Key)
	}
	return result
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
