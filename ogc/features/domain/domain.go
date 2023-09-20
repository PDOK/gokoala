package domain

import (
	"log"

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
	Links []Link `json:"links,omitempty"`

	Type     featureCollectionType `json:"type"`
	Features []*Feature            `json:"features"`
}

// Feature is a GeoJSON Feature with extras such as links
type Feature struct {
	// overwrite ID in geojson.Feature so strings are also allowed as id
	ID    string `json:"id,omitempty"`
	Links []Link `json:"links,omitempty"`

	geojson.Feature
}

// Link according to RFC 8288, https://datatracker.ietf.org/doc/html/rfc8288
type Link struct {
	Length    int64  `json:"length,omitempty"`
	Rel       string `json:"rel"`
	Title     string `json:"title,omitempty"`
	Type      string `json:"type,omitempty"`
	Href      string `json:"href"`
	Hreflang  string `json:"hreflang,omitempty"`
	Templated bool   `json:"templated,omitempty"`
}

type Cursor struct {
	Prev int
	Next int

	IsFirst bool
	IsLast  bool
}

func NewCursor(features []*Feature, column string, limit int, last bool) Cursor {
	if len(features) == 0 {
		return Cursor{}
	}
	max := len(features) - 1

	prev := features[0].Properties[column]
	if prev == nil {
		log.Printf("cursor column '%s' doesn't exists, defaulting to first page\n", column)
		prev = 0
	} else if prev != 0 {
		prev = prev.(int) - max
	}

	next := features[max].Properties[column]
	if next == nil {
		log.Printf("cursor column '%s' doesn't exists, defaulting to first page\n", column)
		next = 0
	}

	return Cursor{
		Prev: prev.(int),
		Next: next.(int),

		IsFirst: next.(int) < limit,
		IsLast:  last,
	}
}
