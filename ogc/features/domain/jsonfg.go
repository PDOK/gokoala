package domain

import "github.com/go-spatial/geom"

// featureType allows the type for Feature to be automatically set during json Marshalling
type featureType struct{}

func (ft *featureType) MarshalJSON() ([]byte, error) {
	return []byte(`"Feature"`), nil
}
func (ft *featureType) UnmarshalJSON([]byte) error { return nil }

// conformsTo allows the JSON-FG conformance to be automatically set during json Marshalling
type conformsTo struct{}

func (ct *conformsTo) MarshalJSON() ([]byte, error) {
	return []byte("[\"http://www.opengis.net/spec/json-fg-1/0.2/conf/core\"]"), nil
}
func (ct *conformsTo) UnmarshalJSON([]byte) error { return nil }

type JSONFGFeatureCollection struct {
	Links          []Link                `json:"links,omitempty"`
	NumberReturned int                   `json:"numberReturned"`
	Type           featureCollectionType `json:"type"`
	ConformsTo     conformsTo            `json:"conformsTo"`
	Features       []*JSONFGFeature      `json:"features"`
}

type JSONFGFeature struct {
	// we overwrite ID since we want to make it a required attribute. We also expect feature ids to be
	// auto-incrementing integers (which is the default in geopackages) since we use it for cursor-based pagination.
	ID         int64       `json:"id"`
	Links      []Link      `json:"links,omitempty"`
	Type       featureType `json:"type"`
	ConformsTo conformsTo  `json:"conformsTo"`
	Time       any         `json:"time"`
	// we don't implement the JSON-FG "3D" conformance class. So Place only
	// supports simple/2D geometries, no 3D geometries like Polyhedron, Prism, etc.
	Place      geom.Geometry          `json:"place"`    // may only contain non-WGS84 geometries
	Geometry   geom.Geometry          `json:"geometry"` // may only contain WGS84 geometries
	Properties map[string]interface{} `json:"properties"`
}
