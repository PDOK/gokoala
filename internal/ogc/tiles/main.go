package tiles

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/engine/util"
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
	collectionsCrumb        = "collections/"
	tilesCrumbTitle         = "Tiles"
)

var (
	// When adding a new projection also add corresponding templates
	allProjections = map[string]string{
		"EPSG:28992": "NetherlandsRDNewQuad",
		"EPSG:3035":  "EuropeanETRS89_LAEAQuad",
		"EPSG:3857":  "WebMercatorQuad",
	}

	tilesBreadcrumbs = []engine.Breadcrumb{
		{
			Name: tilesCrumbTitle,
			Path: "tiles",
		},
	}
	tileMatrixSetsBreadcrumbs = []engine.Breadcrumb{
		{
			Name: "Tile Matrix Sets",
			Path: "tileMatrixSets",
		},
	}
	collectionsBreadcrumb = []engine.Breadcrumb{
		{
			Name: "Collections",
			Path: "collections",
		},
	}
)

type templateData struct {
	// Tiles top-level or collection-level tiles config
	config.Tiles

	// BaseURL part of the url prefixing "/tiles"
	BaseURL string

	// All supported projections for (vector) tiles by GoKoala
	AllProjections map[string]any
}

type Tiles struct {
	engine *engine.Engine
}

func NewTiles(e *engine.Engine) *Tiles {
	tiles := &Tiles{engine: e}

	// TileMatrixSets
	renderTileMatrixTemplates(e)
	e.Router.Get(tileMatrixSetsPath, tiles.TileMatrixSets())
	e.Router.Get(tileMatrixSetsPath+"/{tileMatrixSetId}", tiles.TileMatrixSet())

	// Top-level tiles (dataset tiles in OGC spec)
	if e.Config.OgcAPI.Tiles.DatasetTiles != nil {
		renderTilesTemplates(e, nil, templateData{
			*e.Config.OgcAPI.Tiles.DatasetTiles,
			e.Config.BaseURL.String(),
			util.Cast(allProjections),
		})
		e.Router.Get(tilesPath, tiles.TilesetsList())
		e.Router.Get(tilesPath+"/{tileMatrixSetId}", tiles.Tileset())
		e.Router.Head(tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile(*e.Config.OgcAPI.Tiles.DatasetTiles))
		e.Router.Get(tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile(*e.Config.OgcAPI.Tiles.DatasetTiles))
	}

	// Collection-level tiles (geodata tiles in OGC spec)
	for _, coll := range e.Config.OgcAPI.Tiles.Collections {
		if coll.Tiles == nil {
			continue
		}
		renderTilesTemplates(e, &coll, templateData{
			coll.Tiles.GeoDataTiles,
			e.Config.BaseURL.String() + g.CollectionsPath + "/" + coll.ID,
			util.Cast(allProjections),
		})
		e.Router.Get(g.CollectionsPath+"/{collectionId}"+tilesPath, tiles.TilesetsListForCollection())
		e.Router.Get(g.CollectionsPath+"/{collectionId}"+tilesPath+"/{tileMatrixSetId}", tiles.TilesetForCollection())
		e.Router.Head(g.CollectionsPath+"/{collectionId}"+tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile(coll.Tiles.GeoDataTiles))
		e.Router.Get(g.CollectionsPath+"/{collectionId}"+tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile(coll.Tiles.GeoDataTiles))
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

func (t *Tiles) TilesetsListForCollection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")
		key := engine.NewTemplateKeyWithNameAndLanguage(templatesDir+"tiles.go."+t.engine.CN.NegotiateFormat(r), collectionID, t.engine.CN.NegotiateLanguage(w, r))
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

func (t *Tiles) TilesetForCollection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")
		tileMatrixSetID := chi.URLParam(r, "tileMatrixSetId")
		key := engine.NewTemplateKeyWithNameAndLanguage(templatesDir+tilesLocalPath+tileMatrixSetID+".go."+t.engine.CN.NegotiateFormat(r), collectionID, t.engine.CN.NegotiateLanguage(w, r))
		t.engine.ServePage(w, r, key)
	}
}

