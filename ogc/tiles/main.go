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
	templatesDir            = "ogc/tiles/templates/"
	tilesPath               = "/tiles"
	tilesLocalPath          = "tiles/"
	tileMatrixSetsPath      = "/tileMatrixSets"
	tileMatrixSetsLocalPath = "tileMatrixSets/"
	defaultTilesTmpl        = "{tms}/{z}/{x}/{y}." + engine.FormatMVTAlternative
)

type Tiles struct {
	engine *engine.Engine
}

func NewTiles(e *engine.Engine) *Tiles {
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

	renderTemplatesForSrs(e, "EuropeanETRS89_LAEAQuad", tilesBreadcrumbs, tileMatrixSetsBreadcrumbs)
	renderTemplatesForSrs(e, "NetherlandsRDNewQuad", tilesBreadcrumbs, tileMatrixSetsBreadcrumbs)
	renderTemplatesForSrs(e, "WebMercatorQuad", tilesBreadcrumbs, tileMatrixSetsBreadcrumbs)

	_, err := url.ParseRequestURI(e.Config.OgcAPI.Tiles.TileServer.String())
	if err != nil {
		log.Fatalf("invalid tileserver url provided: %v", err)
	}
	tiles := &Tiles{
		engine: e,
	}

	e.Router.Get(tileMatrixSetsPath, tiles.TileMatrixSets())
	e.Router.Get(tileMatrixSetsPath+"/{tileMatrixSetId}", tiles.TileMatrixSet())
	e.Router.Get(tilesPath, tiles.TilesetsList())
	e.Router.Get(tilesPath+"/{tileMatrixSetId}", tiles.Tileset())
	e.Router.Head(tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile())
	e.Router.Get(tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile())
	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/tiles", tiles.TilesCollection())

	return tiles
}

func renderTemplatesForSrs(e *engine.Engine, srs string, tilesBreadcrumbs []engine.Breadcrumb, tileMatrixSetsBreadcrumbs []engine.Breadcrumb) {
	tilesSrsBreadcrumbs := tilesBreadcrumbs
	tilesSrsBreadcrumbs = append(tilesSrsBreadcrumbs, []engine.Breadcrumb{
		{
			Name: srs,
			Path: tilesLocalPath + srs,
		},
	}...)
	tileMatrixSetsSrsBreadcrumbs := tileMatrixSetsBreadcrumbs
	tileMatrixSetsSrsBreadcrumbs = append(tileMatrixSetsSrsBreadcrumbs, []engine.Breadcrumb{
		{
			Name: srs,
			Path: tileMatrixSetsLocalPath + srs,
		},
	}...)

	e.RenderTemplates(tileMatrixSetsPath+"/"+srs,
		tileMatrixSetsSrsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+tileMatrixSetsLocalPath+srs+".go.json"),
		engine.NewTemplateKey(templatesDir+tileMatrixSetsLocalPath+srs+".go.html"))

	e.RenderTemplates(tilesPath+"/"+srs,
		tilesSrsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+tilesLocalPath+srs+".go.json"),
		engine.NewTemplateKey(templatesDir+tilesLocalPath+srs+".go.html"))

	e.RenderTemplates(tilesPath+"/"+srs,
		tilesSrsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+tilesLocalPath+srs+".go.tilejson"))
}

func (t *Tiles) TileMatrixSets() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKeyWithLanguage(templatesDir+"tileMatrixSets.go."+t.engine.CN.NegotiateFormat(r), t.engine.CN.NegotiateLanguage(w, r))
		t.engine.ServePage(w, r, key)
	}
}

func (t *Tiles) TileMatrixSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tileMatrixSetID := chi.URLParam(r, "tileMatrixSetId")
		key := engine.NewTemplateKeyWithLanguage(templatesDir+tileMatrixSetsLocalPath+tileMatrixSetID+".go."+t.engine.CN.NegotiateFormat(r), t.engine.CN.NegotiateLanguage(w, r))
		t.engine.ServePage(w, r, key)
	}
}

func (t *Tiles) TilesetsList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKeyWithLanguage(templatesDir+"tiles.go."+t.engine.CN.NegotiateFormat(r), t.engine.CN.NegotiateLanguage(w, r))
		t.engine.ServePage(w, r, key)
	}
}

func (t *Tiles) Tileset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tileMatrixSetID := chi.URLParam(r, "tileMatrixSetId")
		key := engine.NewTemplateKeyWithLanguage(templatesDir+tilesLocalPath+tileMatrixSetID+".go."+t.engine.CN.NegotiateFormat(r), t.engine.CN.NegotiateLanguage(w, r))
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
			// if no format is specified, default to mvt
			if format := strings.Replace(t.engine.CN.NegotiateFormat(r), engine.FormatJSON, engine.FormatMVT, 1); format != engine.FormatMVT && format != engine.FormatMVTAlternative {
				engine.RenderProblem(engine.ProblemBadRequest, w, "Specify tile format. Currently only Mapbox Vector Tiles (?f=mvt) tiles are supported")
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
			engine.RenderProblem(engine.ProblemServerError, w)
			return
		}
		t.engine.ReverseProxy(w, r, target, true, engine.MediaTypeMVT)
	}
}

func (t *Tiles) TilesCollection(_ ...any) http.HandlerFunc {
	return func(_ http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")

		// TODO: not implemented, since we don't (yet) support tile collections
		log.Printf("TODO: return tiles for collection %s", collectionID)
	}
}
