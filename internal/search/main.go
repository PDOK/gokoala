package search

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PDOK/gomagpie/internal/engine"
)

type Search struct {
	engine *engine.Engine
}

func NewSearch(e *engine.Engine) *Search {
	s := &Search{
		engine: e,
	}
	e.Router.Get("/search/suggest", s.Suggest())
	return s
}

// Suggest autosuggest locations based on user input
func (s *Search) Suggest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := parseQueryParams(r.URL.Query())
		if err != nil {
			log.Printf("%v", err)
			engine.RenderProblem(engine.ProblemBadRequest, w)
			return
		}
		searchQuery := params["q"]
		delete(params, "q")
		format := params["f"]
		delete(params, "f")
		crs := params["crs"]
		delete(params, "crs")

		log.Printf("crs %s, format %s, query %s, params %v", crs, format, searchQuery, params)
	}
}

func parseQueryParams(query url.Values) (map[string]any, error) {
	result := make(map[string]any, len(query))

	deepObjectParams := make(map[string]map[string]string)
	for key, values := range query {
		if strings.Contains(key, "[") {
			// Extract deepObject parameters
			parts := strings.SplitN(key, "[", 2)
			mainKey := parts[0]
			subKey := strings.TrimSuffix(parts[1], "]")

			if _, exists := deepObjectParams[mainKey]; !exists {
				deepObjectParams[mainKey] = make(map[string]string)
			}
			deepObjectParams[mainKey][subKey] = values[0]
		} else {
			// Extract regular (flat) parameters
			result[key] = values[0]
		}
	}
	for mainKey, subParams := range deepObjectParams {
		result[mainKey] = subParams
	}
	return result, nil
}
