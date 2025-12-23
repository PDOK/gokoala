package datasources

import (
	"context"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/twpayne/go-geom"
)

// Datasource holds all Features for a single object type in a specific projection/CRS.
// This abstraction allows the rest of the system to stay datastore agnostic.
type Datasource interface {

	// GetFeatureIDs returns all IDs of Features matching the given criteria, as well as Cursors for pagination.
	// To be used in concert with GetFeaturesByID
	GetFeatureIDs(ctx context.Context, collection string, criteria FeaturesCriteria) ([]int64, domain.Cursors, error)

	// GetFeaturesByID returns a collection of Features with the given IDs. To be used in concert with GetFeatureIDs
	GetFeaturesByID(ctx context.Context, collection string, featureIDs []int64, axisOrder domain.AxisOrder, profile domain.Profile) (*domain.FeatureCollection, error)

	// GetFeatures returns all Features matching the given criteria and Cursors for pagination
	GetFeatures(ctx context.Context, collection string, criteria FeaturesCriteria, axisOrder domain.AxisOrder, profile domain.Profile) (*domain.FeatureCollection, domain.Cursors, error)

	// GetFeature returns a specific Feature, based on its feature id
	GetFeature(ctx context.Context, collection string, featureID any, outputSRID domain.SRID, axisOrder domain.AxisOrder, profile domain.Profile) (*domain.Feature, error)

	// SearchFeaturesAcrossCollections search features in one or more collections. Collections can be located
	// in this dataset or in other datasets.
	SearchFeaturesAcrossCollections(ctx context.Context, searchQuery domain.SearchQuery, collections domain.CollectionsWithParams,
		srid domain.SRID, bbox *geom.Bounds, bboxSRID domain.SRID, limit int) (*domain.FeatureCollection, error)

	// GetSchema returns the schema (fields, data types, descriptions, etc.) of the table associated with the given collection
	GetSchema(collection string) (*domain.Schema, error)

	// GetPropertyFiltersWithAllowedValues returns configured property filters for the given collection enriched with allowed values.
	// When enrichments don't apply, the returned result should still contain all property filters as specified in the (YAML) config.
	GetPropertyFiltersWithAllowedValues(collection string) PropertyFiltersWithAllowedValues

	// GetCollectionType returns the type of data in the given collection, e.g. 'features' or 'attributes'.
	GetCollectionType(collection string) (geospatial.CollectionType, error)

	// SupportsOnTheFlyTransformation returns whether the datasource supports coordinate transformation/reprojection on-the-fly
	SupportsOnTheFlyTransformation() bool

	// Close closes (connections to) the datasource gracefully
	Close()
}

// FeaturesCriteria to select a certain set of Features.
type FeaturesCriteria struct {
	// pagination (OAF part 1)
	Cursor domain.DecodedCursor
	Limit  int

	// multiple projections support (OAF part 2)
	InputSRID  domain.SRID // derived from bbox or filter param when available, or WGS84 as default
	OutputSRID domain.SRID // derived from crs param when available, or WGS84 as default

	// filtering by bounding box (OAF part 1)
	Bbox *geom.Bounds

	// filtering by reference date/time (OAF part 1)
	TemporalCriteria TemporalCriteria

	// filtering by properties (OAF part 1)
	PropertyFilters map[string]string

	// filtering by CQL (OAF part 3)
	Filter     string
	FilterLang string
}

// TemporalCriteria criteria to filter based on date/time.
type TemporalCriteria struct {
	// reference date
	ReferenceDate time.Time

	// startDate and endDate properties
	StartDateProperty string
	EndDateProperty   string
}

// PropertyFilterWithAllowedValues property filter as configured in the (YAML) config, but enriched with allowed values.
type PropertyFilterWithAllowedValues struct {
	config.PropertyFilter

	// static or dynamic values that are allowed to be used in this property filter
	AllowedValues []string
}

// PropertyFiltersWithAllowedValues one or more PropertyFilterWithAllowedValues indexed by property filter name.
type PropertyFiltersWithAllowedValues map[string]PropertyFilterWithAllowedValues
