package features

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/PDOK/gokoala/internal/engine"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

type openAPIParams struct {
	PropertyFiltersByCollection map[string][]OpenAPIPropertyFilter
	SchemasByCollection         map[string]domain.Schema
}

type OpenAPIPropertyFilter struct {
	Name          string
	Description   string
	DataType      string
	AllowedValues []string
}

// rebuildOpenAPI Rebuild OpenAPI spec for features with additional info from given parameters
func rebuildOpenAPI(e *engine.Engine,
	datasources map[datasourceKey]ds.Datasource,
	filters map[string]ds.PropertyFiltersWithAllowedValues,
	schemas map[string]domain.Schema) {

	propertyFiltersByCollection, err := createPropertyFiltersByCollection(datasources, filters)
	if err != nil {
		log.Fatal(err)
	}
	e.RebuildOpenAPI(openAPIParams{
		PropertyFiltersByCollection: propertyFiltersByCollection,
		SchemasByCollection:         schemas,
	})
}

func createPropertyFiltersByCollection(datasources map[datasourceKey]ds.Datasource,
	filters map[string]ds.PropertyFiltersWithAllowedValues) (map[string][]OpenAPIPropertyFilter, error) {

	result := make(map[string][]OpenAPIPropertyFilter)
	for k, datasource := range datasources {
		configuredPropertyFilters := filters[k.collectionID]
		if len(configuredPropertyFilters) == 0 {
			continue
		}
		featTable, err := datasource.GetSchema(k.collectionID)
		if err != nil {
			continue
		}
		featTableColumns := featTable.FieldsWithDataType()
		propertyFilters := make([]OpenAPIPropertyFilter, 0, len(featTableColumns))
		for _, fc := range configuredPropertyFilters {
			match := false
			for name, dataType := range featTableColumns {
				if fc.Name == name {
					// match found between property filter in config file and database column name
					dataType = datasourceToOpenAPI(dataType)
					propertyFilters = append(propertyFilters, OpenAPIPropertyFilter{
						Name:          name,
						Description:   fc.Description,
						DataType:      dataType,
						AllowedValues: fc.AllowedValues,
					})
					match = true
					break
				}
			}
			if !match {
				return nil, fmt.Errorf("invalid property filter specified, "+
					"column '%s' doesn't exist in datasource attached to collection '%s'", fc.Name, k.collectionID)
			}
		}
		slices.SortFunc(propertyFilters, func(a, b OpenAPIPropertyFilter) int {
			return strings.Compare(a.Name, b.Name)
		})
		result[k.collectionID] = propertyFilters
	}
	return result, nil
}

// translate database data types to OpenAPI data types
func datasourceToOpenAPI(dataType string) string {
	switch strings.ToUpper(dataType) {
	case "INTEGER":
		dataType = "integer"
	case "REAL", "NUMERIC":
		dataType = "number"
	case "TEXT", "VARCHAR":
		dataType = "string"
	default:
		dataType = "string"
	}
	return dataType
}
