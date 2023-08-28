package geovolumes

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/common/geospatial"

	"github.com/go-chi/chi/v5"
)

type ThreeDimensionalGeoVolumes struct {
	engine *engine.Engine
}

func NewThreeDimensionalGeoVolumes(e *engine.Engine, router *chi.Mux) *ThreeDimensionalGeoVolumes {
	_, err := url.ParseRequestURI(e.Config.OgcAPI.GeoVolumes.TileServer.String())
	if err != nil {
		log.Fatalf("invalid tileserver url provided: %v", err)
	}

	geoVolumes := &ThreeDimensionalGeoVolumes{
		engine: e,
	}

	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/3dtiles", geoVolumes.CollectionContent())
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/3dtiles/{explicitTileSet}.json", geoVolumes.ExplicitTileset())
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/3dtiles/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/3dtiles/{tilePathPrefix}/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())

	// '/3dtiles/' path is preferred but optional when requesting the actual tiles/tileset.
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/{explicitTileSet}.json", geoVolumes.ExplicitTileset())
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/{tilePathPrefix}/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())

	return geoVolumes
}

// CollectionContent reverse proxy to tileserver for tileset.json (the latter contains
// data from OGC 3D Tiles, separate spec from OGC 3D GeoVolumes)
func (t *ThreeDimensionalGeoVolumes) CollectionContent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.tileSet(w, r, "tileset.json")
	}
}

// ExplicitTileset reverse proxy to tileserver for a specific JSON tileset (the latter contains
// data from OGC 3D Tiles, separate spec from OGC 3D GeoVolumes)
func (t *ThreeDimensionalGeoVolumes) ExplicitTileset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tileSetName := chi.URLParam(r, "explicitTileSet")
		if tileSetName == "" {
			http.NotFound(w, r)
			return
		}
		t.tileSet(w, r, tileSetName+".json")
	}
}

// Tile reverse proxy to tileserver for actual 3D tiles (from OGC 3D Tiles, separate spec from OGC 3D GeoVolumes)
func (t *ThreeDimensionalGeoVolumes) Tile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		containerID, err := t.containerIDToPathPrefix(chi.URLParam(r, "3dContainerId"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		tilePathPrefix := chi.URLParam(r, "tilePathPrefix") // optional
		tileMatrix := chi.URLParam(r, "tileMatrix")
		tileRow := chi.URLParam(r, "tileRow")
		tileColAndSuffix := chi.URLParam(r, "tileColAndSuffix")

		path, _ := url.JoinPath("/", containerID, tilePathPrefix, tileMatrix, tileRow, tileColAndSuffix)

		t.reverseProxy(w, r, path, true)
	}
}

func (t *ThreeDimensionalGeoVolumes) tileSet(w http.ResponseWriter, r *http.Request, tileSet string) {
	containerID, err := t.containerIDToPathPrefix(chi.URLParam(r, "3dContainerId"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	path, _ := url.JoinPath("/", containerID, tileSet)
	t.reverseProxy(w, r, path, false)
}

func (t *ThreeDimensionalGeoVolumes) reverseProxy(w http.ResponseWriter, r *http.Request, path string, prefer204 bool) {
	target, err := url.Parse(t.engine.Config.OgcAPI.GeoVolumes.TileServer.String() + path)
	if err != nil {
		log.Printf("invalid target url, can't proxy tiles: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	t.engine.ReverseProxy(w, r, target, prefer204, "")
}

func (t *ThreeDimensionalGeoVolumes) containerIDToPathPrefix(cid string) (string, error) {
	for _, collection := range t.engine.Config.OgcAPI.GeoVolumes.Collections {
		if collection.ID == cid {
			if collection.GeoVolumes != nil && collection.GeoVolumes.TileServerPath != nil {
				return *collection.GeoVolumes.TileServerPath, nil
			}
			return cid, nil
		}
	}
	return "", errors.New("no matching collection found")
}
