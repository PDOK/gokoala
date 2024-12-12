package tiles

import (
	"errors"
	"fmt"
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

	// All supported projections by GoKoala (for tiles)
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
			util.Cast(config.AllTileProjections),
		})
		e.Router.Get(tilesPath, tiles.TilesetsList())
		e.Router.Get(tilesPath+"/{tileMatrixSetId}", tiles.Tileset())
		e.Router.Head(tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile(*e.Config.OgcAPI.Tiles.DatasetTiles))
		e.Router.Get(tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.Tile(*e.Config.OgcAPI.Tiles.DatasetTiles))
	}

	// Collection-level tiles (geodata tiles in OGC spec)
	geoDataTiles := map[string]config.Tiles{}
	for _, coll := range e.Config.OgcAPI.Tiles.Collections {
		if coll.Tiles == nil {
			continue
		}
		renderTilesTemplates(e, &coll, templateData{
			coll.Tiles.GeoDataTiles,
			e.Config.BaseURL.String() + g.CollectionsPath + "/" + coll.ID,
			util.Cast(config.AllTileProjections),
		})
		geoDataTiles[coll.ID] = coll.Tiles.GeoDataTiles
	}
	if len(geoDataTiles) != 0 {
		e.Router.Get(g.CollectionsPath+"/{collectionId}"+tilesPath, tiles.TilesetsListForCollection())
		e.Router.Get(g.CollectionsPath+"/{collectionId}"+tilesPath+"/{tileMatrixSetId}", tiles.TilesetForCollection())
		e.Router.Head(g.CollectionsPath+"/{collectionId}"+tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.TileForCollection(geoDataTiles))
		e.Router.Get(g.CollectionsPath+"/{collectionId}"+tilesPath+"/{tileMatrixSetId}/{tileMatrix}/{tileRow}/{tileCol}", tiles.TileForCollection(geoDataTiles))
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

// Tile reverse proxy to configured tileserver/object storage. Assumes the backing resource is publicly accessible.
func (t *Tiles) Tile(tilesConfig config.Tiles) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tileMatrixSetID := chi.URLParam(r, "tileMatrixSetId")
		tileMatrix := chi.URLParam(r, "tileMatrix")
		tileRow := chi.URLParam(r, "tileRow")
		tileCol, err := getTileColumn(r, t.engine.CN.NegotiateFormat(r))
		if err != nil {
			engine.RenderProblemAndLog(engine.ProblemBadRequest, w, err, err.Error())
			return
		}

		target, err := createTilesURL(tileMatrixSetID, tileMatrix, tileCol, tileRow, tilesConfig)
		if err != nil {
			engine.RenderProblemAndLog(engine.ProblemServerError, w, err)
			return
		}
		t.engine.ReverseProxy(w, r, target, true, engine.MediaTypeMVT)
	}
}

// TileForCollection reverse proxy to configured tileserver/object storage for tiles within a given collection.
// Assumes the backing resource is publicly accessible.
func (t *Tiles) TileForCollection(tilesConfigByCollection map[string]config.Tiles) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")
		tileMatrixSetID := chi.URLParam(r, "tileMatrixSetId")
		tileMatrix := chi.URLParam(r, "tileMatrix")
		tileRow := chi.URLParam(r, "tileRow")
		tileCol, err := getTileColumn(r, t.engine.CN.NegotiateFormat(r))
		if err != nil {
			engine.RenderProblemAndLog(engine.ProblemBadRequest, w, err, err.Error())
			return
		}

		tilesConfig, ok := tilesConfigByCollection[collectionID]
		if !ok {
			err = fmt.Errorf("no tiles available for collection: %s", collectionID)
			engine.RenderProblemAndLog(engine.ProblemNotFound, w, err, err.Error())
			return
		}
		target, err := createTilesURL(tileMatrixSetID, tileMatrix, tileCol, tileRow, tilesConfig)
		if err != nil {
			engine.RenderProblemAndLog(engine.ProblemServerError, w, err)
			return
		}
		t.engine.ReverseProxy(w, r, target, true, engine.MediaTypeMVT)
	}
}

func getTileColumn(r *http.Request, format string) (string, error) {
	tileCol := chi.URLParam(r, "tileCol")

	// We support content negotiation using Accept header and ?f= param, but also
	// using the .pbf extension. This is for backwards compatibility.
	if !strings.HasSuffix(tileCol, "."+engine.FormatMVTAlternative) {
		// if no format is specified, default to mvt
		if f := strings.Replace(format, engine.FormatJSON, engine.FormatMVT, 1); f != engine.FormatMVT && f != engine.FormatMVTAlternative {
			return "", errors.New("specify tile format. Currently only Mapbox Vector Tiles (?f=mvt) tiles are supported")
		}
	} else {
		tileCol = tileCol[:len(tileCol)-4] // remove .pbf extension
	}
	return tileCol, nil
}

func createTilesURL(tileMatrixSetID string, tileMatrix string, tileCol string,
	tileRow string, tilesCfg config.Tiles) (*url.URL, error) {

	tilesTmpl := defaultTilesTmpl
	if tilesCfg.URITemplateTiles != nil {
		tilesTmpl = *tilesCfg.URITemplateTiles
	}
	// OGC spec is (default) z/row/col but tileserver is z/col/row (z/x/y)
	replacer := strings.NewReplacer("{tms}", tileMatrixSetID, "{z}", tileMatrix, "{x}", tileCol, "{y}", tileRow)
	path, _ := url.JoinPath("/", replacer.Replace(tilesTmpl))

	target, err := url.Parse(tilesCfg.TileServer.String() + path)
	if err != nil {
		return nil, fmt.Errorf("invalid target url, can't proxy tiles: %w", err)
	}
	return target, nil
}

func renderTileMatrixTemplates(e *engine.Engine) {
	e.RenderTemplates(tileMatrixSetsPath,
		tileMatrixSetsBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"tileMatrixSets.go.json"),
		engine.NewTemplateKey(templatesDir+"tileMatrixSets.go.html"))

	for _, projection := range config.AllTileProjections {
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
	path := tilesPath
	collectionID := ""
	if collection != nil {
		collectionID = collection.ID
		path = g.CollectionsPath + "/" + collectionID + tilesPath

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

	e.RenderTemplatesWithParams(path,
		data,
		breadcrumbs,
		engine.NewTemplateKeyWithName(templatesDir+"tiles.go.json", collectionID),
		engine.NewTemplateKeyWithName(templatesDir+"tiles.go.html", collectionID))

	// Now render metadata about tiles per projection/SRS.
	for _, projection := range config.AllTileProjections {
		path = tilesPath + "/" + projection
		projectionBreadcrumbs := breadcrumbs

		if collection != nil {
			projectionBreadcrumbs = append(projectionBreadcrumbs, []engine.Breadcrumb{
				{
					Name: projection,
					Path: collectionsCrumb + collectionID + path,
				},
			}...)
			path = g.CollectionsPath + "/" + collectionID + tilesPath + "/" + projection
		} else {
			projectionBreadcrumbs = append(projectionBreadcrumbs, []engine.Breadcrumb{
				{
					Name: projection,
					Path: path,
				},
			}...)
		}
		e.RenderTemplatesWithParams(path,
			data,
			projectionBreadcrumbs,
			engine.NewTemplateKeyWithName(templatesDir+tilesLocalPath+projection+".go.json", collectionID),
			engine.NewTemplateKeyWithName(templatesDir+tilesLocalPath+projection+".go.html", collectionID))
		e.RenderTemplatesWithParams(path,
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
