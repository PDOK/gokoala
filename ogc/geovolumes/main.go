package geovolumes

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

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

	// 3D Tiles
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/3dtiles", geoVolumes.CollectionContent("tileset.json"))
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/3dtiles/{explicitTileSet}.json", geoVolumes.ExplicitTileset())
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/3dtiles/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/3dtiles/{tilePathPrefix}/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())

	// DTM/Quantized Mesh
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/quantized-mesh", geoVolumes.CollectionContent("layer.json"))
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/quantized-mesh/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/quantized-mesh/{tilePathPrefix}/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())

	// path '/3dtiles/' or '/quantized-mesh' is preferred but optional when requesting the actual tiles/tileset.
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/{explicitTileSet}.json", geoVolumes.ExplicitTileset())
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())
	router.Get(geospatial.CollectionsPath+"/{3dContainerId}/{tilePathPrefix}/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())

	return geoVolumes
}

// CollectionContent reverse proxy to tileserver for tileset.json  OGC 3D Tiles manifest, separate
// spec from OGC 3D GeoVolumes) or the equivalent manifest (layer.json) for a quantized mesh
func (t *ThreeDimensionalGeoVolumes) CollectionContent(fileName string) http.HandlerFunc {
	if !strings.HasSuffix(fileName, ".json") {
		log.Fatalf("manifest should be a JSON file")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		t.tileSet(w, r, fileName)
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

// Tile reverse proxy to tileserver for actual 3D tiles (from OGC 3D Tiles, separate spec
// from OGC 3D GeoVolumes) or DTM Quantized Mesh tiles
func (t *ThreeDimensionalGeoVolumes) Tile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "3dContainerId")
		collection, err := t.idToCollection(collectionID)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		tileServerPath := collectionID
		if collection.GeoVolumes != nil && collection.GeoVolumes.TileServerPath != nil {
			tileServerPath = *collection.GeoVolumes.TileServerPath
		}
		tilePathPrefix := chi.URLParam(r, "tilePathPrefix") // optional
		tileMatrix := chi.URLParam(r, "tileMatrix")
		tileRow := chi.URLParam(r, "tileRow")
		tileColAndSuffix := chi.URLParam(r, "tileColAndSuffix")

		contentType := ""
		if collection.GeoVolumes != nil && collection.GeoVolumes.URITemplateDTM != nil {
			// DTM has a specialized mediatype, although application/octet-stream will also work
			contentType = engine.MediaTypeQuantizedMesh
		}

		path, _ := url.JoinPath("/", tileServerPath, tilePathPrefix, tileMatrix, tileRow, tileColAndSuffix)
		t.reverseProxy(w, r, path, true, contentType)
	}
}

func (t *ThreeDimensionalGeoVolumes) tileSet(w http.ResponseWriter, r *http.Request, tileSet string) {
	collectionID := chi.URLParam(r, "3dContainerId")
	collection, err := t.idToCollection(collectionID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tileServerPath := collectionID
	if collection.GeoVolumes != nil && collection.GeoVolumes.TileServerPath != nil {
		tileServerPath = *collection.GeoVolumes.TileServerPath
	}

	path, _ := url.JoinPath("/", tileServerPath, tileSet)
	t.reverseProxy(w, r, path, false, "")
}

func (t *ThreeDimensionalGeoVolumes) reverseProxy(w http.ResponseWriter, r *http.Request, path string,
	prefer204 bool, contentTypeOverwrite string) {

	target, err := url.Parse(t.engine.Config.OgcAPI.GeoVolumes.TileServer.String() + path)
	if err != nil {
		log.Printf("invalid target url, can't proxy tiles: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	t.engine.ReverseProxy(w, r, target, prefer204, contentTypeOverwrite)
}

func (t *ThreeDimensionalGeoVolumes) idToCollection(cid string) (*engine.GeoSpatialCollection, error) {
	for _, collection := range t.engine.Config.OgcAPI.GeoVolumes.Collections {
		if collection.ID == cid {
			return &collection, nil
		}
	}
	return nil, errors.New("no matching collection found")
}
