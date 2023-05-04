package features

import (
	"gokoala/engine"
	"gokoala/ogc/common/geospatial"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Features !!! Placeholder implementation, for future reference !!!
type Features struct {
	engine *engine.Engine
}

// NewFeatures !!! Placeholder implementation, for future reference !!!
func NewFeatures(e *engine.Engine, router *chi.Mux) *Features {
	features := &Features{
		engine: e,
	}

	router.Get(geospatial.CollectionsPath+"/{collectionId}/items", features.CollectionContent())
	return features
}

func (t *Features) CollectionContent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")

		// TODO: not implemented yet
		log.Printf("TODO: return features for collection %s", collectionID)
	}
}
