package domain

import "github.com/go-spatial/geom"

// featureType allows the type for Feature to be automatically set during json Marshalling
type featureType struct{}

func (ft *featureType) MarshalJSON() ([]byte, error) {
	return []byte(`"Feature"`), nil
}
func (ft *featureType) UnmarshalJSON([]byte) error { return nil }

type JSONFGFeatureCollection struct {
	Links []Link `json:"links,omitempty"`

	NumberReturned int                   `json:"numberReturned"`
	Type           featureCollectionType `json:"type"`
	Features       []*JSONFGFeature      `json:"features"`
}

type JSONFGFeature struct {
	// we overwrite ID since we want to make it a required attribute. We also expect feature ids to be
	// auto-incrementing integers (which is the default in geopackages) since we use it for cursor-based pagination.
	ID    int64  `json:"id"`
	Links []Link `json:"links,omitempty"`

	Type       featureType            `json:"type"`
	Place      geom.Geometry          `json:"place"`
	Geometry   geom.Geometry          `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}
