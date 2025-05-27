package features

import (
	"fmt"
	"log"
	"slices"
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
		featTable, err := datasource.GetSchema(k.collectionID)
		if err != nil {
			continue
		}
		propertyFilters := make([]OpenAPIPropertyFilter, 0, len(featTable.Fields))
		for _, fc := range configuredPropertyFilters {
			match := false
			for _, field := range featTable.Fields {
				if fc.Name == field.Name {
					// match found between property filter in config file and database column name
					propertyFilters = append(propertyFilters, OpenAPIPropertyFilter{
						Name:          field.Name,
						Description:   fc.Description,
						DataType:      field.ToTypeFormat().Type,
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
