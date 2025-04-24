package styles

import (
	"log"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/engine/util"
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

	stylesBreadcrumbs = []engine.Breadcrumb{
		{
			Name: "Styles",
			Path: "styles",
		},
	}
)

type stylesTemplateData struct {
	// Projection used by default
	DefaultProjection string

	// All supported projections for this dataset
	SupportedProjections []config.SupportedSrs

	// All supported projections by GoKoala (for tiles)
	AllProjections map[string]any
}

type stylesMetadataTemplateData struct {
	// Metadata about this style
	Metadata config.Style

	// Projection used by this style
	Projection string
}

type Styles struct {
	engine                *engine.Engine
	localResourcesHandler http.Handler
}

func NewStyles(e *engine.Engine) *Styles {
	// default style must be the first entry in supported styles
	if e.Config.OgcAPI.Styles.Default != e.Config.OgcAPI.Styles.SupportedStyles[0].ID {
		log.Fatalf("default style must be first entry in supported styles. '%s' does not match '%s'",
			e.Config.OgcAPI.Styles.SupportedStyles[0].ID, e.Config.OgcAPI.Styles.Default)
	}

	allProjections := util.Cast(config.AllTileProjections)
	supportedProjections := e.Config.OgcAPI.Tiles.GetProjections()
	if len(supportedProjections) == 0 {
		log.Fatalf("failed to setup OGC API Styles, no supported projections (SRS) found in OGC API Tiles")
	}
	defaultProjection = strings.ToLower(config.AllTileProjections[supportedProjections[0].Srs])

	e.RenderTemplatesWithParams(stylesPath,
		&stylesTemplateData{defaultProjection, supportedProjections, allProjections},
		stylesBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"styles.go.json"),
		engine.NewTemplateKey(templatesDir+"styles.go.html"))

	renderStylesPerProjection(e, supportedProjections)

	styles := &Styles{
		engine: e,
	}
	e.Router.Get(stylesPath, styles.Styles())
	e.Router.Get(stylesPath+"/{style}", styles.Style())
	e.Router.Get(stylesPath+"/{style}/metadata", styles.Metadata())
	e.Router.Get(stylesPath+"/{style}/legend", styles.Legend())

	if res := e.Config.Resources; e.Config.Resources != nil {
		if res != nil && res.Directory != nil && *res.Directory != "" {
			resourcesPath := *res.Directory
			styles.localResourcesHandler = http.FileServer(http.Dir(resourcesPath))
		}
	}
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
		style, styleID := parseStyleParam(r)
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

func (s *Styles) Metadata() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		style, _ := parseStyleParam(r)
		key := engine.NewTemplateKeyWithNameAndLanguage(
			templatesDir+"styleMetadata.go."+s.engine.CN.NegotiateFormat(r), style, s.engine.CN.NegotiateLanguage(w, r))
		s.engine.ServePage(w, r, key)
	}
}

func (s *Styles) Legend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, styleID := parseStyleParam(r)

		// search matching legend when configured
		var legend string
		for _, supportedStyle := range s.engine.Config.OgcAPI.Styles.SupportedStyles {
			if supportedStyle.ID == styleID && supportedStyle.Legend != nil {
				legend = *supportedStyle.Legend
				break
			}
		}
		if s.engine.Config.Resources == nil {
			engine.RenderProblem(engine.ProblemNotFound, w, "no legends configured")
			return
		}
		if legend == "" {
			engine.RenderProblem(engine.ProblemNotFound, w, "no legend configured for style "+styleID)
			return
		}

		// rewrite legend url to configured legend resource (png file).
		switch {
		case s.localResourcesHandler != nil:
			s.localResourcesHandler.ServeHTTP(w, r)
		case s.engine.Config.Resources.URL != nil:
			legendURL, _ := url.JoinPath(s.engine.Config.Resources.URL.String(), legend)
			target, _ := url.ParseRequestURI(legendURL)
			s.engine.ReverseProxy(w, r, target, false, "")
		default:
			engine.RenderProblem(engine.ProblemServerError, w, "legend not properly configured")
		}
	}
}

func parseStyleParam(r *http.Request) (style string, styleID string) {
	style = chi.URLParam(r, "style")
	styleID = strings.Split(style, projectionDelimiter)[0]

	// Previously, the API did not utilise separate styles per projection; whereas the current implementation
	// advertises all possible combinations of available styles and available projections as separate styles.
	// To ensure that the use of style URLs without projection remains possible for previously published APIs,
	// URLs without an explicit projection are defaulted to the first configured projection.
	if style == styleID {
		style += projectionDelimiter + defaultProjection
	}
	return style, styleID
}

func renderStylesPerProjection(e *engine.Engine, supportedProjections []config.SupportedSrs) {
	for _, style := range e.Config.OgcAPI.Styles.SupportedStyles {
		for _, supportedSrs := range supportedProjections {
			projection := config.AllTileProjections[supportedSrs.Srs]
			zoomLevelRange := supportedSrs.ZoomLevelRange
			styleInstanceID := style.ID + projectionDelimiter + strings.ToLower(projection)
			styleProjectionBreadcrumb := engine.Breadcrumb{
				Name: style.Title + " (" + projection + ")",
				Path: stylesCrumb + styleInstanceID,
			}
			data := &stylesMetadataTemplateData{style, projection}

			// Render metadata template (JSON)
			path := stylesPath + "/" + styleInstanceID + "/metadata"
			e.RenderTemplatesWithParams(path, data, nil,
				engine.NewTemplateKeyWithName(templatesDir+"styleMetadata.go.json", styleInstanceID))

			// Render metadata template (HTML)
			styleMetadataBreadcrumbs := stylesBreadcrumbs
			styleMetadataBreadcrumbs = append(styleMetadataBreadcrumbs, []engine.Breadcrumb{
				styleProjectionBreadcrumb,
				{
					Name: "Metadata",
					Path: stylesCrumb + styleInstanceID + "/metadata",
				},
			}...)
			e.RenderTemplatesWithParams(path, data, styleMetadataBreadcrumbs,
				engine.NewTemplateKeyWithName(templatesDir+"styleMetadata.go.html", styleInstanceID))

			// Add existing style definitions to rendered templates
			renderStylePerFormat(e, style, styleInstanceID, projection, zoomLevelRange, styleProjectionBreadcrumb)
		}
	}
}

func renderStylePerFormat(e *engine.Engine, style config.Style, styleInstanceID string,
	projection string, zoomLevelRange config.ZoomLevelRange, styleProjectionBreadcrumb engine.Breadcrumb) {

	for _, styleFormat := range style.Formats {
		formatExtension := e.CN.GetStyleFormatExtension(styleFormat.Format)
		styleKey := engine.TemplateKey{
			Name:         style.ID + formatExtension,
			Directory:    e.Config.OgcAPI.Styles.StylesDir,
			Format:       styleFormat.Format,
			InstanceName: styleInstanceID + "." + styleFormat.Format,
		}
		path := stylesPath + "/" + styleInstanceID

		// Render template (JSON)
		e.RenderTemplatesWithParams(path, struct {
			Projection     string
			ZoomLevelRange config.ZoomLevelRange
		}{Projection: projection, ZoomLevelRange: zoomLevelRange}, nil, styleKey)

		// Render template (HTML)
		styleBreadCrumbs := stylesBreadcrumbs
		styleBreadCrumbs = append(styleBreadCrumbs, styleProjectionBreadcrumb)
		e.RenderTemplatesWithParams(path, style, styleBreadCrumbs,
			engine.NewTemplateKeyWithName(templatesDir+"style.go.html", styleInstanceID))
	}
}
