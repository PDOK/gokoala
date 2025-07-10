package postgres

import "github.com/jackc/pgx/v5"

// PgxRowsAdapter implements domain.DatasourceRows
type PgxRowsAdapter struct {
	rows pgx.Rows
}

func FromPgxRows(rows pgx.Rows) *PgxRowsAdapter {
	return &PgxRowsAdapter{rows: rows}
}

func (p *PgxRowsAdapter) Columns() ([]string, error) {
	// pgx.Rows does not have a Columns() method like sqlx.Rows,
	// we need to get the field descriptions and extract names.
	fields := p.rows.FieldDescriptions()
	cols := make([]string, len(fields))
	for i, fd := range fields {
		cols[i] = fd.Name
	}
	return cols, nil
}

func (p *PgxRowsAdapter) SliceScan() ([]any, error) {
	values := make([]any, 0)
	err := p.rows.Scan(&values)
	return values, err
}

func (p *PgxRowsAdapter) Next() bool {
	return p.rows.Next()
}

func (p *PgxRowsAdapter) Err() error {
	return p.rows.Err()
}

func (p *PgxRowsAdapter) Close() {
	p.rows.Close()
}
