package geopackage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/common"
	d "github.com/PDOK/gokoala/internal/ogc/features/domain"

	"github.com/jmoiron/sqlx"
)

// assertIndexesExist asserts required indexes in the GeoPackage exists
func assertIndexesExist(
	configuredCollections config.GeoSpatialCollections,
	featureTableByCollectionID map[string]*common.FeatureTable,
	db *sqlx.DB, fidColumn string) error {

	// index needs to contain these columns in the given order
	defaultSpatialBtreeColumns := strings.Join([]string{fidColumn, d.MinxField, d.MaxxField, d.MinyField, d.MaxyField}, ",")

	for collID, table := range featureTableByCollectionID {
		if table == nil {
			return errors.New("given table can't be nil")
		}
		for _, coll := range configuredCollections {
			if coll.ID == collID && coll.Features != nil {
				spatialBtreeColumns := defaultSpatialBtreeColumns

				// assert temporal columns are indexed if configured
				if coll.Metadata != nil && coll.Metadata.TemporalProperties != nil {
					temporalBtreeColumns := strings.Join([]string{coll.Metadata.TemporalProperties.StartDate, coll.Metadata.TemporalProperties.EndDate}, ",")
					spatialBtreeColumns = strings.Join([]string{defaultSpatialBtreeColumns, coll.Metadata.TemporalProperties.StartDate, coll.Metadata.TemporalProperties.EndDate}, ",")
					if err := assertIndexExists(table.TableName, db, temporalBtreeColumns, true, false); err != nil {
						return err
					}
				}

				// assert spatial b-tree index exists, this index substitutes the r-tree when querying large bounding boxes
				// if temporal columns are configured, they must be included in this index as well
				if err := assertIndexExists(table.TableName, db, spatialBtreeColumns, true, false); err != nil {
					return err
				}

				// assert the column for each property filter is indexed.
				for _, propertyFilter := range coll.Features.Filters.Properties {
					if err := assertIndexExists(table.TableName, db, propertyFilter.Name, false, true); err != nil && *propertyFilter.IndexRequired {
						return fmt.Errorf("%w. To disable this check set 'indexRequired' to 'false'", err)
					}
				}
				break
			}
		}
	}
	return nil
}

func assertIndexExists(tableName string, db *sqlx.DB, columns string, prefixMatch bool, containsMatch bool) error {
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
