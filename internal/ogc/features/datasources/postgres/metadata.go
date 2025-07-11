package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/PDOK/gokoala/config"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	d "github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

var newlineRegex = regexp.MustCompile(`[\r\n]+`)

// featureTable Postgres table holding features
type featureTable struct {
	TableName          string
	GeometryColumnName string
	GeometryType       string

	Schema *d.Schema // required
}

// readMetadata reads metadata such as available feature tables, the schema of each table,
// available filters, etc. from the Postgres database. Terminates on failure.
func readMetadata(db *pgxpool.Pool, collections config.GeoSpatialCollections, fidColumn, externalFidColumn, schemaName string) (
	featureTableByCollectionID map[string]*featureTable,
	propertyFiltersByCollectionID map[string]ds.PropertyFiltersWithAllowedValues) {

	metadata, err := readDriverMetadata(db)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(metadata)

	featureTableByCollectionID, err = readFeatureTables(collections, db, fidColumn, externalFidColumn, schemaName)
	if err != nil {
		log.Fatal(err)
	}
	propertyFiltersByCollectionID, err = readPropertyFiltersWithAllowedValues(featureTableByCollectionID, collections, db)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// Read metadata about PostgreSQL and PostGIS
func readDriverMetadata(db *pgxpool.Pool) (string, error) {
	var pgVersion string
	var postGISVersion string

	err := db.QueryRow(context.Background(), `
		select (select version()) as pg_version, (select PostGIS_Version()) as postgis_version;
	`).Scan(&pgVersion, &postGISVersion)

	return fmt.Sprintf("postgresql version: '%s', postgis version: '%s'", pgVersion, postGISVersion), err
}

// Read "geometry_columns" view. This table contains metadata about PostGIS tables. The result is a mapping from
// collection ID -> feature table metadata. We match each feature table to the collection ID by looking at the
// 'f_table_name' column. Also, in case there's no exact match between 'collection ID' and 'f_table_name' we use
// the explicitly configured table name (from the YAML config).
func readFeatureTables(collections config.GeoSpatialCollections, db *pgxpool.Pool,
	fidColumn, externalFidColumn, schemaName string) (map[string]*featureTable, error) {

	query := `
select
	f_table_name::text, f_geometry_column::text, type::text
from
	geometry_columns
where
	f_table_schema = $1`

	rows, err := db.Query(context.Background(), query, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve geometry_columns using query: %v\n, error: %w", query, err)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*featureTable, 10)
	for rows.Next() {
		table := featureTable{}
		if err = rows.Scan(&table.TableName, &table.GeometryColumnName, &table.GeometryType); err != nil {
			return nil, fmt.Errorf("failed to read geometry_columns record, error: %w", err)
		}
		if table.TableName == "" {
			return nil, fmt.Errorf("feature table name is blank, error: %w", err)
		}
		hasCollection := false
		for _, collection := range collections {
			if table.TableName == collection.ID {
				result[collection.ID] = &table
				hasCollection = true
			} else if hasMatchingTableName(collection, table) {
				result[collection.ID] = &table
				hasCollection = true
			}
		}
		if !hasCollection {
			log.Printf("Warning: table %s is present in PostgreSQL but not configured as a collection", table.TableName)
		}
	}
	if len(result) == 0 {
		return nil, errors.New("no records found in PostgreSQL geometry_columns view")
	}

	for _, table := range result {
		table.Schema, err = readSchema(db, *table, fidColumn, externalFidColumn, schemaName, collections)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema for table %s, error: %w", table.TableName, err)
		}
	}

	validateUniqueness(result)
	return result, nil
}

func readPropertyFiltersWithAllowedValues(featTableByCollection map[string]*featureTable,
	collections config.GeoSpatialCollections, db *pgxpool.Pool) (map[string]ds.PropertyFiltersWithAllowedValues, error) {

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
				rows, err := db.Query(context.Background(), query, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to derive allowed values using query: %v\n, error: %w", query, err)
				}
				if err = rows.Scan(&values); err != nil {
					return nil, fmt.Errorf("failed to read result: %w", err)
				}
				// make sure values are valid
				for _, v := range values {
					if newlineRegex.MatchString(v) {
						return nil, fmt.Errorf("failed to derive allowed values, one value contains a "+
							"newline which isn't a valid (OpenAPI) enum value. The value is: %s", v)
					}
				}
				result[collection.ID][pf.Name] = ds.PropertyFilterWithAllowedValues{PropertyFilter: pf, AllowedValues: values}
				continue
			}
		}
	}
	return result, nil
}

func readSchema(db *pgxpool.Pool, table featureTable, fidColumn, externalFidColumn, schemaName string,
	collections config.GeoSpatialCollections) (*d.Schema, error) {

	collectionNames := make([]string, 0, len(collections))
	for _, collection := range collections {
		collectionNames = append(collectionNames, collection.ID)
	}

	query := `
select
    a.attname as column_name,
    pg_catalog.format_type(a.atttypid, a.atttypmod) as data_type,
    a.attnotnull as is_required,
    coalesce(d.description, '') as column_description
from
    pg_catalog.pg_attribute a
join
    pg_catalog.pg_class c on a.attrelid = c.oid
join
    pg_catalog.pg_namespace n on c.relnamespace = n.oid
left join
    pg_catalog.pg_description d on d.objoid = a.attrelid and d.objsubid = a.attnum
where
    n.nspname = $1 
    and c.relname = $2
    and a.attnum > 0 -- Excludes system columns
    and not a.attisdropped -- Excludes columns that have been dropped
order by
    a.attnum;
`

	rows, err := db.Query(context.Background(), query, schemaName, table.TableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fields := make([]d.Field, 0)
	for rows.Next() {
		var colName, colType, colDescription string
		var colNotNull bool
		if err = rows.Scan(&colName, &colType, &colNotNull, &colDescription); err != nil {
			return nil, err
		}
		fields = append(fields, d.Field{
			Name:              colName,
			Type:              colType,
			Description:       colDescription,
			IsRequired:        colNotNull,
			IsPrimaryGeometry: colName == table.GeometryColumnName,
			FeatureRelation:   d.NewFeatureRelation(colName, externalFidColumn, collectionNames),
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

func validateUniqueness(result map[string]*featureTable) {
	uniqueTables := make(map[string]struct{})
	for _, table := range result {
		uniqueTables[table.TableName] = struct{}{}
	}
	if len(uniqueTables) != len(result) {
		log.Printf("Warning: found %d unique table names for %d collections, "+
			"usually each collection is backed by its own unique table\n", len(uniqueTables), len(result))
	}
}
