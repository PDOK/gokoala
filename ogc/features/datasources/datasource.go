package datasources

import (
	"context"

	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/go-spatial/geom"
)

// Datasource holding all the features for a single dataset
type Datasource interface {

	// GetFeatures returns a FeatureCollection from the underlying datasource and Cursors for pagination
	GetFeatures(ctx context.Context, collection string, options FeaturesOptions) (*FeaturesResult, error)

	// GetFeatures returns a FeatureCollection from the underlying datasource and Cursors for pagination
	GetFeaturesByID(ctx context.Context, collection string, featureIDs []int64) (*domain.FeatureCollection, error)

	// GetFeature returns a specific Feature from the FeatureCollection of the underlying datasource
	GetFeature(ctx context.Context, collection string, featureID int64) (*domain.Feature, error)

	// Close closes (connections to) the datasource gracefully
	Close()
}

// FeaturesOptions to select a certain set of Features
type FeaturesOptions struct {
	// pagination
	Cursor domain.DecodedCursor
	Limit  int

	// multiple projections support
	Crs int

	// filtering by bounding box
	Bbox    *geom.Extent
	BboxCrs int

	// filtering by CQL
	Filter    string
	FilterCrs string
}

func (fo *FeaturesOptions) SelectOnlyFeatureIDs() bool {
	return fo.BboxCrs != fo.Crs
}

type FeaturesResult struct {
	Collection *domain.FeatureCollection
	Cursors    domain.Cursors
	FeatureIDs []int64
}
