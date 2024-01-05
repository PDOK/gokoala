package geospatial

import (
	"net/http"

	"github.com/PDOK/gokoala/engine"
	"github.com/go-chi/chi/v5"
)

const (
	CollectionsPath = "/collections"
	templatesDir    = "ogc/common/geospatial/templates/"
)

type Collections struct {
	engine *engine.Engine
}

// NewCollections A collection, also known as a geospatial data resource, is a common way
// to organize data in various OGC APIs.
func NewCollections(e *engine.Engine) *Collections {
	if e.Config.HasCollections() {
		collectionsBreadcrumbs := []engine.Breadcrumb{
			{
				Name: "Collections",
				Path: "collections",
			},
		}
		e.RenderTemplates(CollectionsPath,
			collectionsBreadcrumbs,
			engine.NewTemplateKey(templatesDir+"collections.go.json"),
			engine.NewTemplateKey(templatesDir+"collections.go.html"))

		for _, coll := range e.Config.AllCollections().Unique() {
			title := coll.ID
			if coll.Metadata != nil && coll.Metadata.Title != nil {
				title = *coll.Metadata.Title
			}
			collectionBreadcrumbs := collectionsBreadcrumbs
			collectionBreadcrumbs = append(collectionBreadcrumbs, []engine.Breadcrumb{
				{
					Name: title,
					Path: "collections/" + coll.ID,
				},
			}...)
			e.RenderTemplatesWithParams(coll,
				nil,
				engine.NewTemplateKeyWithName(templatesDir+"collection.go.json", coll.ID))
			e.RenderTemplatesWithParams(coll,
				collectionBreadcrumbs,
				engine.NewTemplateKeyWithName(templatesDir+"collection.go.html", coll.ID))
		}
	}

	instance := &Collections{
		engine: e,
	}

	e.Router.Get(CollectionsPath, instance.Collections())
	e.Router.Get(CollectionsPath+"/{collectionId}", instance.Collection())

	return instance
}

// Collections returns list of collections
func (c *Collections) Collections() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKeyWithLanguage(templatesDir+"collections.go."+c.engine.CN.NegotiateFormat(r), c.engine.CN.NegotiateLanguage(w, r))
		c.engine.ServePage(w, r, key)
	}
}

// Collection provides METADATA about a specific collection. To get the CONTENTS of a collection each OGC API
// building block must provide a separate/specific endpoint.
//
// For example in:
// - OGC API Features you would have: /collections/{collectionId}/items
// - OGC API Tiles could have: /collections/{collectionId}/tiles
// - OGC API Maps could have: /collections/{collectionId}/maps
// - OGC API 3d GeoVolumes would have: /collections/{collectionId}/3dtiles
func (c *Collections) Collection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")

		key := engine.NewTemplateKeyWithNameAndLanguage(templatesDir+"collection.go."+c.engine.CN.NegotiateFormat(r), collectionID, c.engine.CN.NegotiateLanguage(w, r))
		c.engine.ServePage(w, r, key)
	}
}
