package datasources

import "github.com/paulmach/orb/geojson"

type GeoPackage struct {
}

func NewGeoPackage() *GeoPackage {
	return &GeoPackage{}
}

func (GeoPackage) GetFeatures(collection string) geojson.FeatureCollection {
	panic("not implemented yet")
}

func (GeoPackage) GetFeature(collection string, featureID string) geojson.Feature {
	panic("not implemented yet")
}
