package maps

import (
	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
)

const (
	templatesDir = "internal/ogc/maps/templates/"
	mapsHTML     = "maps.go.html"
	mapsMD       = "maps.go.md"
)

type Maps struct {
	engine *engine.Engine

	configuredCollections map[string]config.MapsCollection

	html *htmlMaps
	json *jsonMaps
}

func NewMaps(e *engine.Engine) *Maps {
	m := &Maps{
		engine:                e,
		configuredCollections: cacheConfiguredMapsCollections(e),
		html:                  newHTMLMaps(e),
		json:                  newJSONMaps(e),
	}

	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/map", m.Maps())
	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/styles", m.Maps())
	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/styles/{styleId}/map", m.Maps())

	return m
}

func cacheConfiguredMapsCollections(e *engine.Engine) map[string]config.MapsCollection {
	result := make(map[string]config.MapsCollection)
	for _, collection := range e.Config.OgcAPI.Maps.Collections {
		result[collection.ID] = collection
	}

	return result
}