// Tile reverse proxy to tileserver/object storage. Assumes the backing resources is publicly accessible.
func (t *Tiles) Tile(tileConfig config.Tiles) http.HandlerFunc {
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
		if tileConfig.URITemplateTiles != nil {
			tilesTmpl = *tileConfig.URITemplateTiles
		}
		path, _ := url.JoinPath("/", replacer.Replace(tilesTmpl))

		target, err := url.Parse(tileConfig.TileServer.String() + path)
		if err != nil {
			log.Printf("invalid target url, can't proxy tiles: %v", err)
			engine.RenderProblem(engine.ProblemServerError, w)
			return
		}
		t.engine.ReverseProxy(w, r, target, true, engine.MediaTypeMVT)
	}
}

func renderTileMatrixTemplates(e *engine.Engine) {
	e.RenderTemplates(tileMatrixSetsPath,
		tileMatrixSetsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"tileMatrixSets.go.json"),
		engine.NewTemplateKey(templatesDir+"tileMatrixSets.go.html"))

	for _, projection := range allProjections {
		breadcrumbs := tileMatrixSetsBreadcrumbs
		breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
			{
				Name: projection,
				Path: tileMatrixSetsLocalPath + projection,
			},
		}...)

		e.RenderTemplates(tileMatrixSetsPath+"/"+projection,
			breadcrumbs,
			engine.NewTemplateKey(templatesDir+tileMatrixSetsLocalPath+projection+".go.json"),
			engine.NewTemplateKey(templatesDir+tileMatrixSetsLocalPath+projection+".go.html"))
	}
}

func renderTilesTemplates(e *engine.Engine, collection *config.GeoSpatialCollection, data templateData) {

	var breadcrumbs []engine.Breadcrumb
	collectionID := ""
	if collection != nil {
		collectionID = collection.ID

		breadcrumbs = collectionsBreadcrumb
		breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
			{
				Name: getCollectionTitle(collectionID, collection.Metadata),
				Path: collectionsCrumb + collectionID,
			},
			{
				Name: tilesCrumbTitle,
				Path: collectionsCrumb + collectionID + tilesPath,
			},
		}...)
	} else {
		breadcrumbs = tilesBreadcrumbs
	}

	e.RenderTemplatesWithParamsAndValidate(tilesPath,
		data,
		breadcrumbs,
		engine.NewTemplateKeyWithName(templatesDir+"tiles.go.json", collectionID),
		engine.NewTemplateKeyWithName(templatesDir+"tiles.go.html", collectionID))

	// Now render metadata bout tiles per projection/SRS.
	for _, projection := range allProjections {
		projectionBreadcrumbs := breadcrumbs
		projectionBreadcrumbs = append(projectionBreadcrumbs, []engine.Breadcrumb{
			{
				Name: projection,
				Path: tilesLocalPath + projection,
			},
		}...)

		path := tilesPath + "/" + projection
		if collection != nil {
			path = "/" + collectionID + tilesPath + "/" + projection
		}
		e.RenderTemplatesWithParamsAndValidate(path,
			data,
			projectionBreadcrumbs,
			engine.NewTemplateKeyWithName(templatesDir+tilesLocalPath+projection+".go.json", collectionID),
			engine.NewTemplateKeyWithName(templatesDir+tilesLocalPath+projection+".go.html", collectionID))
		e.RenderTemplatesWithParamsAndValidate(path,
			data,
			projectionBreadcrumbs,
			engine.NewTemplateKeyWithName(templatesDir+tilesLocalPath+projection+".go.tilejson", collectionID))
	}
}

func getCollectionTitle(collectionID string, metadata *config.GeoSpatialCollectionMetadata) string {
	if metadata != nil && metadata.Title != nil {
		return *metadata.Title
	}
	return collectionID
}
