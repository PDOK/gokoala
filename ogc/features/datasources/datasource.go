package datasources

import (
	"github.com/PDOK/gokoala/ogc/features/domain"
)

type Datasource interface {

	// GetFeatures returns a FeatureCollection from the underlying datasource
	GetFeatures(collection string) *domain.FeatureCollection

	// GetFeature returns a specific Feature from the FeatureCollection of the underlying datasource
	GetFeature(collection string, featureID string) *domain.Feature

	// Close closes (connections to) the datasource gracefully
	Close()
}
