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
	filters map[string]ds.QueryablesWithAllowedValues,
	collectionTypes geospatial.CollectionTypes,
	schemas map[string]domain.Schema) {

	propertyFiltersByCollection := createPropertyFiltersByCollection(filters)
	e.RebuildOpenAPI(openAPIParams{
		PropertyFiltersByCollection: propertyFiltersByCollection,
		CollectionTypes:             collectionTypes,
		SchemasByCollection:         schemas,
	})
}

func createPropertyFiltersByCollection(filters map[string]ds.QueryablesWithAllowedValues) map[string][]OpenAPIPropertyFilter {
	result := make(map[string][]OpenAPIPropertyFilter)
	for collectionID, filter := range filters {
		if len(filter) == 0 {
			continue
		}
		filtersForCollection := make([]OpenAPIPropertyFilter, 0, len(filter))
		for _, fc := range filter {
			filtersForCollection = append(filtersForCollection, OpenAPIPropertyFilter{
				Name:          fc.Name,
				Description:   fc.Description,
				DataType:      fc.ToTypeFormat().Type,
				AllowedValues: fc.AllowedValues,
			})
		}
		slices.SortFunc(filtersForCollection, func(a, b OpenAPIPropertyFilter) int {
			return strings.Compare(a.Name, b.Name)
		})
		result[collectionID] = filtersForCollection
	}

	return result
}
