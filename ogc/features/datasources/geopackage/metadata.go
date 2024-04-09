package geopackage

import (
	"errors"
	"fmt"
	"log"

	"github.com/PDOK/gokoala/config"

	"github.com/jmoiron/sqlx"
)

// Read metadata about gpkg and sqlite driver
func readDriverMetadata(db *sqlx.DB) (string, error) {
	type pragma struct {
		UserVersion string `db:"user_version"`
	}
	type metadata struct {
		Sqlite     string `db:"sqlite"`
		Spatialite string `db:"spatialite"`
		Arch       string `db:"arch"`
	}

	var m metadata
	err := db.QueryRowx(`
select sqlite_version() as sqlite,
spatialite_version() as spatialite,
spatialite_target_cpu() as arch`).StructScan(&m)
	if err != nil {
		return "", err
	}

	var gpkgVersion pragma
	_ = db.QueryRowx(`pragma user_version`).StructScan(&gpkgVersion)
	if gpkgVersion.UserVersion == "" {
		gpkgVersion.UserVersion = "unknown"
	}

	return fmt.Sprintf("geopackage version: %s, sqlite version: %s, spatialite version: %s on %s",
		gpkgVersion.UserVersion, m.Sqlite, m.Spatialite, m.Arch), nil
}

// Read gpkg_contents table. This table contains metadata about feature tables. The result is a mapping from
// collection ID -> feature table metadata. We match each feature table to the collection ID by looking at the
// 'identifier' column. Also in case there's no exact match between 'collection ID' and 'identifier' we use
// the explicitly configured table name.
func readGpkgContents(collections config.GeoSpatialCollections, db *sqlx.DB) (map[string]*featureTable, error) {
	query := `
select
	c.table_name, c.data_type, c.identifier, c.description, c.last_change,
	c.min_x, c.min_y, c.max_x, c.max_y, c.srs_id, gc.column_name, gc.geometry_type_name
from
	gpkg_contents c join gpkg_geometry_columns gc on c.table_name == gc.table_name
where
	c.data_type = 'features' and
	c.min_x is not null`

	rows, err := db.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve gpkg_contents using query: %v\n, error: %w", query, err)
	}
	defer rows.Close()

	result := make(map[string]*featureTable, 10)
	for rows.Next() {
		row := featureTable{
			ColumnsWithDateType: make(map[string]string),
		}
		if err = rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("failed to read gpkg_contents record, error: %w", err)
		}
		if row.TableName == "" {
			return nil, fmt.Errorf("feature table name is blank, error: %w", err)
		}
		if err = readFeatureTableInfo(db, row); err != nil {
			return nil, fmt.Errorf("failed to read feature table metadata, error: %w", err)
		}

		for _, collection := range collections {
			if row.Identifier == collection.ID {
				result[collection.ID] = &row
			} else if hasMatchingTableName(collection, row) {
				result[collection.ID] = &row
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, errors.New("no records found in gpkg_contents, can't serve features")
	}
	uniqueTables := make(map[string]struct{})
	for _, table := range result {
		uniqueTables[table.TableName] = struct{}{}
	}
	if len(uniqueTables) != len(result) {
		log.Printf("Warning: found %d unique table names for %d collections, "+
			"usually each collection is backed by its own unique table\n", len(uniqueTables), len(result))
	}
	return result, nil
}

func readFeatureTableInfo(db *sqlx.DB, table featureTable) error {
	rows, err := db.Queryx(fmt.Sprintf("select name, type from pragma_table_info('%s')", table.TableName))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var colName, colType string
		err = rows.Scan(&colName, &colType)
		if err != nil {
			return err
		}
		table.ColumnsWithDateType[colName] = colType
	}
	return nil
}

func hasMatchingTableName(collection config.GeoSpatialCollection, row featureTable) bool {
	return collection.Features != nil && collection.Features.TableName != nil &&
		row.Identifier == *collection.Features.TableName
}
