package tiles

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/common/geospatial"
	"github.com/go-chi/chi/v5"
)

const (
	templatesDir       = "ogc/tiles/templates/"
	tilesPath          = "/tiles"
	tileMatrixSetsPath = "/tileMatrixSets"
	defaultTilesTmpl   = "{tms}/{z}/{x}/{y}.pbf"
)

type Tiles struct {
	engine *engine.Engine
}

func NewTiles(e *engine.Engine, router *chi.Mux) *Tiles {
	tilesBreadcrumbs := []engine.Breadcrumb{
		{
			Name: "Tiles",
			Path: "tiles",
		},
	}
	tileMatrixSetsBreadcrumbs := []engine.Breadcrumb{
		{
			Name: "Tile Matrix Sets",
			Path: "tileMatrixSets",
		},
	}

	e.RenderTemplates(tilesPath,
		tilesBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"tiles.go.json"),
		engine.NewTemplateKey(templatesDir+"tiles.go.html"))
	e.RenderTemplates(tileMatrixSetsPath,
		tileMatrixSetsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"tileMatrixSets.go.json"),
		engine.NewTemplateKey(templatesDir+"tileMatrixSets.go.html"))

	// TODO: i18n for srs pages
	renderTemplatesForSrs(e, "EuropeanETRS89_GRS80Quad_Draft", tilesBreadcrumbs, tileMatrixSetsBreadcrumbs)
	renderTemplatesForSrs(e, "NetherlandsRDNewQuad", tilesBreadcrumbs, tileMatrixSetsBreadcrumbs)
	renderTemplatesForSrs(e, "WebMercatorQuad", tilesBreadcrumbs, tileMatrixSetsBreadcrumbs)

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

func renderTemplatesForSrs(e *engine.Engine, srs string, tilesBreadcrumbs []engine.Breadcrumb, tileMatrixSetsBreadcrumbs []engine.Breadcrumb) {
	tilesSrsBreadcrumbs := tilesBreadcrumbs
	tilesSrsBreadcrumbs = append(tilesSrsBreadcrumbs, []engine.Breadcrumb{
		{
			Name: srs,
			Path: "tiles/" + srs,
		},
	}...)
	tileMatrixSetsSrsBreadcrumbs := tileMatrixSetsBreadcrumbs
	tileMatrixSetsSrsBreadcrumbs = append(tileMatrixSetsSrsBreadcrumbs, []engine.Breadcrumb{
		{
			Name: srs,
			Path: "tileMatrixSets/" + srs,
		},
	}...)

	e.RenderTemplates(tileMatrixSetsPath+"/"+srs,
		tileMatrixSetsSrsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"tileMatrixSets/"+srs+".go.json"),
		engine.NewTemplateKey(templatesDir+"tileMatrixSets/"+srs+".go.html"))

	e.RenderTemplates(tilesPath+"/"+srs,
		tilesSrsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"tiles/"+srs+".go.json"),
		engine.NewTemplateKey(templatesDir+"tiles/"+srs+".go.html"))

	e.RenderTemplates(tilesPath+"/"+srs,
		tilesSrsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"tiles/"+srs+".go.tilejson"))
}

func (t *Tiles) TileMatrixSets() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKeyWithLanguage(templatesDir+"tileMatrixSets.go."+t.engine.CN.NegotiateFormat(r), t.engine.CN.NegotiateLanguage(r))
		t.engine.ServePage(w, r, key)
	}
}

func (t *Tiles) TileMatrixSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tileMatrixSetID := chi.URLParam(r, "tileMatrixSetId")
		key := engine.NewTemplateKeyWithLanguage(templatesDir+"tileMatrixSets/"+tileMatrixSetID+".go."+t.engine.CN.NegotiateFormat(r), t.engine.CN.NegotiateLanguage(r))
		t.engine.ServePage(w, r, key)
	}
}

func (t *Tiles) TilesetsList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKeyWithLanguage(templatesDir+"tiles.go."+t.engine.CN.NegotiateFormat(r), t.engine.CN.NegotiateLanguage(r))
		t.engine.ServePage(w, r, key)
	}
}

func (t *Tiles) Tileset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tileMatrixSetID := chi.URLParam(r, "tileMatrixSetId")
		key := engine.NewTemplateKeyWithLanguage(templatesDir+"tiles/"+tileMatrixSetID+".go."+t.engine.CN.NegotiateFormat(r), t.engine.CN.NegotiateLanguage(r))
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
		} else {
			tileCol = tileCol[:len(tileCol)-4] // remove .pbf extension
		}

		// ogc spec is (default) z/row/col but tileserver is z/col/row (z/x/y)
		replacer := strings.NewReplacer("{tms}", tileMatrixSetID, "{z}", tileMatrix, "{x}", tileCol, "{y}", tileRow)
		tilesTmpl := defaultTilesTmpl
		if t.engine.Config.OgcAPI.Tiles.URITemplateTiles != nil {
			tilesTmpl = *t.engine.Config.OgcAPI.Tiles.URITemplateTiles
		}
		path, _ := url.JoinPath("/", replacer.Replace(tilesTmpl))

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
