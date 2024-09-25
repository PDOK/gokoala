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
		e.Router.Get("/resources/*", proxy(e.ReverseProxy, resourcesURL))
	}
}

func proxy(reverseProxy func(w http.ResponseWriter, r *http.Request, target *url.URL, prefer204 bool, overwrite string),
	resourcesURL string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		resourcePath, _ := url.JoinPath("/", chi.URLParam(r, "*"))
		target, err := url.ParseRequestURI(resourcesURL + resourcePath)
		if err != nil {
			log.Printf("invalid target url, can't proxy resources: %v", err)
			RenderProblem(ProblemServerError, w)
			return
		}
		reverseProxy(w, r, target, true, "")
	}
}
