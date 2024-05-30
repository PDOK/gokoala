package features

import (
	"fmt"
	"log"
	"strings"

	"github.com/PDOK/gokoala/internal/engine"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
)

type OpenAPIPropertyFilter struct {
	Name          string
	Description   string
	DataType      string
	AllowedValues []string
}

// rebuildOpenAPIForFeatures Rebuild OpenAPI spec with additional info from given datasources
func rebuildOpenAPIForFeatures(e *engine.Engine, datasources map[DatasourceKey]ds.Datasource, filters map[string]ds.PropertyFiltersWithAllowedValues) {
	propertyFiltersByCollection, err := createPropertyFiltersByCollection(datasources, filters)
	if err != nil {
		log.Fatal(err)
	}
	e.RebuildOpenAPI(struct {
		PropertyFiltersByCollection map[string][]OpenAPIPropertyFilter
	}{
		PropertyFiltersByCollection: propertyFiltersByCollection,
	})
}

func createPropertyFiltersByCollection(datasources map[DatasourceKey]ds.Datasource,
	filters map[string]ds.PropertyFiltersWithAllowedValues) (map[string][]OpenAPIPropertyFilter, error) {

	result := make(map[string][]OpenAPIPropertyFilter)
	for k, datasource := range datasources {
		configuredPropertyFilters := filters[k.collectionID]
		if len(configuredPropertyFilters) == 0 {
			continue
		}
		featTable, err := datasource.GetFeatureTableMetadata(k.collectionID)
		if err != nil {
			continue
		}
		featTableColumns := featTable.ColumnsWithDataType()
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
