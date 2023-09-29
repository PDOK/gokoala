package domain

import (
	"log"
	"strconv"

	"github.com/go-spatial/geom/encoding/geojson"
	"github.com/sqids/sqids-go"
)

const (
	cursorAlphabet = "1Vti5BYcjOdTXunDozKPm4syvG6galxLM8eIrUS2bWqZCNkwpR309JFAHfh7EQ" // generated on https://sqids.org/playground
)

var (
	cursorCodec, _ = sqids.New(sqids.Options{
		Alphabet:  cursorAlphabet,
		Blocklist: nil, // disable blocklist
		MinLength: 8,
	})
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
	Prev EncodedCursor
	Next EncodedCursor

	IsFirst bool
	IsLast  bool
}

func NewCursor(features []*Feature, column string, limit int, last bool) Cursor {
	if len(features) == 0 {
		return Cursor{}
	}
	max := int64(len(features) - 1)

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

	prev := start.(int64)
	if prev != 0 {
		prev -= max
		if prev < 0 {
			prev = 0
		}
	}
	next := end.(int64)

	return Cursor{
		Prev: encodeCursor(prev),
		Next: encodeCursor(next),

		IsFirst: next < int64(limit),
		IsLast:  last,
	}
}

// EncodedCursor is a scrambled string representation of a consecutive ordered integer cursor
type EncodedCursor string

func encodeCursor(value int64) EncodedCursor {
	encodedValue, err := cursorCodec.Encode([]uint64{uint64(value)})
	if err != nil {
		log.Printf("failed to encode cursor value %d, defaulting to unencoded value.", value)
		return EncodedCursor(strconv.FormatInt(value, 10))
	}
	return EncodedCursor(encodedValue)
}

func (c EncodedCursor) Decode() int64 {
	value := string(c)
	if value == "" {
		return 0
	}
	decodedValue := cursorCodec.Decode(value)
	if len(decodedValue) > 1 {
		log.Printf("encountered more than one cursor value after decoding: '%v', "+
			"this is not allowed! Defaulting to first value.", decodedValue)
	}
	if len(decodedValue) == 0 {
		log.Printf("decoding cursor value '%v' failed, defaulting to first page", decodedValue)
		return 0
	}
	return int64(decodedValue[0])
}
