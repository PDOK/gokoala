package domain

import (
	"github.com/go-spatial/geom"
)

const (
	ConformanceJSONFGCore = "http://www.opengis.net/spec/json-fg-1/0.2/conf/core"
)

// featureType allows the type for Feature to be automatically set during json Marshalling
type featureType struct{}

func (ft *featureType) MarshalJSON() ([]byte, error) {
	return []byte(`"Feature"`), nil
}
func (ft *featureType) UnmarshalJSON([]byte) error { return nil }

type JSONFGFeatureCollection struct {
	Type           featureCollectionType `json:"type"`
	Timestamp      string                `json:"timeStamp,omitempty"`
	CoordRefSys    string                `json:"coordRefSys"`
	Links          []Link                `json:"links,omitempty"`
	ConformsTo     []string              `json:"conformsTo"`
	Features       []*JSONFGFeature      `json:"features"`
	NumberReturned int                   `json:"numberReturned"`
}

type JSONFGFeature struct {
	Type featureType `json:"type"`
	Time any         `json:"time"`
	// we don't implement the JSON-FG "3D" conformance class. So Place only
	// supports simple/2D geometries, no 3D geometries like Polyhedron, Prism, etc.
	Place       geom.Geometry  `json:"place"`    // may only contain non-WGS84 geometries
	Geometry    geom.Geometry  `json:"geometry"` // may only contain WGS84 geometries
	Properties  map[string]any `json:"properties"`
	CoordRefSys string         `json:"coordRefSys,omitempty"`
	Links       []Link         `json:"links,omitempty"`
	ConformsTo  []string       `json:"conformsTo,omitempty"`
	// We expect feature ids to be auto-incrementing integers (which is the default in geopackages)
	// since we use it for cursor-based pagination.
	ID int64 `json:"id"`
}
