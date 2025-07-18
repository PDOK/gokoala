package geopackage

import "github.com/jmoiron/sqlx"

// SqlxRowsAdapter implements domain.DatasourceRows
type SqlxRowsAdapter struct {
	rows *sqlx.Rows
}

func FromSqlxRows(rows *sqlx.Rows) *SqlxRowsAdapter {
	return &SqlxRowsAdapter{rows: rows}
}

func (s *SqlxRowsAdapter) Columns() ([]string, error) {
	return s.rows.Columns()
}

func (s *SqlxRowsAdapter) SliceScan() ([]any, error) {
	return s.rows.SliceScan()
}

func (s *SqlxRowsAdapter) Next() bool {
	return s.rows.Next()
}

func (s *SqlxRowsAdapter) Err() error {
	return s.rows.Err()
}

func (s *SqlxRowsAdapter) Close() {
	_ = s.rows.Close()
}
