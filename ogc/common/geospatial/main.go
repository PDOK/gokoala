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

func NewCollections(e *engine.Engine, router *chi.Mux) *Collections {
	if e.Config.HasCollections() {
		e.RenderTemplates(CollectionsPath,
			engine.NewTemplateKey(templatesDir+"collections.go.json"),
			engine.NewTemplateKey(templatesDir+"collections.go.html"))

		for _, coll := range e.Config.OgcAPI.GeoVolumes.Collections {
			e.RenderTemplatesWithParams(coll,
				engine.NewTemplateKeyWithName(templatesDir+"collection.go.json", coll.ID))
			e.RenderTemplatesWithParams(coll,
				engine.NewTemplateKeyWithName(templatesDir+"collection.go.html", coll.ID))
		}
	}

	instance := &Collections{
		engine: e,
	}

	router.Get(CollectionsPath, instance.Collections())
	router.Get(CollectionsPath+"/{collectionId}", instance.Collection())

	return instance
}

// Collections returns list of collections
func (t *Collections) Collections() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKey(templatesDir + "collections.go." + t.engine.CN.NegotiateFormat(r))
		t.engine.ServePage(w, r, key)
	}
}

func (t *Collections) Collection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")

		key := engine.NewTemplateKeyWithName(templatesDir+"collection.go."+t.engine.CN.NegotiateFormat(r), collectionID)
		t.engine.ServePage(w, r, key)
	}
}

// CollectionSupport a collection, also known as a geospatial data resource, is a common way to organize
// data in various OGC APIs.
type CollectionSupport interface {

	// CollectionContent While the generic /collections/{collectionId} endpoint provides METADATA about
	// a collection this endpoint should provide the CONTENTS of a collection.
	//
	// For example in:
	// - OGC API Features you would have: /collections/{collectionId}/items
	// - OGC API Tiles could have: /collections/{collectionId}/tiles
	// - OGC API Maps could have: /collections/{collectionId}/maps
	// - OGC API 3d GeoVolumes doesn't have sub resources, just /collections/{collectionId}/tileset.json
	// etc.
	CollectionContent() http.HandlerFunc
}
