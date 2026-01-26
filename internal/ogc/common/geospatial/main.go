package geospatial

import (
	"net/http"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/go-chi/chi/v5"
)

const (
	CollectionsPath = "/collections"
	templatesDir    = "internal/ogc/common/geospatial/templates/"
)

type Collections struct {
	engine *engine.Engine
}

// Wrapper around collection+type to make it easier to access in the "collection" template.
type collectionWithType struct {
	Collection any
	Type       CollectionType
}

// NewCollections enables support for OGC APIs that organize data in the concept of collections.
// A collection, also known as a geospatial data resource, is a common way to organize data in various OGC APIs.
func NewCollections(e *engine.Engine, types CollectionTypes) *Collections {
	if e.Config.HasCollections() {
		collectionsBreadcrumbs := []engine.Breadcrumb{
			{
				Name: "Collections",
				Path: "collections",
			},
		}
		e.RenderTemplatesWithParams(CollectionsPath,
			types,
			collectionsBreadcrumbs,
			engine.NewTemplateKey(templatesDir+"collections.go.json"),
			engine.NewTemplateKey(templatesDir+"collections.go.html"))

		for _, coll := range e.Config.AllCollections().Unique() {
			title := coll.GetID()
			if coll.GetMetadata() != nil && coll.GetMetadata().Title != nil {
				title = *coll.GetMetadata().Title
			}
			collectionBreadcrumbs := collectionsBreadcrumbs
			collectionBreadcrumbs = append(collectionBreadcrumbs, []engine.Breadcrumb{
				{
					Name: title,
					Path: "collections/" + coll.GetID(),
				},
			}...)

			collWithType := collectionWithType{coll, types.Get(coll.GetID())}
			e.RenderTemplatesWithParams(CollectionsPath+"/"+coll.GetID(), collWithType, nil,
				engine.NewTemplateKey(templatesDir+"collection.go.json", engine.WithInstanceName(coll.GetID())))
			e.RenderTemplatesWithParams(CollectionsPath+"/"+coll.GetID(), collWithType, collectionBreadcrumbs,
				engine.NewTemplateKey(templatesDir+"collection.go.html", engine.WithInstanceName(coll.GetID())))
		}
	}

	instance := &Collections{
		engine: e,
	}

	e.Router.Get(CollectionsPath, instance.Collections())
	e.Router.Get(CollectionsPath+"/{collectionId}", instance.Collection())

	return instance
}

// Collections returns list of collections.
func (c *Collections) Collections() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKey(templatesDir+"collections.go."+c.engine.CN.NegotiateFormat(r), c.engine.WithNegotiatedLanguage(w, r))
		c.engine.Serve(w, r, engine.ServeTemplate(key))
	}
}

// Collection provides METADATA about a specific collection. To get the CONTENTS of a collection each OGC API
// building block must provide a separate/specific endpoint.
//
// For example, in:
// - OGC API Features you would have: /collections/{collectionId}/items
// - OGC API Tiles could have: /collections/{collectionId}/tiles
// - OGC API Maps could have: /collections/{collectionId}/map
// - OGC API 3d GeoVolumes would have: /collections/{collectionId}/3dtiles.
func (c *Collections) Collection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")

		key := engine.NewTemplateKey(templatesDir+"collection.go."+c.engine.CN.NegotiateFormat(r),
			engine.WithInstanceName(collectionID),
			c.engine.WithNegotiatedLanguage(w, r),
		)
		c.engine.Serve(w, r, engine.ServeTemplate(key))
	}
}
