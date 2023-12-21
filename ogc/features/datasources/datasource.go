package datasources

import (
	"context"

	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/go-spatial/geom"
)

// Datasource holds all Features for a single object type in a specific projection.
type Datasource interface {

	// GetFeatureIDs returns all IDs of Features matching the given criteria, as well as and Cursors for pagination.
	// To be used in concert with GetFeaturesByID
	GetFeatureIDs(ctx context.Context, collection string, criteria FeaturesCriteria) ([]int64, domain.Cursors, error)

	// GetFeaturesByID returns a collection of Features with the given IDs. To be used in concert with GetFeatureIDs
	GetFeaturesByID(ctx context.Context, collection string, featureIDs []int64) (*domain.FeatureCollection, error)

	// GetFeatures returns all Features matching the given criteria and Cursors for pagination
	GetFeatures(ctx context.Context, collection string, criteria FeaturesCriteria) (*domain.FeatureCollection, domain.Cursors, error)

	// GetFeature returns a specific Feature
	GetFeature(ctx context.Context, collection string, featureID int64) (*domain.Feature, error)

	// GetFeatureTableMetadata returns metadata about a feature table associated with the given collection
	GetFeatureTableMetadata(collection string) (FeatureTableMetadata, error)

	// Close closes (connections to) the datasource gracefully
	Close()
}

// FeaturesCriteria to select a certain set of Features
type FeaturesCriteria struct {
	// pagination
	Cursor domain.DecodedCursor
	Limit  int

	// multiple projections support
	InputSRID  int // derived from bbox or filter param when available, or WGS84 as default
	OutputSRID int // derived from crs param when available, or WGS84 as default

	// filtering by bounding box
	Bbox *geom.Extent

	// filtering by CQL
	Filter     string
	FilterLang string
}

// FeatureTableMetadata abstraction to access metadata of a feature table (aka attribute table)
type FeatureTableMetadata interface {

	// ColumnsWithDataType returns a mapping from colum names to column data types.
	// Note: data types could be datasource specific.
	ColumnsWithDataType() map[string]string
}
