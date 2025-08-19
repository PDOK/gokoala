package engine

import "net/http"

func newSitemap(e *Engine) {
	for path, template := range map[string]string{"/sitemap.xml": "sitemap.go.xml", "/robots.txt": "robots.go.txt"} {
		key := NewTemplateKey(templatesDir + template)
		e.renderTemplates(path, nil, nil, false, key)
		e.Router.Get(path, func(w http.ResponseWriter, r *http.Request) {
			e.serve(w, r, &key, false, false, "", nil)
		})
	}
}
