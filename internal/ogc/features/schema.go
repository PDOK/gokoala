package features

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/go-chi/chi/v5"
)

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

		// the schema should be the same regardless of CRS, so we use WGS84 as it's the default and always present
		datasource := f.datasources[DatasourceKey{srid: domain.WGS84SRID, collectionID: collectionID}]
		schema, err := datasource.GetSchema(collection.ID)
		if err != nil {
			engine.RenderProblem(engine.ProblemServerError, w, err.Error())
			return
		}

		log.Printf("%v", schema) // TODO: remove this line when we're done with debugging'

		format := f.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatHTML:
			//f.html.schema(w, r, collectionID)
		case engine.FormatJSON:
			//f.json.schema(w, r, collectionID)
		default:
			engine.RenderProblem(engine.ProblemNotAcceptable, w, fmt.Sprintf("format '%s' is not supported", format))
			return
		}
	}
}
