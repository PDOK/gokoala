package datasources

import "github.com/paulmach/orb/geojson"

type Datasource interface {

	// GetFeatures returns a FeatureCollection from the underlying datasource
	GetFeatures(collection string) *geojson.FeatureCollection

	// GetFeature returns a specific Feature from the FeatureCollection of the underlying datasource
	GetFeature(collection string, featureID string) *geojson.Feature

	// Close closes (connections to) the datasource gracefully
	Close()
}
