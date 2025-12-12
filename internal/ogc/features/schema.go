package features

import (
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	g "github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/go-chi/chi/v5"
)

const schemasPath = "/schema"
const schemaHTML = templatesDir + "schema.go.html"
const schemaJSON = templatesDir + "schema.go.json"

// Schema endpoint serves a schema that describes the features in the collection, either as HTML
// or as JSON schema (https://json-schema.org/)
func (f *Features) Schema() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f.engine.OpenAPI.ValidateRequest(r); err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())

			return
		}

		collectionID := chi.URLParam(r, "collectionId")
		collection, ok := f.configuredCollections[collectionID]
		if !ok {
			handleCollectionNotFound(w, collectionID)

			return
		}

		var key engine.TemplateKey
		format := f.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatHTML:
			key = engine.NewTemplateKey(schemaHTML,
				engine.WithInstanceName(collection.ID),
				f.engine.WithNegotiatedLanguage(w, r))
		case engine.FormatJSON:
			key = engine.NewTemplateKey(schemaJSON,
				engine.WithInstanceName(collection.ID),
				f.engine.WithNegotiatedLanguage(w, r),
				engine.WithMediaTypeOverwrite(engine.MediaTypeJSONSchema)) // JSON format, but specific mediatype.
		default:
			handleFormatNotSupported(w, format)

			return
		}
		f.engine.Serve(w, r, engine.ServeTemplate(key))
	}
}

type schemaTemplateData struct {
	domain.Schema

	CollectionID          string
	CollectionTitle       string
	CollectionDescription *string
}

// renderSchemas pre-renders HTML and JSON schemas describing each feature collection.
func renderSchemas(e *engine.Engine, datasources map[datasourceKey]ds.Datasource) map[string]domain.Schema {
	schemasByCollection := make(map[string]domain.Schema)
	for _, collection := range e.Config.OgcAPI.Features.Collections {
		title, description := getCollectionTitleAndDesc(collection)

		breadcrumbs := collectionsBreadcrumb
		breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
			{
				Name: title,
				Path: collectionsCrumb + collection.ID,
			},
			{
				Name: "Schema",
				Path: collectionsCrumb + collection.ID + schemasPath,
			},
		}...)

		// the schema should be the same regardless of CRS, so we use WGS84 as it's the default and always present
		datasource := datasources[datasourceKey{srid: domain.WGS84SRID, collectionID: collection.ID}]
		schema, err := datasource.GetSchema(collection.ID)
		if err != nil {
			log.Printf("Failed to render OGC API Features part 5 Schema for collection %s: %v", collection.ID, err)

			continue
		}

		// expand the schema with details about temporal fields
		if collection.Metadata != nil && collection.Metadata.TemporalProperties != nil {
			for i := range schema.Fields {
				// OAF part 5: If the features have multiple temporal properties, the roles "primary-interval-start"
				// and "primary-interval-end" can be used to identify the primary temporal information of the features.
				if collection.Metadata.TemporalProperties.StartDate == schema.Fields[i].Name {
					schema.Fields[i].IsPrimaryIntervalStart = true
				} else if collection.Metadata.TemporalProperties.EndDate == schema.Fields[i].Name {
					schema.Fields[i].IsPrimaryIntervalEnd = true
				}
			}
		}

		if !requiresSpecificOrder(collection) {
			// stable field order
			slices.SortFunc(schema.Fields, func(a, b domain.Field) int {
				return strings.Compare(a.Name, b.Name)
			})
		}

		// pre-render the schema, catches issues early on during start-up.
		e.RenderTemplatesWithParams(g.CollectionsPath+"/"+collection.ID+schemasPath,
			schemaTemplateData{
				*schema,
				collection.ID,
				title,
				description,
			},
			breadcrumbs,
			engine.NewTemplateKey(schemaJSON,
				engine.WithInstanceName(collection.ID),
				engine.WithMediaTypeOverwrite(engine.MediaTypeJSONSchema),
			),
			engine.NewTemplateKey(schemaHTML,
				engine.WithInstanceName(collection.ID),
			),
		)

		schemasByCollection[collection.ID] = *schema
	}

	return schemasByCollection
}

func requiresSpecificOrder(collection config.GeoSpatialCollection) bool {
	if collection.Features != nil && collection.Features.FeatureProperties != nil {
		return collection.Features.PropertiesInSpecificOrder
	}

	return false
}

func getCollectionTitleAndDesc(collection config.GeoSpatialCollection) (string, *string) {
	var description *string
	if collection.Metadata != nil {
		description = collection.Metadata.Description
	}

	return getCollectionTitle(collection.ID, collection.Metadata), description
}
