package search

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PDOK/gomagpie/config"
	"github.com/PDOK/gomagpie/internal/engine"
	ds "github.com/PDOK/gomagpie/internal/search/datasources"
	"github.com/PDOK/gomagpie/internal/search/datasources/postgres"
	"github.com/PDOK/gomagpie/internal/search/domain"
)

const (
	timeout = time.Second * 15
)

type Search struct {
	engine         *engine.Engine
	datasource     ds.Datasource
	queryExpansion *QueryExpansion
	json           *jsonFeatures
}

func NewSearch(e *engine.Engine, dbConn string, searchIndex string, rewritesFile string, synonymsFile string) (*Search, error) {
	queryExpansion, err := NewQueryExpansion(rewritesFile, synonymsFile)
	if err != nil {
		return nil, err
	}
	s := &Search{
		engine:         e,
		datasource:     newDatasource(e, dbConn, searchIndex),
		json:           newJSONFeatures(e),
		queryExpansion: queryExpansion,
	}
	e.Router.Get("/search", s.Search())
	return s, nil
}

// Search autosuggest locations based on user input
func (s *Search) Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.engine.OpenAPI.ValidateRequest(r); err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}
		collections, searchTerms, outputSRID, limit, err := parseQueryParams(r.URL.Query())
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}

		fc, err := s.datasource.SearchFeaturesAcrossCollections(r.Context(),
			s.queryExpansion.Expand(searchTerms), collections, outputSRID, limit)
		if err != nil {
			handleQueryError(w, err)
			return
		}
		if err = s.enrichFeaturesWithHref(fc); err != nil {
			engine.RenderProblem(engine.ProblemServerError, w, err.Error())
			return
		}

		format := s.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatGeoJSON, engine.FormatJSON:
			s.json.featuresAsGeoJSON(w, r, *s.engine.Config.BaseURL.URL, fc)
		default:
			engine.RenderProblem(engine.ProblemNotAcceptable, w, fmt.Sprintf("format '%s' is not supported", format))
			return
		}
	}
}

func (s *Search) enrichFeaturesWithHref(fc *domain.FeatureCollection) error {
	for _, feat := range fc.Features {
		collectionID, ok := feat.Properties[domain.PropCollectionID]
		if !ok || collectionID == "" {
			return fmt.Errorf("collection reference not found in feature %s", feat.ID)
		}
		collection := config.CollectionByID(s.engine.Config, collectionID.(string))
		if collection.Search != nil {
			for _, ogcColl := range collection.Search.OGCCollections {
				geomType, ok := feat.Properties[domain.PropGeomType]
				if !ok || geomType == "" {
					return fmt.Errorf("geometry type not found in feature %s", feat.ID)
				}
				if strings.EqualFold(ogcColl.GeometryType, geomType.(string)) {
					href, err := url.JoinPath(ogcColl.APIBaseURL.String(), "collections", ogcColl.CollectionID, "items", feat.ID)
					if err != nil {
						return fmt.Errorf("failed to construct API url %w", err)
					}
					href += "?f=json"

					// add href to feature both in GeoJSON properties (for broad compatibility and in line with OGC API Features part 5) and as a Link.
					feat.Properties[domain.PropHref] = href
					feat.Links = []domain.Link{
						{
							Rel:   "canonical",
							Title: "The actual feature in the corresponding OGC API",
							Type:  "application/geo+json",
							Href:  href,
						},
					}
				}
			}
		}
	}
	return nil
}

func newDatasource(e *engine.Engine, dbConn string, searchIndex string) ds.Datasource {
	datasource, err := postgres.NewPostgres(dbConn, timeout, searchIndex)
	if err != nil {
		log.Fatalf("failed to create datasource: %v", err)
	}
	e.RegisterShutdownHook(datasource.Close)
	return datasource
}
