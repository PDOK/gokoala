package geospatial

import (
	"net/http"

	"github.com/PDOK/gomagpie/internal/engine"
)

const (
	CollectionsPath = "/collections"
	templatesDir    = "internal/ogc/common/geospatial/templates/"
)

type Collections struct {
	engine *engine.Engine
}

// NewCollections enables support for OGC APIs that organize data in the concept of collections.
// A collection, also known as a geospatial data resource, is a common way to organize data in various OGC APIs.
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
	}

	instance := &Collections{
		engine: e,
	}

	e.Router.Get(CollectionsPath, instance.Collections())

	return instance
}

// Collections returns list of collections
func (c *Collections) Collections() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKeyWithLanguage(templatesDir+"collections.go."+c.engine.CN.NegotiateFormat(r), c.engine.CN.NegotiateLanguage(w, r))
		c.engine.ServePage(w, r, key)
	}
}
