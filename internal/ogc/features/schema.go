package features

import (
	"log"
	"net/http"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/go-chi/chi/v5"
)

const schemasPath = "/schema"

type schemaTemplateData struct {
	domain.Schema

	CollectionID string
	Metadata     *config.GeoSpatialCollectionMetadata
}

func renderSchemas(e *engine.Engine, datasources map[DatasourceKey]ds.Datasource) {
	for _, collection := range e.Config.OgcAPI.Features.Collections {
		breadcrumbs := collectionsBreadcrumb
		breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
			{
				Name: getCollectionTitle(collection.ID, collection.Metadata),
				Path: collectionsCrumb + collection.ID,
			},
			{
				Name: "Schema",
				Path: collectionsCrumb + collection.ID + schemasPath,
			},
		}...)

		// the schema should be the same regardless of CRS, so we use WGS84 as it's the default and always present
		datasource := datasources[DatasourceKey{srid: domain.WGS84SRID, collectionID: collection.ID}]
		schema, err := datasource.GetSchema(collection.ID)
		if err != nil {
			log.Printf("Failed to render OGC API Features part 5 Schema for collection %s: %v", collection.ID, err)
			continue
		}

		schemaTemplateData := schemaTemplateData{
			schema,
			collection.ID,
			collection.Metadata,
		}

		e.RenderTemplatesWithParams(geospatial.CollectionsPath+"/"+collection.ID+schemasPath,
			schemaTemplateData,
			breadcrumbs,
			engine.NewTemplateKey(templatesDir+"schema.go.json",
				engine.WithInstanceName(collection.ID),
				engine.WithMediaTypeOverwrite(engine.MediaTypeJSONSchema),
			),
			engine.NewTemplateKey(templatesDir+"schema.go.html",
				engine.WithInstanceName(collection.ID),
			))
	}
}

// Schema serves a schema that describes the features in the collection, either as HTML
// or as JSON schema (https://json-schema.org/)
func (f *Features) Schema() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f.engine.OpenAPI.ValidateRequest(r); err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}

		collectionID := chi.URLParam(r, "collectionId")
		collection, ok := configuredCollections[collectionID]
		if !ok {
			handleCollectionNotFound(w, collectionID)
			return
		}

		var key engine.TemplateKey
		switch f.engine.CN.NegotiateFormat(r) {
		case engine.FormatHTML:
			key = engine.NewTemplateKey(templatesDir+"schema.go.html",
				engine.WithInstanceName(collection.ID), f.engine.WithNegotiatedLanguage(w, r))
		case engine.FormatJSON:
			key = engine.NewTemplateKey(templatesDir+"schema.go.json",
				engine.WithInstanceName(collection.ID), f.engine.WithNegotiatedLanguage(w, r),
				engine.WithMediaTypeOverwrite(engine.MediaTypeJSONSchema))
		default:
			engine.RenderProblem(engine.ProblemNotAcceptable, w, "format is not supported")
			return
		}
		f.engine.Serve(w, r, engine.ServeTemplate(key))
	}
}
