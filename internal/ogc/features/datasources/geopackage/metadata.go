package geopackage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/PDOK/gokoala/config"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	d "github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/jmoiron/sqlx"
)

// featureTable according to spec https://www.geopackage.org/spec121/index.html#_contents
type featureTable struct {
	TableName          string          `db:"table_name"`
	DataType           string          `db:"data_type"` // always 'features'
	Identifier         string          `db:"identifier"`
	Description        string          `db:"description"`
	GeometryColumnName string          `db:"column_name"`
	GeometryType       string          `db:"geometry_type_name"`
	LastChange         time.Time       `db:"last_change"`
	MinX               sql.NullFloat64 `db:"min_x"` // bbox
	MinY               sql.NullFloat64 `db:"min_y"` // bbox
	MaxX               sql.NullFloat64 `db:"max_x"` // bbox
	MaxY               sql.NullFloat64 `db:"max_y"` // bbox
	SRS                sql.NullInt64   `db:"srs_id"`

	Schema *d.Schema // required
}

// readMetadata reads metadata such as available feature tables, the schema of each table,
// available filters, etc. from the GeoPackage. Terminates on failure.
func readMetadata(db *sqlx.DB, collections config.GeoSpatialCollections, fidColumn, externalFidColumn string) (
	featureTableByCollectionID map[string]*featureTable,
	propertyFiltersByCollectionID map[string]ds.PropertyFiltersWithAllowedValues) {

	metadata, err := readDriverMetadata(db)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(metadata)

	featureTableByCollectionID, err = readGpkgContents(collections, db, fidColumn, externalFidColumn)
	if err != nil {
		log.Fatal(err)
	}
	propertyFiltersByCollectionID, err = readPropertyFiltersWithAllowedValues(featureTableByCollectionID, collections, db)
	if err != nil {
		log.Fatal(err)
	}
	return
}

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
		return "", fmt.Errorf("failed to connect with GeoPackage: %w", err)
	}

	var gpkgVersion pragma
	_ = db.QueryRowx(`pragma user_version`).StructScan(&gpkgVersion)
	if gpkgVersion.UserVersion == "" {
		gpkgVersion.UserVersion = "unknown"
	}

	return fmt.Sprintf("geopackage version: %s, sqlite version: %s, spatialite version: %s on %s",
		gpkgVersion.UserVersion, m.Sqlite, m.Spatialite, m.Arch), nil
}

// Read "gpkg_contents" table. This table contains metadata about feature tables. The result is a mapping from
// collection ID -> feature table metadata. We match each feature table to the collection ID by looking at the
// 'table_name' column. Also, in case there's no exact match between 'collection ID' and 'table_name' we use
// the explicitly configured table name (from the YAML config).
func readGpkgContents(collections config.GeoSpatialCollections, db *sqlx.DB,
	fidColumn, externalFidColumn string) (map[string]*featureTable, error) {

	query := `
select
	c.table_name, c.data_type, c.identifier, c.description, c.last_change,
	c.min_x, c.min_y, c.max_x, c.max_y, c.srs_id, gc.column_name, gc.geometry_type_name
from
	gpkg_contents c join gpkg_geometry_columns gc on c.table_name == gc.table_name
where
	c.data_type = 'features'`

	rows, err := db.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve gpkg_contents using query: %v\n, error: %w", query, err)
	}
	defer rows.Close()

	result := make(map[string]*featureTable, 10)
	for rows.Next() {
		row := featureTable{}
		if err = rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("failed to read gpkg_contents record, error: %w", err)
		}
		if row.TableName == "" {
			return nil, fmt.Errorf("feature table name is blank, error: %w", err)
		}
		row.Schema, err = readSchema(db, row, fidColumn, externalFidColumn)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema for table %s, error: %w", row.TableName, err)
		}

		for _, collection := range collections {
			if row.TableName == collection.ID {
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
		return nil, errors.New("no records for 'features' found in gpkg_contents and/or gpkg_geometry_columns")
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

func readPropertyFiltersWithAllowedValues(featTableByCollection map[string]*featureTable,
	collections config.GeoSpatialCollections, db *sqlx.DB) (map[string]ds.PropertyFiltersWithAllowedValues, error) {

	result := make(map[string]ds.PropertyFiltersWithAllowedValues)
	for _, collection := range collections {
		if collection.Features == nil {
			continue
		}
		result[collection.ID] = make(map[string]ds.PropertyFilterWithAllowedValues)
		featTable := featTableByCollection[collection.ID]

		for _, pf := range collection.Features.Filters.Properties {
			// the result should contain ALL configured property filters, with or without allowed values.
			// when available, allowed values can be either static (from YAML config) or derived from the geopackage
			result[collection.ID][pf.Name] = ds.PropertyFilterWithAllowedValues{PropertyFilter: pf}
			if pf.AllowedValues != nil {
				result[collection.ID][pf.Name] = ds.PropertyFilterWithAllowedValues{PropertyFilter: pf, AllowedValues: pf.AllowedValues}
				continue
			}
			if *pf.DeriveAllowedValuesFromDatasource {
				if !*pf.IndexRequired {
					log.Printf("Warning: index is disabled for column %s, deriving allowed values "+
						"from may take a long time. Index on this column is recommended", pf.Name)
				}
				// select distinct values from given column
				query := fmt.Sprintf("select distinct ft.%s from %s ft", pf.Name, featTable.TableName)
				var values []string
				err := db.Select(&values, query)
				if err != nil {
					return nil, fmt.Errorf("failed to derive allowed values using query: %v\n, error: %w", query, err)
				}
				result[collection.ID][pf.Name] = ds.PropertyFilterWithAllowedValues{PropertyFilter: pf, AllowedValues: values}
				continue
			}
		}
	}
	return result, nil
}

func readSchema(db *sqlx.DB, table featureTable, fidColumn, externalFidColumn string) (*d.Schema, error) {
	rows, err := db.Queryx(fmt.Sprintf("select name, type, \"notnull\" from pragma_table_info('%s')", table.TableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fields := make([]d.Field, 0)
	for rows.Next() {
		var colName, colType, colNotNull string
		err = rows.Scan(&colName, &colType, &colNotNull)
		if err != nil {
			return nil, err
		}
		fields = append(fields, d.Field{
			Name:              colName,
			Type:              colType,
			IsRequired:        colNotNull == "1",
			IsPrimaryGeometry: colName == table.GeometryColumnName,
			FeatureRelation:   d.NewFeatureRelation(colName, externalFidColumn),
		})
	}
	schema, err := d.NewSchema(fields, fidColumn, externalFidColumn)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func hasMatchingTableName(collection config.GeoSpatialCollection, row featureTable) bool {
	return collection.Features != nil && collection.Features.TableName != nil &&
		row.TableName == *collection.Features.TableName
}
