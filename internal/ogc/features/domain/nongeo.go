package domain

import "github.com/twpayne/go-geom/encoding/geojson"

// -----
// "NonGeo JSON" is NOT a formal standard. It's defined by PDOK as a way to return non-spatial data from certain collections
// alongside collections that do contain spatial data. NonGeo JSON is modeled after GeoJSON but does NOT contain a geometry.
// -----

// NonGeoCollection is a FeatureCollection with only attributes and NO geometries.
type NonGeoCollection struct {
	Features []*NonGeo `json:"features"`
	FeatureCollection
}

// NonGeo is a Feature with only attributes and NO geometry.
type NonGeo struct {
	Type       featureType       `json:"type"`
	Properties FeatureProperties `json:"properties"`
	// We support 'null' geometries, don't add an 'omitempty' tag here.
	Geometry *geojson.Geometry `json:"geometry"`
	// We expect ids to be auto-incrementing integers (which is the default in geopackages)
	// since we use it for cursor-based pagination.
	ID    string `json:"id"`
	Links []Link `json:"links,omitempty"`
}

// Keys of the NonGeo properties.
func (f *NonGeo) Keys() []string {
	return f.Properties.Keys()
}
