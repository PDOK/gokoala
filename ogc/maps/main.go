package maps

import (
	"log"
	"net/http"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/common/geospatial"

	"github.com/go-chi/chi/v5"
)

// Maps !!! Placeholder implementation, for future reference !!!
type Maps struct {
	engine *engine.Engine
}

// NewMaps !!! Placeholder implementation, for future reference !!!
func NewMaps(e *engine.Engine) *Maps {
	maps := &Maps{
		engine: e,
	}

	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/map", maps.CollectionContent())
	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/map/tiles", maps.CollectionContent())
	return maps
}

func (t *Maps) CollectionContent(_ ...any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")

		// TODO: not implemented yet
		log.Printf("TODO: return maps for collection %s", collectionID)
	}
}
