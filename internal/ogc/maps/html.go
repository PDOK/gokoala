package maps

import (
	"net/http"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/twpayne/go-geom"
)

var (
	mapsHTMLKey = engine.NewTemplateKey(templatesDir + "maps.go.html")
	mapsMDKey   = engine.NewTemplateKey(templatesDir + "maps.go.md")

	mapsByFormat = map[string]engine.TemplateKey{
		engine.FormatHTML:     mapsHTMLKey,
		engine.FormatMarkdown: mapsMDKey,
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

func (hm *htmlMaps) maps(w http.ResponseWriter, r *http.Request, format string,
	crs domain.ContentCrs, srid domain.SRID, bbox *geom.Bounds, width int, height int, err error) {
	pageContent, breadcrumbs := hm.toViewerPage()
	hm.engine.RenderAndServe(w, r,
		engine.ExpandTemplateKey(mapsByFormat[format], hm.engine.CN.NegotiateLanguage(w, r)),
		pageContent, breadcrumbs, nil)
}

func (hm *htmlMaps) toViewerPage() (*mapsPage, []engine.Breadcrumb) {
	return &mapsPage{}, nil
}
