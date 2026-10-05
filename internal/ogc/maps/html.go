package maps

import "github.com/PDOK/gokoala/internal/engine"

var (
	mapsHTMLKey = engine.NewTemplateKey(templatesDir + "maps.go.html")
	mapsMDKey   = engine.NewTemplateKey(templatesDir + "maps.go.md")
)

type htmlMaps struct {
	engine *engine.Engine
}

// newHTMLMaps render the requested as HTML or Markdown
func newHTMLMaps(e *engine.Engine) *htmlMaps {
	e.ParseTemplate(mapsHTMLKey)

	e.ParseTemplate(mapsMDKey)

	return &htmlMaps{
		engine: e,
	}
}
