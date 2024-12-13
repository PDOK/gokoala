package datasources

import (
	"context"

	"github.com/PDOK/gomagpie/internal/search/domain"
)

// Datasource knows how make different kinds of queries/actions on the underlying actual datastore.
// This abstraction allows the rest of the system to stay datastore agnostic.
type Datasource interface {
	Search(ctx context.Context, searchTerm string, collections domain.CollectionsWithParams,
		srid domain.SRID, limit int) (*domain.FeatureCollection, error)

	// Close closes (connections to) the datasource gracefully
	Close()
}
