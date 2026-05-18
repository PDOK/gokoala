package features

import (
	"slices"
	"strings"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

type openAPIParams struct {
	PropertyFiltersByCollection map[string][]OpenAPIPropertyFilter
	CollectionTypes             geospatial.CollectionTypes
	SchemasByCollection         map[string]domain.Schema
}

type OpenAPIPropertyFilter struct {
	Name          string
	Description   string
	DataType      string
	AllowedValues []string
}

// rebuildOpenAPI Rebuild OpenAPI spec for features with additional info from given parameters.
func rebuildOpenAPI(e *engine.Engine,
	queryablesByCollection map[string]ds.Queryables,
	collectionTypes geospatial.CollectionTypes,
	schemas map[string]domain.Schema) {

	propertyFiltersByCollection := toOpenAPIFilters(queryablesByCollection)
	e.RebuildOpenAPI(openAPIParams{
		PropertyFiltersByCollection: propertyFiltersByCollection,
		CollectionTypes:             collectionTypes,
		SchemasByCollection:         schemas,
	})
}

func toOpenAPIFilters(queryablesByCollection map[string]ds.Queryables) map[string][]OpenAPIPropertyFilter {
	result := make(map[string][]OpenAPIPropertyFilter)
	for collectionID, queryables := range queryablesByCollection {
		if len(queryables) == 0 {
			continue
		}
		filters := make([]OpenAPIPropertyFilter, 0, len(queryables))
		for _, queryable := range queryables {
			if queryable.IsPrimaryGeometry {
				// no need to expose geometry as a property filter (but can be used in CQL)
				continue
			}
			filters = append(filters, OpenAPIPropertyFilter{
				Name:          queryable.Name,
				Description:   queryable.Description,
				DataType:      queryable.ToTypeFormat().Type,
				AllowedValues: queryable.AllowedValues,
			})
		}
		slices.SortFunc(filters, func(a, b OpenAPIPropertyFilter) int {
			return strings.Compare(a.Name, b.Name)
		})
		result[collectionID] = filters
	}

	return result
}
