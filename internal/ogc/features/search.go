package features

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
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

// Search autosuggest locations based on user input
func (f *Features) Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate
		if err := f.engine.OpenAPI.ValidateRequest(r); err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}
		collections, searchTerms, outputSRID, outputCRS, bbox, bboxSRID, limit, err := parseQueryParams(r.URL.Query())
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}

		// Query expansion
		searchQuery, err := f.queryExpansion.Expand(r.Context(), searchTerms)
		if err != nil {
			handleSearchQueryError(w, err)
			return
		}

		// Perform actual search
		var datasource ds.Datasource
		for _, v := range f.datasources {
			datasource = v
			break // Stop after the first entry found
		}
		fc, err := datasource.SearchFeaturesAcrossCollections(r.Context(), *searchQuery, collections, outputSRID, bbox, bboxSRID, limit)
		if err != nil {
			handleSearchQueryError(w, err)
			return
		}
		if err = f.enrichFeaturesWithHref(fc, outputCRS); err != nil {
			engine.RenderProblem(engine.ProblemServerError, w, err.Error())
			return
		}

		// Output
		format := f.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatGeoJSON, engine.FormatJSON:
			f.json.searchResultsAsGeoJSON(w, r, *f.engine.Config.BaseURL.URL, fc)
		default:
			engine.RenderProblem(engine.ProblemNotAcceptable, w, fmt.Sprintf("format '%s' is not supported", format))
			return
		}
	}
}

//nolint:nestif
func (f *Features) enrichFeaturesWithHref(fc *domain.FeatureCollection, outputCRS string) error {
	for _, feat := range fc.Features {
		collectionID := feat.Properties.Value(domain.PropCollectionID)
		if collectionID == "" {
			return fmt.Errorf("collection reference not found in feature %s", feat.ID)
		}
		var collection *config.GeoSpatialCollection
		for _, coll := range f.engine.Config.AllCollections() {
			if collectionID == coll.ID && coll.Features != nil && coll.FeaturesSearch != nil {
				collection = &coll
				break
			}
		}
		if collection != nil {
			for _, ogcColl := range collection.FeaturesSearch.Search.OGCCollections {
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

					if outputCRS != "" {
						href += "&crs=" + outputCRS
					}

					// add href to feature both in GeoJSON properties (for broad compatibility and in line with OGC API Features part 5) and as a Link.
					feat.Properties.Set(domain.PropHref, href)
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

// log error, but send generic message to client to prevent possible information leakage from datasource
func handleSearchQueryError(w http.ResponseWriter, err error) {
	msg := "failed to fulfill search request"
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		// provide more context when user hits the query timeout
		msg += ": querying took too long (timeout encountered). Simplify your request and try again, or contact support"
	}
	log.Printf("%s, error: %v\n", msg, err)
	engine.RenderProblem(engine.ProblemServerError, w, msg) // don't include sensitive information in details msg
}
