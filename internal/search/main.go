package search

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	featdomain "github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/PDOK/gokoala/internal/search/domain"
)

type Search struct {
	engine         *engine.Engine
	datasource     ds.Datasource
	queryExpansion *QueryExpansion
	json           *jsonSearchResults
}

func NewSearch(e *engine.Engine, datasources map[features.DatasourceKey]ds.Datasource,
	_ map[int]featdomain.AxisOrder, rewritesFile, synonymsFile string) (*Search, error) {

	queryExpansion, err := NewQueryExpansion(rewritesFile, synonymsFile)
	if err != nil {
		return nil, err
	}

	// TODO come up with something smarter
	var firstDS ds.Datasource
	for _, v := range datasources {
		firstDS = v
		break
	}

	s := &Search{
		engine:         e,
		datasource:     firstDS,
		json:           newJSONSearchResults(e),
		queryExpansion: queryExpansion,
	}
	e.Router.Get("/search", s.Search())
	return s, nil
}

// Search autosuggest locations based on user input
func (s *Search) Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate
		if err := s.engine.OpenAPI.ValidateRequest(r); err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}
		collections, searchTerms, outputSRID, contentCrs, bbox, bboxSRID, limit, err := parseQueryParams(r.URL.Query())
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}
		w.Header().Add(engine.HeaderContentCrs, contentCrs.ToLink())

		// Query expansion
		searchQuery, err := s.queryExpansion.Expand(r.Context(), searchTerms)
		if err != nil {
			handleQueryError(w, err)
			return
		}

		// Perform actual search
		fc, err := s.datasource.SearchFeaturesAcrossCollections(r.Context(), ds.FeaturesSearchCriteria{
			SearchQuery: *searchQuery,
			Settings:    s.engine.Config.OgcAPI.FeaturesSearch.SearchSettings,
			Limit:       limit,
			InputSRID:   bboxSRID,
			OutputSRID:  outputSRID,
			Bbox:        bbox,
		}, collections)
		if err != nil {
			handleQueryError(w, err)
			return
		}
		if err = s.enrichFeaturesWithHref(fc, contentCrs); err != nil {
			engine.RenderProblem(engine.ProblemServerError, w, err.Error())
			return
		}

		// Output
		format := s.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatGeoJSON, engine.FormatJSON:
			s.json.searchResultsAsGeoJSON(w, r, *s.engine.Config.BaseURL.URL, fc)
		case engine.FormatJSONFG:
			s.json.searchResultsAsJSONFG(w, r, *s.engine.Config.BaseURL.URL, fc, contentCrs)
		default:
			engine.RenderProblem(engine.ProblemNotAcceptable, w, fmt.Sprintf("format '%s' is not supported", format))
			return
		}
	}
}

func (s *Search) enrichFeaturesWithHref(fc *featdomain.FeatureCollection, contentCrs featdomain.ContentCrs) error {
	for _, feat := range fc.Features {
		collectionID := feat.Properties.Value(domain.PropCollectionID)
		if collectionID == "" {
			return fmt.Errorf("collection reference not found in feature %s", feat.ID)
		}
		var collection *config.GeoSpatialCollection
		for _, coll := range s.engine.Config.AllCollections() {
			if collectionID == coll.ID && coll.Features != nil && coll.FeaturesSearch != nil {
				collection = &coll
				break
			}
		}
		if collection == nil {
			continue
		}
		for _, ogcColl := range collection.FeaturesSearch.OGCCollections {
			geomType := feat.Properties.Value(domain.PropGeomType)
			if geomType == "" {
				return fmt.Errorf("geometry type not found in feature %s", feat.ID)
			}
			if strings.EqualFold(ogcColl.GeometryType, geomType.(string)) {
				href, err := url.JoinPath(ogcColl.APIBaseURL.String(), "collections", ogcColl.CollectionID, "items", feat.ID)
				if err != nil {
					return fmt.Errorf("failed to construct API url %w", err)
				}
				href += "?f=json"

				if contentCrs != "" && !contentCrs.IsWGS84() {
					href += fmt.Sprintf("&crs=%s", contentCrs)
				}

				// add href to feature both in GeoJSON properties (for broad compatibility and in line with OGC API Features part 5) and as a Link.
				feat.Properties.Set(domain.PropHref, href)
				feat.Links = []featdomain.Link{
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
	return nil
}

// log error, but send generic message to client to prevent possible information leakage from datasource
func handleQueryError(w http.ResponseWriter, err error) {
	msg := "failed to fulfill search request"
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		// provide more context when user hits the query timeout
		msg += ": querying took too long (timeout encountered). Simplify your request and try again, or contact support"
	}
	log.Printf("%s, error: %v\n", msg, err)
	engine.RenderProblem(engine.ProblemServerError, w, msg) // don't include sensitive information in details msg
}
