package search

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PDOK/gomagpie/config"
	"github.com/PDOK/gomagpie/internal/engine"
	ds "github.com/PDOK/gomagpie/internal/search/datasources"
	"github.com/PDOK/gomagpie/internal/search/datasources/postgres"
	"github.com/PDOK/gomagpie/internal/search/domain"
)

const (
	queryParam = "q"
	limitParam = "limit"
	crsParam   = "crs"

	limitDefault = 10
	limitMax     = 50

	timeout = time.Second * 15
)

var (
	deepObjectParamRegex = regexp.MustCompile(`\w+\[\w+]`)
)

type Search struct {
	engine     *engine.Engine
	datasource ds.Datasource
}

func NewSearch(e *engine.Engine, dbConn string, searchIndex string) *Search {
	s := &Search{
		engine:     e,
		datasource: newDatasource(e, dbConn, searchIndex),
	}
	e.Router.Get("/search", s.Search())
	return s
}

// Search autosuggest locations based on user input
func (s *Search) Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collections, searchTerm, outputSRID, limit, err := parseQueryParams(r.URL.Query())
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}
		fc, err := s.datasource.Search(r.Context(), searchTerm, collections, outputSRID, limit)
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
			featuresAsGeoJSON(w, *s.engine.Config.BaseURL.URL, fc)
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

func parseQueryParams(query url.Values) (collections ds.CollectionsWithParams, searchTerm string, outputSRID domain.SRID, limit int, err error) {
	err = validateNoUnknownParams(query)
	if err != nil {
		return
	}
	searchTerm, searchTermErr := parseSearchTerm(query)
	collections = parseDeepObjectParams(query)
	outputSRID, outputSRIDErr := parseCrsToSRID(query, crsParam)
	limit, limitErr := parseLimit(query)
	err = errors.Join(searchTermErr, limitErr, outputSRIDErr)
	return
}

// Parse "deep object" params, e.g. paramName[prop1]=value1&paramName[prop2]=value2&....
func parseDeepObjectParams(query url.Values) ds.CollectionsWithParams {
	deepObjectParams := make(ds.CollectionsWithParams, len(query))
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
		}
	}
	return deepObjectParams
}

func parseSearchTerm(query url.Values) (searchTerm string, err error) {
	searchTerm = query.Get(queryParam)
	if searchTerm == "" {
		err = fmt.Errorf("no search term provided, '%s' query parameter is required", queryParam)
	}
	return
}

func newDatasource(e *engine.Engine, dbConn string, searchIndex string) ds.Datasource {
	datasource, err := postgres.NewPostgres(dbConn, timeout, searchIndex)
	if err != nil {
		log.Fatalf("failed to create datasource: %v", err)
	}
	e.RegisterShutdownHook(datasource.Close)
	return datasource
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

// implements req 7.6 (https://docs.ogc.org/is/17-069r4/17-069r4.html#query_parameters)
func validateNoUnknownParams(query url.Values) error {
	copyParams := clone(query)
	copyParams.Del(engine.FormatParam)
	copyParams.Del(queryParam)
	copyParams.Del(limitParam)
	copyParams.Del(crsParam)
	for key := range query {
		if deepObjectParamRegex.MatchString(key) {
			copyParams.Del(key)
		}
	}
	if len(copyParams) > 0 {
		return fmt.Errorf("unknown query parameter(s) found: %v", copyParams.Encode())
	}
	return nil
}

func clone(params url.Values) url.Values {
	copyParams := url.Values{}
	for k, v := range params {
		copyParams[k] = v
	}
	return copyParams
}

func parseCrsToSRID(params url.Values, paramName string) (domain.SRID, error) {
	param := params.Get(paramName)
	if param == "" {
		return domain.UndefinedSRID, nil
	}
	param = strings.TrimSpace(param)
	if !strings.HasPrefix(param, domain.CrsURIPrefix) {
		return domain.UndefinedSRID, fmt.Errorf("%s param should start with %s, got: %s", paramName, domain.CrsURIPrefix, param)
	}
	var srid domain.SRID
	lastIndex := strings.LastIndex(param, "/")
	if lastIndex != -1 {
		crsCode := param[lastIndex+1:]
		if crsCode == domain.WGS84CodeOGC {
			return domain.WGS84SRIDPostgis, nil // CRS84 is WGS84, just like EPSG:4326 (only axis order differs but SRID is the same)
		}
		val, err := strconv.Atoi(crsCode)
		if err != nil {
			return 0, fmt.Errorf("expected numerical CRS code, received: %s", crsCode)
		}
		srid = domain.SRID(val)
	}
	return srid, nil
}

func parseLimit(params url.Values) (int, error) {
	limit := limitDefault
	var err error
	if params.Get(limitParam) != "" {
		limit, err = strconv.Atoi(params.Get(limitParam))
		if err != nil {
			err = errors.New("limit must be numeric")
		}
		// "If the value of the limit parameter is larger than the maximum value, this SHALL NOT result
		//  in an error (instead use the maximum as the parameter value)."
		if limit > limitMax {
			limit = limitMax
		}
	}
	if limit < 0 {
		err = errors.New("limit can't be negative")
	}
	return limit, err
}
