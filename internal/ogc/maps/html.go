package maps

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/twpayne/go-geom"
)

const (
	collectionsCrumb = "collections/"
)

var (
	mapsHTMLKey = engine.NewTemplateKey(templatesDir + "maps.go.html")
	mapsMDKey   = engine.NewTemplateKey(templatesDir + "maps.go.md")

	mapsByFormat = map[string]engine.TemplateKey{
		engine.FormatHTML:     mapsHTMLKey,
		engine.FormatMarkdown: mapsMDKey,
	}

	collectionsBreadcrumb = []engine.Breadcrumb{
		{
			Name: "Collections",
			Path: "collections",
		},
	}
)

type htmlMaps struct {
	engine *engine.Engine
}

type mapsPage struct {
	CollectionID string
	StyleID      string
	Title        string

	CRS         string
	Bbox        string
	Transparent bool
	MMPerPixel  float64
}

// newHTMLMaps render the requested as HTML or Markdown
func newHTMLMaps(e *engine.Engine) *htmlMaps {
	e.ParseTemplate(mapsHTMLKey)

	e.ParseTemplate(mapsMDKey)

	return &htmlMaps{
		engine: e,
	}
}

func (hm *htmlMaps) maps(w http.ResponseWriter, r *http.Request, format string, collection config.MapsCollection,
	request MapRequest) {

	bbox := request.bbox
	if r.URL.Query().Get(BboxParam) == "" {
		bbox = nil
	}
	style := collection.GetStyle(request.style)
	breadcrumbs, pageContent := hm.toViewerPage(collection, style, request.crs, bbox)
	pageContent.Transparent = request.transparent

	hm.engine.RenderAndServe(w, r,
		engine.ExpandTemplateKey(mapsByFormat[format], hm.engine.CN.NegotiateLanguage(w, r)),
		pageContent, breadcrumbs, nil)
}

func (hm *htmlMaps) toViewerPage(collection config.MapsCollection, style *config.MapStyle,
	crs domain.ContentCrs, bbox *geom.Bounds) ([]engine.Breadcrumb, *mapsPage) {
	title := getCollectionTitle(collection.GetID(), collection.GetMetadata())
	breadcrumbs := collectionsBreadcrumb
	breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
		{
			Name: title,
			Path: collectionsCrumb + collection.GetID(),
		},
		{
			Name: "Maps",
			Path: collectionsCrumb + collection.GetID() + "/map",
		},
	}...)

	page := &mapsPage{
		CollectionID: collection.GetID(),
		Title:        title,
		CRS:          string(crs),
		Bbox:         hm.getBbox(collection, bbox),
	}
	if style != nil {
		page.StyleID = style.ID
	}

	return breadcrumbs, page
}

func (hm *htmlMaps) getBbox(collection config.MapsCollection, bbox *geom.Bounds) string {
	if bbox != nil {
		return strings.Join([]string{
			strconv.FormatFloat(bbox.Min(0), 'f', -1, 64),
			strconv.FormatFloat(bbox.Min(1), 'f', -1, 64),
			strconv.FormatFloat(bbox.Max(0), 'f', -1, 64),
			strconv.FormatFloat(bbox.Max(1), 'f', -1, 64),
		}, ",")
	}
	if len(collection.Bbox) == 4 {
		return strings.Join(collection.Bbox, ",")
	}

	return strings.Join(hm.engine.Config.OgcAPI.Maps.DefaultMap.Bbox, ",")
}

func getCollectionTitle(collectionID string, metadata *config.GeoSpatialCollectionMetadata) string {
	if metadata != nil && metadata.Title != nil {
		return *metadata.Title
	}

	return collectionID
}
