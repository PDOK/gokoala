package geovolumes

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PDOK/gokoala/config"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/common/geospatial"

	"github.com/go-chi/chi/v5"
)

type ThreeDimensionalGeoVolumes struct {
	engine *engine.Engine
}

func NewThreeDimensionalGeoVolumes(e *engine.Engine) *ThreeDimensionalGeoVolumes {
	_, err := url.ParseRequestURI(e.Config.OgcAPI.GeoVolumes.TileServer.String())
	if err != nil {
		log.Fatalf("invalid tileserver url provided: %v", err)
	}

	geoVolumes := &ThreeDimensionalGeoVolumes{
		engine: e,
	}

	// 3D Tiles
	e.Router.Get(geospatial.CollectionsPath+"/{3dContainerId}/3dtiles", geoVolumes.Tileset("tileset.json"))
	e.Router.Get(geospatial.CollectionsPath+"/{3dContainerId}/3dtiles/{explicitTileSet}.json", geoVolumes.ExplicitTileset())
	e.Router.Get(geospatial.CollectionsPath+"/{3dContainerId}/3dtiles/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())
	e.Router.Get(geospatial.CollectionsPath+"/{3dContainerId}/3dtiles/{tilePathPrefix}/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())

	// DTM/Quantized Mesh
	e.Router.Get(geospatial.CollectionsPath+"/{3dContainerId}/quantized-mesh", geoVolumes.Tileset("layer.json"))
	e.Router.Get(geospatial.CollectionsPath+"/{3dContainerId}/quantized-mesh/{explicitTileSet}.json", geoVolumes.ExplicitTileset())
	e.Router.Get(geospatial.CollectionsPath+"/{3dContainerId}/quantized-mesh/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())
	e.Router.Get(geospatial.CollectionsPath+"/{3dContainerId}/quantized-mesh/{tilePathPrefix}/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())

	// path '/3dtiles' or '/quantized-mesh' is preferred but optional when requesting the actual tiles/tileset.
	e.Router.Get(geospatial.CollectionsPath+"/{3dContainerId}/{explicitTileSet}.json", geoVolumes.ExplicitTileset())
	e.Router.Get(geospatial.CollectionsPath+"/{3dContainerId}/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())
	e.Router.Get(geospatial.CollectionsPath+"/{3dContainerId}/{tilePathPrefix}/{tileMatrix}/{tileRow}/{tileColAndSuffix}", geoVolumes.Tile())

	return geoVolumes
}

// Tileset serves tileset.json manifest in case of OGC 3D Tiles (= separate spec from OGC 3D GeoVolumes) requests or
// layer.json manifest in case of quantized mesh requests. Both requests will be proxied to the configured tileserver.
func (t *ThreeDimensionalGeoVolumes) Tileset(fileName string) http.HandlerFunc {
	if !strings.HasSuffix(fileName, ".json") {
		log.Fatalf("manifest should be a JSON file")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		t.tileSet(w, r, fileName)
	}
}

// ExplicitTileset serves OGC 3D Tiles manifest (= separate spec from OGC 3D GeoVolumes) or
// quantized mesh manifest. All requests will be proxied to the configured tileserver.
func (t *ThreeDimensionalGeoVolumes) ExplicitTileset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tileSetName := chi.URLParam(r, "explicitTileSet")
		if tileSetName == "" {
			engine.RenderProblem(engine.ProblemNotFound, w)
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
			engine.RenderProblem(engine.ProblemNotFound, w, err.Error())
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
		if collection.GeoVolumes != nil && collection.GeoVolumes.HasDTM() {
			// DTM has a specialized mediatype, although application/octet-stream will also work with Cesium
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
		engine.RenderProblem(engine.ProblemNotFound, w, err.Error())
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
		engine.RenderProblem(engine.ProblemServerError, w)
		return
	}
	t.engine.ReverseProxy(w, r, target, prefer204, contentTypeOverwrite)
}

func (t *ThreeDimensionalGeoVolumes) idToCollection(cid string) (*config.GeoSpatialCollection, error) {
	for _, collection := range t.engine.Config.OgcAPI.GeoVolumes.Collections {
		if collection.ID == cid {
			return &collection, nil
		}
	}
	return nil, errors.New("no matching collection found")
}
