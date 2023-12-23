package features

import (
	"strings"

	"github.com/PDOK/gokoala/engine"
	ds "github.com/PDOK/gokoala/ogc/features/datasources"
)

type OpenAPIPropertyFilter struct {
	Name        string
	Description string
	DataType    string
}

// rebuildOpenAPIForFeatures Rebuild OpenAPI spec with additional info from given datasources
func rebuildOpenAPIForFeatures(e *engine.Engine, datasources map[DatasourceKey]ds.Datasource) {
	e.RebuildOpenAPI(struct {
		PropertyFiltersByCollection map[string][]OpenAPIPropertyFilter
	}{
		PropertyFiltersByCollection: createPropertyFiltersByCollection(e.Config.OgcAPI.Features, datasources),
	})
}

func createPropertyFiltersByCollection(config *engine.OgcAPIFeatures,
	datasources map[DatasourceKey]ds.Datasource) map[string][]OpenAPIPropertyFilter {

	result := make(map[string][]OpenAPIPropertyFilter)
	for k, datasource := range datasources {
		filtersConfig := config.PropertyFiltersForCollection(k.collectionID)
		if len(filtersConfig) == 0 {
			continue
		}
		featTable, err := datasource.GetFeatureTableMetadata(k.collectionID)
		if err != nil {
			continue
		}
		featTableColumns := featTable.ColumnsWithDataType()
		propertyFilters := make([]OpenAPIPropertyFilter, 0, len(featTableColumns))
		for name, dataType := range featTableColumns {
			for _, fc := range filtersConfig {
				if fc.Name == name {
					// match found between property filter in config file and database column name
					dataType = datasourceToOpenAPI(dataType)
					propertyFilters = append(propertyFilters, OpenAPIPropertyFilter{
						Name:        name,
						Description: fc.Description,
						DataType:    dataType,
					})
				}
			}
		}
		result[k.collectionID] = propertyFilters
	}
	return result
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
