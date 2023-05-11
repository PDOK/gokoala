package styles

import (
	"log"
	"net/http"

	"github.com/PDOK/gokoala/engine"

	"github.com/go-chi/chi/v5"
)

const (
	templatesDir = "ogc/styles/templates/"
	stylesPath   = "/styles"
)

type Styles struct {
	engine *engine.Engine
}

func NewStyles(e *engine.Engine, router *chi.Mux) *Styles {
	if e.Config.ResourcesDir == "" {
		// TODO: Should become optional once we support CDNs for resources
		log.Fatalf("resources-dir is required when using OGC styles")
	}

	stylesBreadcrumbs := []engine.Breadcrumb{
		engine.Breadcrumb{
			Name: "Styles",
			Path: "styles",
		},
	}

	e.RenderTemplates(stylesPath,
		stylesBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"styles.go.json"),
		engine.NewTemplateKey(templatesDir+"styles.go.html"))

	for _, style := range e.Config.OgcAPI.Styles.SupportedStyles {
		// Render metadata templates
		e.RenderTemplatesWithParams(style, nil, engine.NewTemplateKeyWithName(templatesDir+"styleMetadata.go.json", style.ID))
		styleMetadataBreadcrumbs := append(stylesBreadcrumbs, []engine.Breadcrumb{
			engine.Breadcrumb{
				Name: style.ID,
				Path: "styles/" + style.ID,
			},
			engine.Breadcrumb{
				Name: "Metadata",
				Path: "styles/" + style.ID + "/metadata",
			},
		}...)
		e.RenderTemplatesWithParams(style, styleMetadataBreadcrumbs, engine.NewTemplateKeyWithName(templatesDir+"styleMetadata.go.html", style.ID))

		// Add existing style definitions to rendered templates
		for _, stylesheet := range style.Stylesheets {
			formatExtension := e.CN.GetStyleFormatExtension(*stylesheet.Link.Format)
			styleKey := engine.TemplateKey{
				Name:         style.ID + formatExtension,
				Directory:    e.Config.ResourcesDir,
				Format:       *stylesheet.Link.Format,
				InstanceName: style.ID + "." + *stylesheet.Link.Format,
			}
			e.RenderTemplatesWithParams(nil, nil, styleKey)
		}
	}

	styles := &Styles{
		engine: e,
	}

	router.Get(stylesPath, styles.Styles())
	router.Get(stylesPath+"/{style}", styles.Style())
	router.Get(stylesPath+"/{style}/metadata", styles.StyleMetadata())

	return styles
}

func (s *Styles) Styles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKey(templatesDir + "styles.go." + s.engine.CN.NegotiateFormat(r))
		s.engine.ServePage(w, r, key)
	}
}

func (s *Styles) Style() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		styleID := chi.URLParam(r, "style")
		styleFormat := s.engine.CN.NegotiateFormat(r)
		var instanceName string
		if engine.Contains(s.engine.CN.GetSupportedStyleFormats(), styleFormat) {
			instanceName = styleID + "." + styleFormat
		} else {
			styleFormat = "mapbox"
			instanceName = styleID + ".mapbox"
		}
		key := engine.TemplateKey{
			Name:         styleID + s.engine.CN.GetStyleFormatExtension(styleFormat),
			Directory:    s.engine.Config.ResourcesDir,
			Format:       styleFormat,
			InstanceName: instanceName,
		}
		s.engine.ServePage(w, r, key)
	}
}

func (s *Styles) StyleMetadata() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		styleID := chi.URLParam(r, "style")
		key := engine.NewTemplateKeyWithName(templatesDir+"styleMetadata.go."+s.engine.CN.NegotiateFormat(r), styleID)
		s.engine.ServePage(w, r, key)
	}
}
