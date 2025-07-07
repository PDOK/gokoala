package engine

import (
	htmltemplate "html/template"
	"net/http"
	"path/filepath"
)

var (
	styleTemplate = "theme.go.css"
)

func newCSSEndpoint(e *Engine) {
	templatePath := filepath.Join(templatesDir, styleTemplate)
	template := htmltemplate.Must(
		htmltemplate.New(styleTemplate).ParseFiles(templatePath),
	)

	data := &TemplateData{
		Theme: e.Templates.Theme,
	}

	e.Router.Get("/css/theme.css", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/css")

		if err := template.Execute(w, data); err != nil {
			http.Error(w, "Failed to render theme CSS", http.StatusInternalServerError)
		}
	})

	
}
