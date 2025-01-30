package datasources

import (
	"context"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	geom2 "github.com/twpayne/go-geom"
)

// Datasource holds all Features for a single object type in a specific projection/CRS.
// This abstraction allows the rest of the system to stay datastore agnostic.
type Datasource interface {

	// GetFeatureIDs returns all IDs of Features matching the given criteria, as well as Cursors for pagination.
	// To be used in concert with GetFeaturesByID
	GetFeatureIDs(ctx context.Context, collection string, criteria FeaturesCriteria) ([]int64, domain.Cursors, error)

	// GetFeaturesByID returns a collection of Features with the given IDs. To be used in concert with GetFeatureIDs
	GetFeaturesByID(ctx context.Context, collection string, featureIDs []int64, profile domain.Profile) (*domain.FeatureCollection, error)

	// GetFeatures returns all Features matching the given criteria and Cursors for pagination
	GetFeatures(ctx context.Context, collection string, criteria FeaturesCriteria, profile domain.Profile) (*domain.FeatureCollection, domain.Cursors, error)

	// GetFeature returns a specific Feature, based on its feature id
	GetFeature(ctx context.Context, collection string, featureID any, profile domain.Profile) (*domain.Feature, error)

	// GetFeatureTableMetadata returns metadata about a feature table associated with the given collection
	GetFeatureTableMetadata(collection string) (FeatureTableMetadata, error)

	// GetPropertyFiltersWithAllowedValues returns configured property filters for the given collection enriched with allowed values.
	// When enrichments don't apply the returned result should still contain all property filters as specified in the (YAML) config.
	GetPropertyFiltersWithAllowedValues(collection string) PropertyFiltersWithAllowedValues

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
	Bbox *geom2.Bounds

	// filtering by reference date/time
	TemporalCriteria TemporalCriteria

	// filtering by properties (OAF part 1)
	PropertyFilters map[string]string

	// filtering by CQL (OAF part 3)
	Filter     string
	FilterLang string
}

// TemporalCriteria criteria to filter based on date/time
type TemporalCriteria struct {
	// reference date
	ReferenceDate time.Time

	// startDate and endDate properties
	StartDateProperty string
	EndDateProperty   string
}

// FeatureTableMetadata abstraction to access metadata of a feature table (aka attribute table)
type FeatureTableMetadata interface {

	// ColumnsWithDataType returns a mapping from column names to column data types.
	// Note: data types can be datasource specific.
	ColumnsWithDataType() map[string]string
}

// PropertyFilterWithAllowedValues property filter as configured in the (YAML) config, but enriched with allowed values
type PropertyFilterWithAllowedValues struct {
	config.PropertyFilter

	// static or dynamic values that are allowed to be used in this property filter
	AllowedValues []string
}

// PropertyFiltersWithAllowedValues one or more PropertyFilterWithAllowedValues indexed by property filter name
type PropertyFiltersWithAllowedValues map[string]PropertyFilterWithAllowedValues
