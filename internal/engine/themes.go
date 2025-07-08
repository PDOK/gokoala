package engine

import (
	htmltemplate "html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/PDOK/gokoala/config"
)

const (
	styleTemplate = "theme.go.css"
)

func initializeTheme(theme *config.Theme, e *Engine) {
	newCSSEndpoint(e)
	newThemeEndpointsAndCreateLogosWithServerPaths(e, theme)
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
		w.Header().Set("Content-Type", "text/css")

		if err := template.Execute(w, data); err != nil {
			log.Fatal("Failed to render theme CSS")
		}
	})
}

func newThemeEndpointsAndCreateLogosWithServerPaths(e *Engine, theme *config.Theme) {
	// Replace the theme Logo properties with the absolute paths for the template
	theme.Logo = &config.ThemeLogo{
		Header:    newStaticEndppoint(e, theme.Logo.Header),
		Footer:    newStaticEndppoint(e, theme.Logo.Footer),
		Opengraph: newStaticEndppoint(e, theme.Logo.Header),
	}
}

func newStaticEndppoint(e *Engine, file string) string {
	// Get the clean dir from config (remove any "./" prefixes if added)
	dir := filepath.Dir(file)

	// Prefix so http#StripPrefix knows what to remove from URL
	prefix := "/" + dir

	// Actual route for chi
	route := prefix + "/*"

	// Serve the route
	fs := http.StripPrefix(prefix, http.FileServer(http.Dir(dir)))
	e.Router.Get(route, func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})

	// Return the absolute path
	return "/" + filepath.Clean(file)
}
