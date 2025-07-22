package common

// DatasourceRows defines an abstraction over rows/records retrieved from a datasource.
// Can be implemented using libraries such as jackc/pgx, jmoiron/sqlx, database/sql, etc.
type DatasourceRows interface {
	// Columns provided the column names
	Columns() ([]string, error)

	// SliceScan scans the current row into a slice of any
	SliceScan() ([]any, error)

	// Next advances the row pointer to the next row
	Next() bool

	// Err any error that occurred during iteration
	Err() error

	// Close closes the row set, releasing any resources
	Close()
}
