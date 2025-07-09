package engine

import (
	htmltemplate "html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/PDOK/gokoala/config"
)

const (
	styleTemplate = "theme.go.css"
)

func newThemeEndpoints(theme *config.Theme, e *Engine) {
	newCSSEndpoint(e)

	// Replace the theme Logo properties with the absolute paths for the template
	theme.Logo = &config.ThemeLogo{
		Header:    newThemeAssetEndpoint(e, theme.Path, theme.Logo.Header),
		Footer:    newThemeAssetEndpoint(e, theme.Path, theme.Logo.Footer),
		Opengraph: newThemeAssetEndpoint(e, theme.Path, theme.Logo.Opengraph),
		Favicon:   newThemeAssetEndpoint(e, theme.Path, theme.Logo.Favicon),
		Favicon16: newThemeAssetEndpoint(e, theme.Path, theme.Logo.Favicon16),
		Favicon32: newThemeAssetEndpoint(e, theme.Path, theme.Logo.Favicon32),
	}
}

func newCSSEndpoint(e *Engine) {
	templatePath := filepath.Join(templatesDir, styleTemplate)
	template := htmltemplate.Must(
		htmltemplate.New(styleTemplate).ParseFiles(templatePath),
	)

	data := &TemplateData{
		Theme: e.Templates.Theme,
	}

	// Parse CSS with variables from the config file
	e.Router.Get("/css/theme.css", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(HeaderContentType, "text/css")

		if err := template.Execute(w, data); err != nil {
			log.Fatal("Failed to render theme CSS")
		}
	})
}

func newThemeAssetEndpoint(e *Engine, themePath string, file string) string {
	// Get the (full) clean dir from config, relative to theme.yaml (remove any "./" prefixes if added)
	cleanPath := filepath.Dir(file)

	dir := filepath.Join(themePath, cleanPath)

	// Prefix so http#StripPrefix knows what to remove from URL
	prefix := "/theme/" + cleanPath

	// Actual route for chi
	route := prefix + "/*"

	// Serve the route
	fs := http.StripPrefix(prefix, http.FileServer(http.Dir(dir)))
	e.Router.Get(route, func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})

	var absolutePath string
	if !strings.HasPrefix(file, "/") {
		absolutePath = "/"
	}

	// Return the new (absolute) path
	return absolutePath + filepath.Clean(file)
}
