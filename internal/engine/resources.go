package engine

import (
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Resources endpoint to serve static assets, either from local storage or through reverse proxy.
func newResourcesEndpoint(e *Engine) {
	res := e.Config.Resources
	if res == nil {
		return
	}

	assets := gatherAssets(e)

	for asset := range assets {
		assetFilename := extractFilenameFromPath(asset)
		if assetFilename == "" {
			continue
		}
		if isURL(&asset) {
			// Provision the reverse proxy resource
			e.Router.Handle("/resources/"+assetFilename,
				proxy(e.ReverseProxy, strings.TrimSuffix(asset, "/"+assetFilename), assetFilename),
			)
		} else {
			e.Router.Handle("/resources/"+assetFilename, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, asset)
			}))
		}
	}

	var resourcesHandler http.Handler
	if res.Directory != nil && *res.Directory != "" {
		resourcesPath := *res.Directory
		resourcesHandler = http.StripPrefix("/resources", http.FileServer(http.Dir(resourcesPath)))
	} else if res.URL != nil && res.URL.String() != "" {
		resourcesHandler = proxy(e.ReverseProxy, res.URL.String(), "")
	}
	// The wildcard handle is added last since specific routes should take priority over the wildcard route.
	e.Router.Handle("/resources/*", resourcesHandler)
}

// Get the assets, for now, they are thumbnails that can be a filename (with or without path) or a URL.
func gatherAssets(e *Engine) map[string]struct{} {
	cfg := e.Config
	assets := make(map[string]struct{})

	var resourcesDir string
	if cfg.Resources != nil && cfg.Resources.Directory != nil {
		resourcesDir = strings.TrimPrefix(*cfg.Resources.Directory, ".")
	}

	registerAsset(assets, cfg.Thumbnail, resourcesDir)

	for _, coll := range cfg.AllCollections() {
		if metadata := coll.GetMetadata(); metadata != nil {
			registerAsset(assets, metadata.Thumbnail, resourcesDir)
		}
	}

	if cfg.OgcAPI.Styles != nil {
		for _, style := range cfg.OgcAPI.Styles.SupportedStyles {
			registerAsset(assets, style.Thumbnail, resourcesDir)
		}
	}

	return assets
}

func registerAsset(assets map[string]struct{}, thumbnail *string, resourcesDir string) {
	if thumbnail == nil {
		return
	}

	filename := extractFilenameFromPath(*thumbnail)
	if filename == "" {
		return
	}

	if !isURL(thumbnail) && resourcesDir != "" && strings.Contains(*thumbnail, resourcesDir) {
		// File already lives inside the directory served by the /resources/* wildcard.
		*thumbnail = filename
		return
	}

	assets[*thumbnail] = struct{}{}
	*thumbnail = filename
}

type revProxy func(w http.ResponseWriter, r *http.Request, target *url.URL, prefer204 bool, overwrite string)

func proxy(reverseProxy revProxy, resourcesURL string, resourceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resourcePath string
		if resourceName != "" {
			resourcePath = "/" + resourceName
		} else {
			resourcePath, _ = url.JoinPath("/", chi.URLParam(r, "*"))
		}
		target, err := url.ParseRequestURI(resourcesURL + resourcePath)
		if err != nil {
			log.Printf("invalid target url, can't proxy resources: %v", err)
			RenderProblem(ProblemServerError, w)

			return
		}
		reverseProxy(w, r, target, false, "")
	}
}

func extractFilenameFromPath(urlPath string) string {
	u, err := url.Parse(urlPath)
	if err != nil {
		return ""
	}
	ext := path.Ext(u.Path)
	if ext == "" {
		return ""
	}
	// Extension found (above), so we can assume there is a filename in the path.
	urlParts := strings.Split(urlPath, "/")
	if len(urlParts) <= 1 {
		return ""
	}
	return urlParts[len(urlParts)-1]
}

func isURL(thumbnail *string) bool {
	return strings.HasPrefix(*thumbnail, "http://") ||
		strings.HasPrefix(*thumbnail, "https://")
}
