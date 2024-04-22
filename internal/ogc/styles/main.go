package styles

import (
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"

	"github.com/go-chi/chi/v5"
)

const (
	templatesDir        = "internal/ogc/styles/templates/"
	stylesPath          = "/styles"
	stylesCrumb         = "styles/"
	projectionDelimiter = "__"
)

var (
	defaultProjection = ""
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

	projections := map[string]string{"EPSG:28992": "NetherlandsRDNewQuad", "EPSG:3035": "EuropeanETRS89_LAEAQuad", "EPSG:3857": "WebMercatorQuad"}
	defaultProjection = strings.ToLower(projections[e.Config.OgcAPI.Tiles.SupportedSrs[0].Srs])

	for _, style := range e.Config.OgcAPI.Styles.SupportedStyles {
		for _, supportedSrs := range e.Config.OgcAPI.Tiles.SupportedSrs {
			projection := projections[supportedSrs.Srs]
			zoomLevelRange := supportedSrs.ZoomLevelRange
			styleInstanceID := style.ID + projectionDelimiter + strings.ToLower(projection)
			// Render metadata templates
			e.RenderTemplatesWithParams(struct {
				Metadata   config.Style
				Projection string
			}{Metadata: style, Projection: projection},
				nil,
				engine.NewTemplateKeyWithName(templatesDir+"styleMetadata.go.json", styleInstanceID))
			styleMetadataBreadcrumbs := stylesBreadcrumbs
			styleMetadataBreadcrumbs = append(styleMetadataBreadcrumbs, []engine.Breadcrumb{
				{
					Name: style.Title + " (" + projection + ")",
					Path: stylesCrumb + styleInstanceID,
				},
				{
					Name: "Metadata",
					Path: stylesCrumb + styleInstanceID + "/metadata",
				},
			}...)
			e.RenderTemplatesWithParams(struct {
				Metadata   config.Style
				Projection string
			}{Metadata: style, Projection: projection},
				styleMetadataBreadcrumbs,
				engine.NewTemplateKeyWithName(templatesDir+"styleMetadata.go.html", styleInstanceID))

			// Add existing style definitions to rendered templates
			for _, styleFormat := range style.Formats {
				formatExtension := e.CN.GetStyleFormatExtension(styleFormat.Format)
				styleKey := engine.TemplateKey{
					Name:         style.ID + formatExtension,
					Directory:    e.Config.OgcAPI.Styles.StylesDir,
					Format:       styleFormat.Format,
					InstanceName: styleInstanceID + "." + styleFormat.Format,
				}
				e.RenderTemplatesWithParams(struct {
					Projection     string
					ZoomLevelRange config.ZoomLevelRange
				}{Projection: projection, ZoomLevelRange: zoomLevelRange}, nil, styleKey)
				styleBreadCrumbs := stylesBreadcrumbs
				styleBreadCrumbs = append(styleBreadCrumbs, []engine.Breadcrumb{
					{
						Name: style.Title + " (" + projection + ")",
						Path: stylesCrumb + styleInstanceID,
					},
				}...)
				e.RenderTemplatesWithParams(style,
					styleBreadCrumbs,
					engine.NewTemplateKeyWithName(templatesDir+"style.go.html", styleInstanceID))
			}
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
		style := chi.URLParam(r, "style")
		styleID := strings.Split(style, projectionDelimiter)[0]
		// Previously, the API did not utilise separate styles per projection; whereas the current implementation
		// advertises all possible combinations of available styles and available projections as separate styles.
		// To ensure that the use of style URLs without projection remains possible for previously published APIs,
		// URLs without an explicit projection are defaulted to the first configured projection.
		if style == styleID {
			style += projectionDelimiter + defaultProjection
		}
		styleFormat := s.engine.CN.NegotiateFormat(r)
		var key engine.TemplateKey
		if styleFormat == engine.FormatHTML {
			key = engine.NewTemplateKeyWithNameAndLanguage(
				templatesDir+"style.go.html", style, s.engine.CN.NegotiateLanguage(w, r))
		} else {
			var instanceName string
			if slices.Contains(s.engine.CN.GetSupportedStyleFormats(), styleFormat) {
				instanceName = style + "." + styleFormat
			} else {
				styleFormat = engine.FormatMapboxStyle
				instanceName = style + "." + engine.FormatMapboxStyle
			}
			key = engine.TemplateKey{
				Name:         styleID + s.engine.CN.GetStyleFormatExtension(styleFormat),
				Directory:    s.engine.Config.OgcAPI.Styles.StylesDir,
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
		style := chi.URLParam(r, "style")
		styleID := strings.Split(style, projectionDelimiter)[0]
		// Previously, the API did not utilise separate styles per projection; whereas the current implementation
		// advertises all possible combinations of available styles and available projections as separate styles.
		// To ensure that the use of style URLs without projection remains possible for previously published APIs,
		// URLs without an explicit projection are defaulted to the first configured projection.
		if style == styleID {
			style += projectionDelimiter + defaultProjection
		}
		key := engine.NewTemplateKeyWithNameAndLanguage(
			templatesDir+"styleMetadata.go."+s.engine.CN.NegotiateFormat(r), style, s.engine.CN.NegotiateLanguage(w, r))
		s.engine.ServePage(w, r, key)
	}
}
