package datasources

import (
	"log"

	"github.com/paulmach/orb/geojson"
)

type GeoPackage struct {
}

func NewGeoPackage() *GeoPackage {
	return &GeoPackage{}
}

func (GeoPackage) Close() {
	// TODO: clean up DB connection to gpkg
}

func (GeoPackage) GetFeatures(collection string) *geojson.FeatureCollection {
	// TODO: not implemented yet
	log.Printf("TODO: return data from gpkg for collection %s", collection)
	return nil
}

func (GeoPackage) GetFeature(collection string, featureID string) *geojson.Feature {
	// TODO: not implemented yet
	log.Printf("TODO: return feature %s from gpkg in collection %s", featureID, collection)
	return nil
}
