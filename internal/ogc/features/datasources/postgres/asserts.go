package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/common"
	"github.com/jackc/pgx/v5/pgxpool"
)

// assertIndexesExist asserts required indexes in Postgres exists
//
//nolint:nestif
func assertIndexesExist(configuredCollections config.FeaturesCollections, tableByCollectionID map[string]*common.Table, db *pgxpool.Pool, spatialIndexRequired bool) error {

	for collID, table := range tableByCollectionID {
		if table == nil {
			return errors.New("given table can't be nil")
		}
		for _, coll := range configuredCollections {
			if coll.GetID() == collID {
				// assert temporal columns are indexed if configured
				if coll.GetMetadata() != nil && coll.GetMetadata().TemporalProperties != nil {
					temporalColumns := strings.Join([]string{coll.GetMetadata().TemporalProperties.StartDate, coll.GetMetadata().TemporalProperties.EndDate}, ",")
					if err := assertIndexExists(table.Name, db, temporalColumns, true, false); err != nil {
						return err
					}
				}

				// assert geometry column has GIST (rtree) index
				if spatialIndexRequired {
					if err := assertSpatialIndex(table.Name, db, table.GeometryColumnName); err != nil {
						return err
					}
				}

				// assert the column for each property filter is indexed
				for _, propertyFilter := range coll.Filters.Properties {
					if err := assertIndexExists(table.Name, db, propertyFilter.Name, false, true); err != nil && *propertyFilter.IndexRequired {
						return fmt.Errorf("%w. To disable this check set 'indexRequired' to 'false'", err)
					}
				}

				break
			}
		}
	}

	return nil
}

func assertSpatialIndex(tableName string, db *pgxpool.Pool, geometryColumn string) error {
	query := `
select count(*)
from pg_index idx
join pg_class tbl on tbl.oid = idx.indrelid
join pg_class idx_class on idx_class.oid = idx.indexrelid
join pg_am am on idx_class.relam = am.oid
join pg_attribute attr on attr.attrelid = tbl.oid
where tbl.relname = $1
  and am.amname = 'gist'
  and attr.attnum = any(idx.indkey)
  and attr.attname = $2;`

	var count int
	err := db.QueryRow(context.Background(), query, tableName, geometryColumn).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check spatial index on table '%s', column '%s': %w",
			tableName, geometryColumn, err)
	}
	if count == 0 {
		return fmt.Errorf("missing required spatial index (GIST): no index exists on geometry column '%s' in table '%s'",
			geometryColumn, tableName)
	}

	return nil
}

func assertIndexExists(tableName string, db *pgxpool.Pool, columns string, prefixMatch bool, containsMatch bool) error {
	query := `
select string_agg(a.attname, ',' order by array_position(idx.indkey, a.attnum)) as indexed_columns
from pg_class c
join pg_index idx on c.oid = idx.indrelid
join pg_attribute a on a.attrelid = c.oid and a.attnum = any(idx.indkey)
where c.relname = $1
group by idx.indexrelid;`

	rows, err := db.Query(context.Background(), query, tableName)
	if err != nil {
		return fmt.Errorf("failed to read indexes from table '%s'", tableName)
	}
	exists := false
	for rows.Next() {
		var indexedColumns string
		_ = rows.Scan(&indexedColumns)
		switch {
		case columns == indexedColumns:
			exists = true
		case prefixMatch && strings.HasPrefix(indexedColumns, columns):
			exists = true
		case containsMatch && strings.Contains(indexedColumns, columns):
			exists = true
		}
	}
	defer rows.Close()
	if !exists {
		return fmt.Errorf("missing required index: no index exists on column(s) '%s' in table '%s'",
			columns, tableName)
	}

	return nil
}
