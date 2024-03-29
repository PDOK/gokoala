package engine

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
)

func newResourcesEndpoint(e *Engine) {
	// Serve static assets either from local storage or through reverse proxy
	resourcesDir := ""
	if e.Config.Resources.Directory != nil {
		resourcesDir = *e.Config.Resources.Directory
	}
	resourcesURL := ""
	if e.Config.Resources.URL != nil {
		resourcesURL = e.Config.Resources.URL.String()
	}

	if resourcesDir != "" {
		resourcesPath := strings.TrimSuffix(resourcesDir, "/resources")
		e.Router.Handle("/resources/*", http.FileServer(http.Dir(resourcesPath)))
	} else if resourcesURL != "" {
		e.Router.Get("/resources/*",
			func(w http.ResponseWriter, r *http.Request) {
				resourcePath, _ := url.JoinPath("/", chi.URLParam(r, "*"))
				target, err := url.Parse(resourcesURL + resourcePath)
				if err != nil {
					log.Printf("invalid target url, can't proxy resources: %v", err)
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				e.ReverseProxy(w, r, target, true, "")
			})
	}
}
