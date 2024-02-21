package geopackage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PDOK/gokoala/engine"
	"github.com/jmoiron/sqlx"
)

// assertIndexesExist asserts required indexes in the GeoPackage exists
func assertIndexesExist(
	configuredCollections engine.GeoSpatialCollections,
	featureTableByCollectionID map[string]*featureTable,
	db *sqlx.DB, fidColumn string) error {

	// index needs to contain these columns in the given order
	spatialBtreeColumns := strings.Join([]string{fidColumn, "minx", "maxx", "miny", "maxy"}, ",")

	for collID, table := range featureTableByCollectionID {
		if table == nil {
			return errors.New("given table can't be nil")
		}
		for _, coll := range configuredCollections {
			if coll.ID == collID && coll.Features != nil {
				// assert spatial b-tree index exists, this index substitutes the r-tree when querying large bounding boxes
				if err := assertIndexExists(table.TableName, db, spatialBtreeColumns); err != nil {
					return err
				}

				// assert temporal columns are indexed if configured
				if coll.Metadata.TemporalProperties != nil {
					temporalBtreeColumns := strings.Join([]string{coll.Metadata.TemporalProperties.StartDate, coll.Metadata.TemporalProperties.EndDate}, ",")
					if err := assertIndexExists(table.TableName, db, temporalBtreeColumns); err != nil {
						return err
					}
				}

				// assert the column for each property filter is indexed.
				for _, propertyFilter := range coll.Features.Filters.Properties {
					if err := assertIndexExists(table.TableName, db, propertyFilter.Name); err != nil {
						return err
					}
				}
				break
			}
		}
	}
	return nil
}

func assertIndexExists(tableName string, db *sqlx.DB, columns string) error {
	query := fmt.Sprintf(`
select group_concat(info.name) as indexed_columns
from pragma_index_list('%s') as list,
     pragma_index_info(list.name) as info
group by list.name`, tableName)

	rows, err := db.Queryx(query)
	if err != nil {
		return fmt.Errorf("failed to read indexes from table '%s'", tableName)
	}
	exists := false
	for rows.Next() {
		var indexedColumns string
		_ = rows.Scan(&indexedColumns)
		if columns == indexedColumns {
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
