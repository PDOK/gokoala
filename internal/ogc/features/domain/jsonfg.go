package domain

import (
	geojson2 "github.com/twpayne/go-geom/encoding/geojson"
)

const (
	ConformanceJSONFGCore = "http://www.opengis.net/spec/json-fg-1/0.2/conf/core"
)

// JSONFGFeatureCollection FeatureCollection according to the JSON-FG standard
// Note: fields in this struct are sorted for optimal memory usage (field alignment)
type JSONFGFeatureCollection struct {
	Type           featureCollectionType `json:"type"`
	Timestamp      string                `json:"timeStamp,omitempty"`
	CoordRefSys    string                `json:"coordRefSys"`
	Links          []Link                `json:"links,omitempty"`
	ConformsTo     []string              `json:"conformsTo"`
	Features       []*JSONFGFeature      `json:"features"`
	NumberReturned int                   `json:"numberReturned"`
}

// JSONFGFeature Feature according to the JSON-FG standard
// Note: fields in this struct are sorted for optimal memory usage (field alignment)
type JSONFGFeature struct {
	// We expect feature ids to be auto-incrementing integers (which is the default in geopackages)
	// since we use it for cursor-based pagination.
	ID   string      `json:"id"`
	Type featureType `json:"type"`
	Time any         `json:"time"`
	// We don't implement the JSON-FG "3D" conformance class. So Place only
	// supports simple/2D geometries, no 3D geometries like Polyhedron, Prism, etc.
	Place       *geojson2.Geometry `json:"place"`    // may only contain non-WGS84 geometries
	Geometry    *geojson2.Geometry `json:"geometry"` // may only contain WGS84 geometries
	Properties  FeatureProperties  `json:"properties"`
	CoordRefSys string             `json:"coordRefSys,omitempty"`
	Links       []Link             `json:"links,omitempty"`
	ConformsTo  []string           `json:"conformsTo,omitempty"`
}
