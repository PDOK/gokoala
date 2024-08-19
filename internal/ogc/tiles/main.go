package tiles

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	g "github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	"github.com/go-chi/chi/v5"
)

const (
	templatesDir            = "internal/ogc/tiles/templates/"
	tilesPath               = "/tiles"
	tilesLocalPath          = "tiles/"
	tileMatrixSetsPath      = "/tileMatrixSets"
	tileMatrixSetsLocalPath = "tileMatrixSets/"
	defaultTilesTmpl        = "{tms}/{z}/{x}/{y}." + engine.FormatMVTAlternative
)

var (
	tilesBreadcrumbs = []engine.Breadcrumb{
		{
			Name: "Tiles",
			Path: "tiles",
		},
	}
	tileMatrixSetsBreadcrumbs = []engine.Breadcrumb{
		{
			Name: "Tile Matrix Sets",
			Path: "tileMatrixSets",
		},
	}
)

type tilesTemplateData struct {
	// Tiles top-level or collection-level tiles config
	config.Tiles

	// baseURL part of the url prefixing /tiles
	BaseURL string
}

type Tiles struct {
	engine *engine.Engine
}

func NewTiles(e *engine.Engine) *Tiles {
	tiles := &Tiles{engine: e}

	// TileMatrixSets
	e.Router.Get(tileMatrixSetsPath, tiles.TileMatrixSets())
	e.Router.Get(tileMatrixSetsPath+"/{tileMatrixSetId}", tiles.TileMatrixSet())

	// Top-level tiles
	if e.Config.OgcAPI.Tiles.DatasetTiles != nil {
		templateData := tilesTemplateData{
			*e.Config.OgcAPI.Tiles.DatasetTiles,
			e.Config.BaseURL.String(),
		}
		renderTilesTemplates(e, templateData, tilesBreadcrumbs, tileMatrixSetsBreadcrumbs)

		e.Router.Get(tilesPath, tiles.TilesetsList())
		e.Router.Get(tilesPath+"/{tileMatrixSetId}", tiles.Tileset())
		e.Router.Head(tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile())
		e.Router.Get(tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile())
	}
	// Collection-level tiles
	for _, coll := range e.Config.OgcAPI.Tiles.Collections {
		if coll.Tiles == nil {
			continue
		}
		templateData := tilesTemplateData{
			coll.Tiles.GeoDataTiles,
			fmt.Sprintf("%s/%s/%s", e.Config.BaseURL.String(), g.CollectionsPath, coll.ID),
		}
		renderTilesTemplates(e, templateData, tilesBreadcrumbs, tileMatrixSetsBreadcrumbs)

		e.Router.Get(g.CollectionsPath+"/{collectionId}"+tilesPath, tiles.TilesCollection())
		e.Router.Get(g.CollectionsPath+"/{collectionId}"+tilesPath+"/{tileMatrixSetId}", tiles.TilesCollection())
		e.Router.Head(g.CollectionsPath+"/{collectionId}"+tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile())
		e.Router.Get(g.CollectionsPath+"/{collectionId}"+tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile())
	}
	return tiles
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

// Tile reverse proxy to tileserver/object storage. Assumes the backing resources is publicly accessible.
func (t *Tiles) Tile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tileMatrixSetID := chi.URLParam(r, "tileMatrixSetId")
		tileMatrix := chi.URLParam(r, "tileMatrix")
		tileRow := chi.URLParam(r, "tileRow")
		tileCol := chi.URLParam(r, "tileCol")

		// We support content negotiation using Accept header and ?f= param, but also
		// using the .pbf extension. This is for backwards compatibility.
		if !strings.HasSuffix(tileCol, "."+engine.FormatMVTAlternative) {
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
		if t.engine.Config.OgcAPI.Tiles.DatasetTiles.URITemplateTiles != nil {
			tilesTmpl = *t.engine.Config.OgcAPI.Tiles.DatasetTiles.URITemplateTiles
		}
		path, _ := url.JoinPath("/", replacer.Replace(tilesTmpl))

		target, err := url.Parse(t.engine.Config.OgcAPI.Tiles.DatasetTiles.TileServer.String() + path)
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

func renderTilesTemplates(e *engine.Engine, data tilesTemplateData,
	tilesBreadcrumbs []engine.Breadcrumb, tileMatrixSetsBreadcrumbs []engine.Breadcrumb) {

	e.RenderTemplatesWithParamsAndValidate(tilesPath,
		data,
		tilesBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"tiles.go.json"),
		engine.NewTemplateKey(templatesDir+"tiles.go.html"))
	e.RenderTemplatesWithParamsAndValidate(tileMatrixSetsPath,
		data,
		tileMatrixSetsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"tileMatrixSets.go.json"),
		engine.NewTemplateKey(templatesDir+"tileMatrixSets.go.html"))

	renderTilesTemplatesForSrs(e, "EuropeanETRS89_LAEAQuad", data, tilesBreadcrumbs, tileMatrixSetsBreadcrumbs)
	renderTilesTemplatesForSrs(e, "NetherlandsRDNewQuad", data, tilesBreadcrumbs, tileMatrixSetsBreadcrumbs)
	renderTilesTemplatesForSrs(e, "WebMercatorQuad", data, tilesBreadcrumbs, tileMatrixSetsBreadcrumbs)
}

func renderTilesTemplatesForSrs(e *engine.Engine, srs string, data tilesTemplateData,
	tilesBreadcrumbs []engine.Breadcrumb, tileMatrixSetsBreadcrumbs []engine.Breadcrumb) {

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

	e.RenderTemplatesWithParamsAndValidate(tileMatrixSetsPath+"/"+srs,
		data,
		tileMatrixSetsSrsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+tileMatrixSetsLocalPath+srs+".go.json"),
		engine.NewTemplateKey(templatesDir+tileMatrixSetsLocalPath+srs+".go.html"))

	e.RenderTemplatesWithParamsAndValidate(tilesPath+"/"+srs,
		data,
		tilesSrsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+tilesLocalPath+srs+".go.json"),
		engine.NewTemplateKey(templatesDir+tilesLocalPath+srs+".go.html"))

	e.RenderTemplatesWithParamsAndValidate(tilesPath+"/"+srs,
		data,
		tilesSrsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+tilesLocalPath+srs+".go.tilejson"))
}
