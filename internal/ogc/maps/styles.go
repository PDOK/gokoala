package maps

import (
	"net/http"
	"slices"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	"github.com/go-chi/chi/v5"
)

const (
	stylesCrumb       = "/styles"
	stylesTemplateKey = "styles.go."
)

type stylesPage struct {
	Collection config.MapsCollection
	StyleIDs []string
}

func renderStyles(e *engine.Engine, collections map[string]config.MapsCollection) {
	for _, collection := range collections {
		path := geospatial.CollectionsPath + "/" + collection.ID + stylesCrumb
		page := &stylesPage{
			Collection: collection,
			StyleIDs: getStyleIDsFromCollection(collection),
		}
		instanceName := engine.WithInstanceName(collection.ID)

		e.RenderTemplatesWithParams(path, page, stylesBreadcrumbs(collection),
			engine.NewTemplateKey(templatesDir+stylesTemplateKey+engine.FormatJSON, instanceName),
			engine.NewTemplateKey(templatesDir+stylesTemplateKey+engine.FormatHTML, instanceName))
	}
}

func (m *Maps) Styles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")
		if _, ok := m.configuredCollections[collectionID]; !ok {
			handleCollectionNotFound(w, collectionID)
			return
		}

		format := m.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatJSON, engine.FormatHTML:
			key := engine.NewTemplateKey(templatesDir+stylesTemplateKey+format,
				engine.WithInstanceName(collectionID), m.engine.WithNegotiatedLanguage(w, r))
			m.engine.Serve(w, r, engine.ServeTemplate(key))
		default:
			handleFormatNotSupported(w, format)
		}
	}
}

func getStyleIDsFromCollection(collection config.MapsCollection) []string {
	var ids []string
	for _, layer := range collection.Layers {
		for _, style := range layer.Styles {
			if !slices.Contains(ids, style.ID) {
				ids = append(ids, style.ID)
			}
		}
	}

	return ids
}

func stylesBreadcrumbs(collection config.MapsCollection) []engine.Breadcrumb {
	return []engine.Breadcrumb{
		{
			Name: "Collections",
			Path: "collections",
		},
		{
			Name: getCollectionTitle(collection.GetID(), collection.GetMetadata()),
			Path: collectionsCrumb + collection.ID,
		},
		{
			Name: "Styles",
			Path: collectionsCrumb + collection.ID + stylesCrumb,
		},
	}
}
