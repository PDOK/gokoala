package tiles

import (
	"gokoala/engine"
	"gokoala/ogc/common/geospatial"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
)

const (
	templatesDir       = "ogc/tiles/templates/"
	tilesPath          = "/tiles"
	tileMatrixSetsPath = "/tileMatrixSets"
)

type Tiles struct {
	engine *engine.Engine
}

func NewTiles(e *engine.Engine, router *chi.Mux) *Tiles {
	e.RenderTemplates(tilesPath,
		engine.NewTemplateKey(templatesDir+"tiles.go.json"),
		engine.NewTemplateKey(templatesDir+"tiles.go.html"))
	e.RenderTemplates(tileMatrixSetsPath,
		engine.NewTemplateKey(templatesDir+"tileMatrixSets.go.json"),
		engine.NewTemplateKey(templatesDir+"tileMatrixSets.go.html"))

	renderTemplatesForSrs(e, "EuropeanETRS89_GRS80Quad_Draft")
	renderTemplatesForSrs(e, "NetherlandsRDNewQuad")
	renderTemplatesForSrs(e, "WebMercatorQuad")

	_, err := url.ParseRequestURI(e.Config.OgcAPI.Tiles.TileServer.String())
	if err != nil {
		log.Fatalf("invalid tileserver url provided: %v", err)
	}
	tiles := &Tiles{
		engine: e,
	}

	router.Get(tileMatrixSetsPath, tiles.TileMatrixSets())
	router.Get(tileMatrixSetsPath+"/{tileMatrixSetId}", tiles.TileMatrixSet())
	router.Get(tilesPath, tiles.TilesetsList())
	router.Get(tilesPath+"/{tileMatrixSetId}", tiles.Tileset())
	router.Get(tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile())
	router.Get(geospatial.CollectionsPath+"/{collectionId}/tiles", tiles.CollectionContent())

	return tiles
}

func renderTemplatesForSrs(e *engine.Engine, srs string) {
	e.RenderTemplates(tileMatrixSetsPath+"/"+srs,
		engine.NewTemplateKey(templatesDir+"tileMatrixSets/"+srs+".go.json"),
		engine.NewTemplateKey(templatesDir+"tileMatrixSets/"+srs+".go.html"))

	e.RenderTemplates(tilesPath+"/"+srs,
		engine.NewTemplateKey(templatesDir+"tiles/"+srs+".go.json"),
		engine.NewTemplateKey(templatesDir+"tiles/"+srs+".go.html"))

	e.RenderTemplates(tilesPath+"/"+srs,
		engine.NewTemplateKey(templatesDir+"tiles/"+srs+".go.tilejson"))
}

func (t *Tiles) TileMatrixSets() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKey(templatesDir + "tileMatrixSets.go." + t.engine.CN.NegotiateFormat(r))
		t.engine.ServePage(w, r, key)
	}
}

func (t *Tiles) TileMatrixSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tileMatrixSetID := chi.URLParam(r, "tileMatrixSetId")
		key := engine.NewTemplateKey(templatesDir + "tileMatrixSets/" + tileMatrixSetID + ".go." + t.engine.CN.NegotiateFormat(r))
		t.engine.ServePage(w, r, key)
	}
}

func (t *Tiles) TilesetsList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKey(templatesDir + "tiles.go." + t.engine.CN.NegotiateFormat(r))
		t.engine.ServePage(w, r, key)
	}
}

func (t *Tiles) Tileset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tileMatrixSetID := chi.URLParam(r, "tileMatrixSetId")
		key := engine.NewTemplateKey(templatesDir + "tiles/" + tileMatrixSetID + ".go." + t.engine.CN.NegotiateFormat(r))
		t.engine.ServePage(w, r, key)
	}
}

// Tile reverse proxy to Azure Blob, assumes blob bucket/container is public
func (t *Tiles) Tile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tileMatrixSetID := chi.URLParam(r, "tileMatrixSetId")
		tileMatrix := chi.URLParam(r, "tileMatrix")
		tileRow := chi.URLParam(r, "tileRow")
		tileCol := chi.URLParam(r, "tileCol")

		// We support content negotiation using Accept header and ?f= param, but also
		// using the .pbf extension. This is for backwards compatibility.
		if !strings.HasSuffix(tileCol, ".pbf") {
			if t.engine.CN.NegotiateFormat(r) != "mvt" {
				http.Error(w, "Specify tile format. Currently only"+
					" Mapbox Vector Tiles (?f=mvt) tiles are supported", http.StatusBadRequest)
				return
			}
			tileCol += ".pbf"
		}

		path, _ := url.JoinPath("/", tileMatrixSetID, tileMatrix, tileRow, tileCol)

		target, err := url.Parse(t.engine.Config.OgcAPI.Tiles.TileServer.String() + path)
		if err != nil {
			log.Printf("invalid target url, can't proxy tiles: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		t.engine.ReverseProxy(w, r, target, true, "application/vnd.mapbox-vector-tile")
	}
}

func (t *Tiles) CollectionContent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")

		// TODO: not implemented, since we don't (yet) support tile collections
		log.Printf("TODO: return tiles for collection %s", collectionID)
	}
}
