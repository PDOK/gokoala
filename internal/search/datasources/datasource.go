package datasources

import (
	"context"
)

// Datasource knows how make different kinds of queries/actions on the underlying actual datastore.
// This abstraction allows the rest of the system to stay datastore agnostic.
type Datasource interface {
	Suggest(ctx context.Context, suggestForThis string) ([]string, error)

	// Close closes (connections to) the datasource gracefully
	Close()
}
