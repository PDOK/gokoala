package datasources

import (
	"github.com/PDOK/gokoala/ogc/features/domain"
)

// Datasource holding all the features for a single dataset
type Datasource interface {

	// GetFeatures returns a FeatureCollection from the underlying datasource and a Cursor for pagination
	GetFeatures(collection string, cursor string, limit int) (*domain.FeatureCollection, domain.Cursor)

	// GetFeature returns a specific Feature from the FeatureCollection of the underlying datasource
	GetFeature(collection string, featureID string) *domain.Feature

	// Close closes (connections to) the datasource gracefully
	Close()
}
