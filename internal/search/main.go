package search

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PDOK/gomagpie/internal/engine"
	ds "github.com/PDOK/gomagpie/internal/search/datasources"
	"github.com/PDOK/gomagpie/internal/search/datasources/postgres"
)

const timeout = time.Second * 15

type Search struct {
	engine     *engine.Engine
	datasource ds.Datasource
}

func NewSearch(e *engine.Engine, dbConn string, searchIndex string) *Search {
	s := &Search{
		engine:     e,
		datasource: newDatasource(e, dbConn, searchIndex),
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
		limit := params["limit"]
		delete(params, "limit")

		log.Printf("crs %s, limit %d, format %s, query %s, params %v", crs, limit, format, searchQuery, params)

		suggestions, err := s.datasource.Suggest(r.Context(), r.URL.Query().Get("q"))
		if err != nil {
			engine.RenderProblem(engine.ProblemServerError, w, err.Error())
			return
		}
		serveJSON(suggestions, engine.MediaTypeGeoJSON, w)
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

func newDatasource(e *engine.Engine, dbConn string, searchIndex string) ds.Datasource {
	datasource, err := postgres.NewPostgres(dbConn, timeout, searchIndex)
	if err != nil {
		log.Fatalf("failed to create datasource: %v", err)
	}
	e.RegisterShutdownHook(datasource.Close)
	return datasource
}
