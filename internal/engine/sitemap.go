package engine

import "net/http"

// newSitemap renders sitemap.xml but also robots.txt and llms.txt (which is like a sitemap for agentic browsing).
func newSitemap(e *Engine) {
	discoveryTemplates := map[string]string{
		"/sitemap.xml": "sitemap.go.xml",
		"/robots.txt":  "robots.go.txt",
		"/llms.txt":    "llms.go.txt", // follows https://llmstxt.org/ v2
	}

	for path, template := range discoveryTemplates {
		key := NewTemplateKey(templatesDir + template)
		e.renderTemplates(path, nil, nil, false, key)
		e.Router.Get(path, func(w http.ResponseWriter, r *http.Request) {
			e.Serve(w, r, ServeTemplate(key), ServeValidation(false, false))
		})
	}
}
