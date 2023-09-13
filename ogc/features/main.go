package features

import (
	"net/http"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/common/geospatial"
	"github.com/PDOK/gokoala/ogc/features/datasources"

	"github.com/go-chi/chi/v5"
)

type Features struct {
	engine     *engine.Engine
	datasource datasources.Datasource
}

func NewFeatures(e *engine.Engine, router *chi.Mux) *Features {
	var datasource datasources.Datasource
	if e.Config.OgcAPI.Features.Datasource.FakeDB {
		datasource = datasources.NewFakeDB()
	} else if e.Config.OgcAPI.Features.Datasource.GeoPackage != nil {
		datasource = datasources.NewGeoPackage()
	}
	// TODO: call datasource.Close() once server exists.

	features := &Features{
		engine:     e,
		datasource: datasource,
	}

	router.Get(geospatial.CollectionsPath+"/{collectionId}/items", features.CollectionContent())
	router.Get(geospatial.CollectionsPath+"/{collectionId}/items/{featureId}", features.Feature())
	return features
}

// CollectionContent serve FeatureCollection with the given collectionId
func (f *Features) CollectionContent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")

		fc := f.datasource.GetFeatures(collectionID)
		fcJSON, err := fc.MarshalJSON()
		if err != nil {
			http.Error(w, "Failed to marshall FeatureCollection to JSON", http.StatusInternalServerError)
			return
		}
		engine.SafeWrite(w.Write, fcJSON)
	}
}

// Feature serves a specific Feature
func (f *Features) Feature() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")
		featureID := chi.URLParam(r, "featureId")

		feat := f.datasource.GetFeature(collectionID, featureID)
		featJSON, err := feat.MarshalJSON()
		if err != nil {
			http.Error(w, "Failed to marshall Feature to JSON", http.StatusInternalServerError)
			return
		}
		engine.SafeWrite(w.Write, featJSON)
	}
}
