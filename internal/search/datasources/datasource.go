package datasources

import (
	"context"

	"github.com/PDOK/gomagpie/internal/search/domain"
)

// Datasource knows how make different kinds of queries/actions on the underlying actual datastore.
// This abstraction allows the rest of the system to stay datastore agnostic.
type Datasource interface {
	Search(ctx context.Context, searchTerm string, collections CollectionsWithParams, srid domain.SRID, limit int) (*domain.FeatureCollection, error)

	// Close closes (connections to) the datasource gracefully
	Close()
}

// CollectionsWithParams collection name with associated CollectionParams
// These are provided though a URL query string as "deep object" params, e.g. paramName[prop1]=value1&paramName[prop2]=value2&....
type CollectionsWithParams map[string]CollectionParams

// CollectionParams parameter key with associated value
type CollectionParams map[string]string
