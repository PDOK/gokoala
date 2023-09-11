package features

import (
	"log"
	"net/http"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/common/geospatial"
	"github.com/PDOK/gokoala/ogc/features/datasources"

	"github.com/go-chi/chi/v5"
)

// Features !!! Placeholder implementation, for future reference !!!
type Features struct {
	engine     *engine.Engine
	datasource datasources.Datasource
}

// NewFeatures !!! Placeholder implementation, for future reference !!!
func NewFeatures(e *engine.Engine, router *chi.Mux) *Features {
	var datasource datasources.Datasource
	if e.Config.OgcAPI.Features.Datasource.FakeDB {
		datasource = datasources.NewFakeDB()
	} else if e.Config.OgcAPI.Features.Datasource.GeoPackage != nil {
		datasource = datasources.NewGeoPackage()
	}

	features := &Features{
		engine:     e,
		datasource: datasource,
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
