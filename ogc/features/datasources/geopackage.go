package datasources

import (
	"log"

	"github.com/PDOK/gokoala/ogc/features/domain"
)

type GeoPackage struct {
}

func NewGeoPackage() *GeoPackage {
	return &GeoPackage{}
}

func (GeoPackage) Close() {
	// TODO: clean up DB connection to gpkg
}

func (GeoPackage) GetFeatures(collection string, cursor string, limit int) (*domain.FeatureCollection, domain.Cursor) {
	// TODO: not implemented yet
	log.Printf("TODO: return data from gpkg for collection %s using cursor %s with limt %d",
		collection, cursor, limit)
	return nil, domain.Cursor{}
}

func (GeoPackage) GetFeature(collection string, featureID string) *domain.Feature {
	// TODO: not implemented yet
	log.Printf("TODO: return feature %s from gpkg in collection %s", featureID, collection)
	return nil
}
