package engine

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Serve static assets either from local storage or through reverse proxy
func newResourcesEndpoint(e *Engine) {
	if e.Config.Resources.Directory != nil && *e.Config.Resources.Directory != "" {
		resourcesPath := strings.TrimSuffix(*e.Config.Resources.Directory, "/resources")
		e.Router.Handle("/resources/*", http.FileServer(http.Dir(resourcesPath)))
	} else if e.Config.Resources.URL != nil && e.Config.Resources.URL.String() != "" {
		e.Router.Get("/resources/*", proxy(e.ReverseProxy, e.Config.Resources.URL.String()))
	}
}

type reverseProxy func(w http.ResponseWriter, r *http.Request, target *url.URL, prefer204 bool, overwrite string)

func proxy(rp reverseProxy, resourcesURL string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resourcePath, _ := url.JoinPath("/", chi.URLParam(r, "*"))
		target, err := url.ParseRequestURI(resourcesURL + resourcePath)
		if err != nil {
			log.Printf("invalid target url, can't proxy resources: %v", err)
			RenderProblem(ProblemServerError, w)
			return
		}
		rp(w, r, target, true, "")
	}
}
