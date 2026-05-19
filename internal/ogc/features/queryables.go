package features

import (
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/PDOK/gokoala/internal/engine"
	g "github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/go-chi/chi/v5"
)

const queryablesPath = "/queryables"

const (
	queryablesHTML = templatesDir + "queryables.go.html"
	queryablesJSON = templatesDir + "queryables.go.json"
)

// Queryables endpoint describes the properties of each feature that can be used for filtering, either as HTML
// or as JSON schema (https://json-schema.org/)
//
//nolint:dupl
func (f *Features) Queryables() http.HandlerFunc {
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
			key = engine.NewTemplateKey(queryablesHTML,
				engine.WithInstanceName(collection.GetID()),
				engine.WithInclude(fieldsIncludeHTML),
				f.engine.WithNegotiatedLanguage(w, r))
		case engine.FormatJSON:
			key = engine.NewTemplateKey(queryablesJSON,
				engine.WithInstanceName(collection.GetID()),
				engine.WithInclude(fieldsIncludeJSON),
				f.engine.WithNegotiatedLanguage(w, r),
				engine.WithMediaTypeOverwrite(engine.MediaTypeJSONSchema)) // JSON format, but specific mediatype.
		default:
			handleFormatNotSupported(w, format)

			return
		}
		f.engine.Serve(w, r, engine.ServeTemplate(key))
	}
}

type queryablesTemplateData struct {
	Fields []domain.Field

	CollectionID          string
	CollectionTitle       string
	CollectionDescription *string
}

// renderQueryables pre-renders HTML and JSON queryables describing each feature collection.
func renderQueryables(e *engine.Engine, queryablesByCollection map[string]ds.Queryables) {
	for _, collection := range e.Config.OgcAPI.Features.Collections {
		title, description := getCollectionTitleAndDesc(collection)

		breadcrumbs := collectionsBreadcrumb
		breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
			{
				Name: title,
				Path: collectionsCrumb + collection.ID,
			},
			{
				Name: "Queryables",
				Path: collectionsCrumb + collection.ID + queryablesPath,
			},
		}...)

		queryables, ok := queryablesByCollection[collection.ID]
		if !ok {
			log.Printf("Queryables for collection %s not found, skipping rendering", collection.ID)
			continue
		}
		queryableFields := queryables.Fields()

		if !requiresSpecificOrder(collection) {
			// stable field order
			slices.SortFunc(queryableFields, func(a, b domain.Field) int {
				return strings.Compare(a.Name, b.Name)
			})
		}

		// pre-render the queryables, catches issues early on during start-up.
		e.RenderTemplatesWithParams(g.CollectionsPath+"/"+collection.ID+queryablesPath,
			queryablesTemplateData{
				queryableFields,
				collection.ID,
				title,
				description,
			},
			breadcrumbs,
			engine.NewTemplateKey(queryablesJSON,
				engine.WithInstanceName(collection.ID),
				engine.WithInclude(fieldsIncludeJSON),
				engine.WithMediaTypeOverwrite(engine.MediaTypeJSONSchema),
			),
			engine.NewTemplateKey(queryablesHTML,
				engine.WithInstanceName(collection.ID),
				engine.WithInclude(fieldsIncludeHTML),
			),
		)
	}
}
