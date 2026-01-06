package search

import (
	"net/http"
	"net/url"
	"time"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

var (
	now = time.Now // allow mocking
)

type jsonSearchResults struct {
	engine           *engine.Engine
	validateResponse bool
}

func newJSONSearchResults(e *engine.Engine) *jsonSearchResults {
	return &jsonSearchResults{
		engine:           e,
		validateResponse: *e.Config.OgcAPI.FeaturesSearch.ValidateResponses,
	}
}

// GeoJSON.
func (jsr *jsonSearchResults) searchResultsAsGeoJSON(w http.ResponseWriter, r *http.Request, baseURL url.URL,
	fc *domain.FeatureCollection) {

	fc.Timestamp = now().Format(time.RFC3339)
	fc.Links = createLinks(baseURL)

	jsr.serve(&fc, engine.MediaTypeGeoJSON, r, w)
}

// JSON-FG.
func (jsr *jsonSearchResults) searchResultsAsJSONFG(w http.ResponseWriter, r *http.Request, baseURL url.URL,
	fc *domain.FeatureCollection, crs domain.ContentCrs) {

	fgFC := domain.JSONFGFeatureCollection{}
	fgFC.ConformsTo = []string{domain.ConformanceJSONFGCore}
	fgFC.CoordRefSys = string(crs)
	if len(fc.Features) == 0 {
		fgFC.Features = make([]*domain.JSONFGFeature, 0)
	} else {
		for _, f := range fc.Features {
			fgF := domain.JSONFGFeature{
				ID:         f.ID,
				Links:      f.Links,
				Properties: f.Properties,
			}
			fgF.SetGeom(crs, f.Geometry)
			fgFC.Features = append(fgFC.Features, &fgF)
		}
	}
	fgFC.NumberReturned = fc.NumberReturned
	fgFC.Timestamp = now().Format(time.RFC3339)
	fgFC.Links = createLinks(baseURL)

	jsr.serve(&fc, engine.MediaTypeJSONFG, r, w)
}

func (jsr *jsonSearchResults) serve(input any, contentType string, r *http.Request, w http.ResponseWriter) {
	jsr.engine.Serve(w, r,
		engine.ServeJSON(input),
		engine.ServeValidation(false /* performed earlier */, jsr.validateResponse),
		engine.ServeContentType(contentType))
}

func createLinks(baseURL url.URL) []domain.Link {
	links := make([]domain.Link, 0)

	href := baseURL.JoinPath("search")
	query := href.Query()
	query.Set(engine.FormatParam, engine.FormatJSON)
	href.RawQuery = query.Encode()

	links = append(links, domain.Link{
		Rel:   "self",
		Title: "This document as GeoJSON",
		Type:  engine.MediaTypeGeoJSON,
		Href:  href.String(),
	})
	// TODO: support HTML and JSON-FG output in location API
	//  links = append(links, domain.Link{
	//	Rel:   "alternate",
	//	Title: "This document as JSON-FG",
	//	Type:  engine.MediaTypeJSONFG,
	//	Href:  featuresURL.toSelfURL(collectionID, engine.FormatJSONFG),
	//  })
	//  links = append(links, domain.Link{
	//	Rel:   "alternate",
	//	Title: "This document as HTML",
	//	Type:  engine.MediaTypeHTML,
	//	Href:  featuresURL.toSelfURL(collectionID, engine.FormatHTML),
	//  })
	return links
}
