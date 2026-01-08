package features_search

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
func (jsr *jsonSearchResults) asGeoJSON(w http.ResponseWriter, r *http.Request, baseURL url.URL,
	fc *domain.FeatureCollection) {

	fc.Timestamp = now().Format(time.RFC3339)
	fc.Links = createLinks(baseURL)

	jsr.serve(&fc, engine.MediaTypeGeoJSON, r, w)
}

// JSON-FG.
func (jsr *jsonSearchResults) asJSONFG(w http.ResponseWriter, r *http.Request, baseURL url.URL,
	fc *domain.FeatureCollection, crs domain.ContentCrs) {

	fgFC := domain.FeatureCollectionToJSONFG(*fc, crs)
	fgFC.Timestamp = now().Format(time.RFC3339)
	fgFC.Links = createLinks(baseURL)

	jsr.serve(&fgFC, engine.MediaTypeJSONFG, r, w)
}

func (jsr *jsonSearchResults) serve(input any, contentType string, r *http.Request, w http.ResponseWriter) {
	jsr.engine.Serve(w, r,
		engine.ServeJSON(input),
		engine.ServeValidation(false /* performed earlier */, jsr.validateResponse),
		engine.ServeContentType(contentType))
}

func createLinks(baseURL url.URL) []domain.Link {
	links := make([]domain.Link, 0)

	links = append(links, domain.Link{
		Rel:   "self",
		Title: "This document as GeoJSON",
		Type:  engine.MediaTypeGeoJSON,
		Href:  toSelfURL(baseURL, engine.FormatJSON),
	})
	links = append(links, domain.Link{
		Rel:   "alternate",
		Title: "This document as JSON-FG",
		Type:  engine.MediaTypeJSONFG,
		Href:  toSelfURL(baseURL, engine.FormatJSONFG),
	})
	links = append(links, domain.Link{
		Rel:   "alternate",
		Title: "This document as HTML",
		Type:  engine.MediaTypeHTML,
		Href:  toSelfURL(baseURL, engine.FormatHTML),
	})
	return links
}

func toSelfURL(baseURL url.URL, format string) string {
	href := baseURL.JoinPath("search")
	query := href.Query()
	query.Set(engine.FormatParam, format)
	href.RawQuery = query.Encode()
	return href.String()
}
