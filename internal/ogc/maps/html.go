package maps

import (
	"net/http"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
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

type mapsPage struct{}

// newHTMLMaps render the requested as HTML or Markdown
func newHTMLMaps(e *engine.Engine) *htmlMaps {
	e.ParseTemplate(mapsHTMLKey)

	e.ParseTemplate(mapsMDKey)

	return &htmlMaps{
		engine: e,
	}
}

func (hm *htmlMaps) maps(w http.ResponseWriter, r *http.Request, format string, collection config.MapsCollection,
	_ MapRequest) {

	breadcrumbs, pageContent := hm.toViewerPage(collection)

	hm.engine.RenderAndServe(w, r,
		engine.ExpandTemplateKey(mapsByFormat[format], hm.engine.CN.NegotiateLanguage(w, r)),
		pageContent, breadcrumbs, nil)
}

func (hm *htmlMaps) toViewerPage(collection config.MapsCollection) ([]engine.Breadcrumb, *mapsPage) {
	breadcrumbs := collectionsBreadcrumb
	breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
		{
			Name: getCollectionTitle(collection.GetID(), collection.GetMetadata()),
			Path: collectionsCrumb + collection.GetID(),
		},
		{
			Name: "Maps",
			Path: collectionsCrumb + collection.GetID() + "/maps",
		},
	}...)
	return breadcrumbs, &mapsPage{}
}

func getCollectionTitle(collectionID string, metadata *config.GeoSpatialCollectionMetadata) string {
	if metadata != nil && metadata.Title != nil {
		return *metadata.Title
	}

	return collectionID
}
