package engine

import (
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

// Resources endpoint to serve static assets, either from local storage or through reverse proxy.
func newResourcesEndpoint(e *Engine) {
	res := e.Config.Resources
	if res == nil {
		return
	}
	var resourcesHandler http.Handler
	if res.Directory != nil && *res.Directory != "" {
		resourcesPath := *res.Directory
		resourcesHandler = http.StripPrefix("/resources", http.FileServer(http.Dir(resourcesPath)))
	} else if res.URL != nil && res.URL.String() != "" {
		resourcesHandler = proxy(e.ReverseProxy, res.URL.String())
	}
	e.Router.Handle("/resources/*", resourcesHandler)
}

type revProxy func(w http.ResponseWriter, r *http.Request, target *url.URL, prefer204 bool, overwrite string)

func proxy(reverseProxy revProxy, resourcesURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resourcePath, _ := url.JoinPath("/", chi.URLParam(r, "*"))
		target, err := url.ParseRequestURI(resourcesURL + resourcePath)
		if err != nil {
			log.Printf("invalid target url, can't proxy resources: %v", err)
			RenderProblem(ProblemServerError, w)

			return
		}
		reverseProxy(w, r, target, false, "")
	}
}
