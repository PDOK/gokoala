package features

import (
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

func renderSchemas(e *engine.Engine, datasources map[DatasourceKey]ds.Datasource) error {
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
			return err
		}

		schemaTemplateData := schemaTemplateData{
			schema,
			collection.ID,
			collection.Metadata,
		}

		e.RenderTemplatesWithParams(geospatial.CollectionsPath+"/"+collection.ID+schemasPath,
			schemaTemplateData,
			breadcrumbs,
			engine.NewTemplateKeyWithName(templatesDir+"schema.go.json", collection.ID),
			engine.NewTemplateKeyWithName(templatesDir+"schema.go.html", collection.ID))
	}
	return nil
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

		key := engine.NewTemplateKeyWithNameAndLanguage(
			templatesDir+"schema.go."+f.engine.CN.NegotiateFormat(r), collection.ID, f.engine.CN.NegotiateLanguage(w, r))
		f.engine.Serve(w, r, engine.ServeTemplate(key))
	}
}
