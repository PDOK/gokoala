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
	// we overwrite ID since we want to make it a required attribute. We also expect feature ids to be
	// auto-incrementing integers (which is the default in geopackages) since we use it for cursor-based pagination.
	ID    int64  `json:"id"`
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

func NewCursor(features []*Feature, limit int, last bool) Cursor {
	if len(features) == 0 {
		return Cursor{}
	}

	start := features[0].ID
	end := features[len(features)-1].ID

	prev := start
	if prev != 0 {
		prev -= int64(limit + 1)
		if prev < 0 {
			prev = 0
		}
	}
	next := end

	return Cursor{
		Prev: encodeCursor(prev),
		Next: encodeCursor(next),

		IsFirst: next <= int64(limit),
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
	} else if len(decodedValue) == 0 {
		log.Printf("decoding cursor value '%v' failed, defaulting to first page", decodedValue)
		return 0
	}
	result := int64(decodedValue[0])
	if result < 0 {
		result = 0
	}
	return result
}
