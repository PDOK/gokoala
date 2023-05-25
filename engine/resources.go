package engine

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Resources struct {
	engine *Engine
}

func NewResources(e *Engine, router *chi.Mux) *Resources {
	resources := &Resources{
		engine: e,
	}

	// Serve static assets either from local storage or through reverse proxy
	if resourcesDir := e.Config.ResourcesServer.Directory; resourcesDir != "" {
		resourcesPath := strings.TrimSuffix(resourcesDir, "/resources")
		router.Handle("/resources/*", http.FileServer(http.Dir(resourcesPath)))
	} else if resourcesURL := e.Config.ResourcesServer.URL.String(); resourcesURL != "" {
		router.Get("/resources/*",
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

	return resources
}
