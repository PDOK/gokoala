package datasources

import "github.com/paulmach/orb/geojson"

type Datasource interface {
	GetFeatures(collection string) geojson.FeatureCollection

	GetFeature(collection string, featureID string) geojson.Feature
}
