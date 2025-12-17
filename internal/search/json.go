package search

import (
	"bytes"
	stdjson "encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/search/domain"
	perfjson "github.com/goccy/go-json"
)

var (
	now                            = time.Now // allow mocking
	disableJSONPerfOptimization, _ = strconv.ParseBool(os.Getenv("DISABLE_JSON_PERF_OPTIMIZATION"))
)

type jsonFeatures struct {
	engine           *engine.Engine
	validateResponse bool
}

func newJSONFeatures(e *engine.Engine) *jsonFeatures {
	return &jsonFeatures{
		engine:           e,
		validateResponse: true, // TODO make configurable
	}
}

func (jf *jsonFeatures) featuresAsGeoJSON(w http.ResponseWriter, r *http.Request, baseURL url.URL, fc *domain.FeatureCollection) {
	fc.Timestamp = now().Format(time.RFC3339)
	fc.Links = createFeatureCollectionLinks(baseURL) // TODO add links

	if jf.validateResponse {
		jf.serveAndValidateJSON(&fc, engine.MediaTypeGeoJSON, r, w)
	} else {
		jf.serveJSON(&fc, engine.MediaTypeGeoJSON, w)
	}
}

// serveAndValidateJSON serves JSON after performing OpenAPI response validation.
func (jf *jsonFeatures) serveAndValidateJSON(input any, contentType string, r *http.Request, w http.ResponseWriter) {
	json := &bytes.Buffer{}
	if err := getEncoder(json).Encode(input); err != nil {
		handleJSONEncodingFailure(err, w)
		return
	}
	jf.engine.Serve(w, r,
		engine.ServeValidation(false /* performed earlier */, jf.validateResponse),
		engine.ServeContentType(contentType),
		engine.ServeOutput(json.Bytes()))
}

// serveJSON serves JSON *WITHOUT* OpenAPI validation by writing directly to the response output stream
func (jf *jsonFeatures) serveJSON(input any, contentType string, w http.ResponseWriter) {
	w.Header().Set(engine.HeaderContentType, contentType)

	if err := getEncoder(w).Encode(input); err != nil {
		handleJSONEncodingFailure(err, w)
		return
	}
}

func createFeatureCollectionLinks(baseURL url.URL) []domain.Link {
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

type jsonEncoder interface {
	Encode(input any) error
}

// Create JSONEncoder. Note escaping of '<', '>' and '&' is disabled (HTMLEscape is false).
// Especially the '&' is important since we use this character in the next/prev links.
func getEncoder(w io.Writer) jsonEncoder {
	if disableJSONPerfOptimization {
		// use Go stdlib JSON encoder
		encoder := stdjson.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		return encoder
	}
	// use ~7% overall faster 3rd party JSON encoder (in case of issues switch back to stdlib using env variable)
	encoder := perfjson.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder
}

func handleJSONEncodingFailure(err error, w http.ResponseWriter) {
	log.Printf("JSON encoding failed: %v", err)
	engine.RenderProblem(engine.ProblemServerError, w, "Failed to write JSON response")
}
