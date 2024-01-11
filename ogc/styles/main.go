package styles

import (
	"log"
	"net/http"
	"slices"

	"github.com/PDOK/gokoala/engine"

	"github.com/go-chi/chi/v5"
)

const (
	templatesDir = "ogc/styles/templates/"
	stylesPath   = "/styles"
	stylesCrumb  = "styles/"
)

type Styles struct {
	engine *engine.Engine
}

func NewStyles(e *engine.Engine) *Styles {
	// default style must be the first entry in supportedstyles
	if e.Config.OgcAPI.Styles.Default != e.Config.OgcAPI.Styles.SupportedStyles[0].ID {
		log.Fatalf("default style must be first entry in supported styles. '%s' does not match '%s'",
			e.Config.OgcAPI.Styles.SupportedStyles[0].ID, e.Config.OgcAPI.Styles.Default)
	}

	stylesBreadcrumbs := []engine.Breadcrumb{
		{
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
		e.RenderTemplatesWithParams(style,
			nil,
			engine.NewTemplateKeyWithName(templatesDir+"styleMetadata.go.json", style.ID))
		styleMetadataBreadcrumbs := stylesBreadcrumbs
		styleMetadataBreadcrumbs = append(styleMetadataBreadcrumbs, []engine.Breadcrumb{
			{
				Name: style.Title,
				Path: stylesCrumb + style.ID,
			},
			{
				Name: "Metadata",
				Path: stylesCrumb + style.ID + "/metadata",
			},
		}...)
		e.RenderTemplatesWithParams(style,
			styleMetadataBreadcrumbs,
			engine.NewTemplateKeyWithName(templatesDir+"styleMetadata.go.html", style.ID))

		// Add existing style definitions to rendered templates
		for _, stylesheet := range style.Stylesheets {
			formatExtension := e.CN.GetStyleFormatExtension(*stylesheet.Link.Format)
			styleKey := engine.TemplateKey{
				Name:         style.ID + formatExtension,
				Directory:    e.Config.OgcAPI.Styles.MapboxStylesPath,
				Format:       *stylesheet.Link.Format,
				InstanceName: style.ID + "." + *stylesheet.Link.Format,
			}
			e.RenderTemplatesWithParams(nil, nil, styleKey)
			styleBreadCrumbs := stylesBreadcrumbs
			styleBreadCrumbs = append(styleBreadCrumbs, []engine.Breadcrumb{
				{
					Name: style.Title,
					Path: stylesCrumb + style.ID,
				},
			}...)
			e.RenderTemplatesWithParams(style,
				styleBreadCrumbs,
				engine.NewTemplateKeyWithName(templatesDir+"style.go.html", style.ID))
		}
	}

	styles := &Styles{
		engine: e,
	}

	e.Router.Get(stylesPath, styles.Styles())
	e.Router.Get(stylesPath+"/{style}", styles.Style())
	e.Router.Get(stylesPath+"/{style}/metadata", styles.StyleMetadata())

	return styles
}

func (s *Styles) Styles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKeyWithLanguage(
			templatesDir+"styles.go."+s.engine.CN.NegotiateFormat(r), s.engine.CN.NegotiateLanguage(w, r))
		s.engine.ServePage(w, r, key)
	}
}

func (s *Styles) Style() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		styleID := chi.URLParam(r, "style")
		styleFormat := s.engine.CN.NegotiateFormat(r)
		// TODO: improve?
		var key engine.TemplateKey
		if styleFormat == engine.FormatHTML {
			key = engine.NewTemplateKeyWithNameAndLanguage(
				templatesDir+"style.go.html", styleID, s.engine.CN.NegotiateLanguage(w, r))
		} else {
			var instanceName string
			if slices.Contains(s.engine.CN.GetSupportedStyleFormats(), styleFormat) {
				instanceName = styleID + "." + styleFormat
			} else {
				styleFormat = "mapbox"
				instanceName = styleID + ".mapbox"
			}
			key = engine.TemplateKey{
				Name:         styleID + s.engine.CN.GetStyleFormatExtension(styleFormat),
				Directory:    s.engine.Config.OgcAPI.Styles.MapboxStylesPath,
				Format:       styleFormat,
				InstanceName: instanceName,
				Language:     s.engine.CN.NegotiateLanguage(w, r),
			}
		}
		s.engine.ServePage(w, r, key)
	}
}

func (s *Styles) StyleMetadata() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		styleID := chi.URLParam(r, "style")
		key := engine.NewTemplateKeyWithNameAndLanguage(
			templatesDir+"styleMetadata.go."+s.engine.CN.NegotiateFormat(r), styleID, s.engine.CN.NegotiateLanguage(w, r))
		s.engine.ServePage(w, r, key)
	}
}
