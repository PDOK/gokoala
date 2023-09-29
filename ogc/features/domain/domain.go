package domain

import (
	"log"
	"strconv"

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

	NumberReturned int                   `json:"numberReturned"`
	Type           featureCollectionType `json:"type"`
	Features       []*Feature            `json:"features"`
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

// Cursor since we use cursor-based pagination as opposed to offset-based pagination
type Cursor struct {
	Prev EncodedCursorValue
	Next EncodedCursorValue

	IsFirst bool
	IsLast  bool
}

type EncodedCursorValue string

func (c EncodedCursorValue) Decode() int {
	decoded := string(c)
	var result int
	if decoded == "" {
		result = 0
	} else {
		result, _ = strconv.Atoi(decoded)
		if result < 0 {
			result = 0
		}
	}
	return result
}

func encodeCursorValue(value int) EncodedCursorValue {
	return EncodedCursorValue(strconv.Itoa(value))
}

func NewCursor(features []*Feature, column string, limit int, last bool) Cursor {
	if len(features) == 0 {
		return Cursor{}
	}
	max := len(features) - 1

	start := features[0].Properties[column]
	end := features[max].Properties[column]

	if start == nil {
		log.Printf("cursor column '%s' doesn't exists, defaulting to first page\n", column)
		start = 0
	}
	if end == nil {
		log.Printf("cursor column '%s' doesn't exists, defaulting to first page\n", column)
		end = 0
	}

	prev := start.(int)
	if prev != 0 {
		prev -= max
		if prev < 0 {
			prev = 0
		}
	}
	next := end.(int)

	return Cursor{
		Prev: encodeCursorValue(prev),
		Next: encodeCursorValue(next),

		IsFirst: next < limit,
		IsLast:  last,
	}
}
