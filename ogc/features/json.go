package features

import (
	"bytes"
	stdjson "encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/features/domain"
	perfjson "github.com/goccy/go-json"
)

var (
	now                            = time.Now
	disableJSONPerfOptimization, _ = strconv.ParseBool(os.Getenv("DISABLE_JSON_PERF_OPTIMIZATION"))
)

type jsonFeatures struct {
	engine           *engine.Engine
	validateResponse bool
}

func newJSONFeatures(e *engine.Engine) *jsonFeatures {
	if *e.Config.OgcAPI.Features.ValidateResponses {
		log.Println("JSON response validation is enabled (by default). When serving large feature collections " +
			"set 'validateResponses' to 'false' to improve performance")
	}
	return &jsonFeatures{
		engine:           e,
		validateResponse: *e.Config.OgcAPI.Features.ValidateResponses,
	}
}

func (jf *jsonFeatures) featuresAsGeoJSON(w http.ResponseWriter, r *http.Request, collectionID string,
	cursor domain.Cursors, featuresURL featureCollectionURL, fc *domain.FeatureCollection) {

	fc.Timestamp = now().Format(time.RFC3339)
	fc.Links = jf.createFeatureCollectionLinks(engine.FormatGeoJSON, collectionID, cursor, featuresURL)

	if jf.validateResponse {
		jf.serveAndValidateJSON(&fc, engine.MediaTypeGeoJSON, r, w)
	} else {
		serveJSON(&fc, engine.MediaTypeGeoJSON, w)
	}
}

func (jf *jsonFeatures) featureAsGeoJSON(w http.ResponseWriter, r *http.Request, collectionID string,
	feat *domain.Feature, url featureURL) {

	feat.Links = jf.createFeatureLinks(engine.FormatGeoJSON, url, collectionID, feat.ID)
	if jf.validateResponse {
		jf.serveAndValidateJSON(&feat, engine.MediaTypeGeoJSON, r, w)
	} else {
		serveJSON(&feat, engine.MediaTypeGeoJSON, w)
	}
}

func (jf *jsonFeatures) featuresAsJSONFG(w http.ResponseWriter, r *http.Request, collectionID string,
	cursor domain.Cursors, featuresURL featureCollectionURL, fc *domain.FeatureCollection, crs ContentCrs) {

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
			setGeom(crs, &fgF, f)
			fgFC.Features = append(fgFC.Features, &fgF)
		}
	}
	fgFC.NumberReturned = fc.NumberReturned
	fgFC.Timestamp = now().Format(time.RFC3339)
	fgFC.Links = jf.createFeatureCollectionLinks(engine.FormatJSONFG, collectionID, cursor, featuresURL)

	if jf.validateResponse {
		jf.serveAndValidateJSON(&fgFC, engine.MediaTypeJSONFG, r, w)
	} else {
		serveJSON(&fgFC, engine.MediaTypeJSONFG, w)
	}
}

func (jf *jsonFeatures) featureAsJSONFG(w http.ResponseWriter, r *http.Request, collectionID string,
	f *domain.Feature, url featureURL, crs ContentCrs) {

	fgF := domain.JSONFGFeature{
		ID:          f.ID,
		Links:       f.Links,
		ConformsTo:  []string{domain.ConformanceJSONFGCore},
		CoordRefSys: string(crs),
		Properties:  f.Properties,
	}
	setGeom(crs, &fgF, f)
	fgF.Links = jf.createFeatureLinks(engine.FormatJSONFG, url, collectionID, fgF.ID)

	if jf.validateResponse {
		jf.serveAndValidateJSON(&fgF, engine.MediaTypeJSONFG, r, w)
	} else {
		serveJSON(&fgF, engine.MediaTypeJSONFG, w)
	}
}

func (jf *jsonFeatures) createFeatureCollectionLinks(currentFormat string, collectionID string,
	cursor domain.Cursors, featuresURL featureCollectionURL) []domain.Link {

	links := make([]domain.Link, 0)
	switch currentFormat {
	case engine.FormatGeoJSON:
		links = append(links, domain.Link{
			Rel:   "self",
			Title: "This document as GeoJSON",
			Type:  engine.MediaTypeGeoJSON,
			Href:  featuresURL.toSelfURL(collectionID, engine.FormatJSON),
		})
		links = append(links, domain.Link{
			Rel:   "alternate",
			Title: "This document as JSON-FG",
			Type:  engine.MediaTypeJSONFG,
			Href:  featuresURL.toSelfURL(collectionID, engine.FormatJSONFG),
		})
	case engine.FormatJSONFG:
		links = append(links, domain.Link{
			Rel:   "self",
			Title: "This document as JSON-FG",
			Type:  engine.MediaTypeJSONFG,
			Href:  featuresURL.toSelfURL(collectionID, engine.FormatJSONFG),
		})
		links = append(links, domain.Link{
			Rel:   "alternate",
			Title: "This document as GeoJSON",
			Type:  engine.MediaTypeGeoJSON,
			Href:  featuresURL.toSelfURL(collectionID, engine.FormatJSON),
		})
	}

	links = append(links, domain.Link{
		Rel:   "alternate",
		Title: "This document as HTML",
		Type:  engine.MediaTypeHTML,
		Href:  featuresURL.toSelfURL(collectionID, engine.FormatHTML),
	})

	if cursor.HasNext {
		switch currentFormat {
		case engine.FormatGeoJSON:
			links = append(links, domain.Link{
				Rel:   "next",
				Title: "Next page",
				Type:  engine.MediaTypeGeoJSON,
				Href:  featuresURL.toPrevNextURL(collectionID, cursor.Next, engine.FormatJSON),
			})
		case engine.FormatJSONFG:
			links = append(links, domain.Link{
				Rel:   "next",
				Title: "Next page",
				Type:  engine.MediaTypeJSONFG,
				Href:  featuresURL.toPrevNextURL(collectionID, cursor.Next, engine.FormatJSONFG),
			})
		}
	}

	if cursor.HasPrev {
		switch currentFormat {
		case engine.FormatGeoJSON:
			links = append(links, domain.Link{
				Rel:   "prev",
				Title: "Previous page",
				Type:  engine.MediaTypeGeoJSON,
				Href:  featuresURL.toPrevNextURL(collectionID, cursor.Prev, engine.FormatJSON),
			})
		case engine.FormatJSONFG:
			links = append(links, domain.Link{
				Rel:   "prev",
				Title: "Previous page",
				Type:  engine.MediaTypeJSONFG,
				Href:  featuresURL.toPrevNextURL(collectionID, cursor.Prev, engine.FormatJSONFG),
			})
		}
	}
	return links
}

func (jf *jsonFeatures) createFeatureLinks(currentFormat string, url featureURL,
	collectionID string, featureID int64) []domain.Link {

	links := make([]domain.Link, 0)
	switch currentFormat {
	case engine.FormatGeoJSON:
		links = append(links, domain.Link{
			Rel:   "self",
			Title: "This document as GeoJSON",
			Type:  engine.MediaTypeGeoJSON,
			Href:  url.toSelfURL(collectionID, featureID, engine.FormatJSON),
		})
		links = append(links, domain.Link{
			Rel:   "alternate",
			Title: "This document as JSON-FG",
			Type:  engine.MediaTypeJSONFG,
			Href:  url.toSelfURL(collectionID, featureID, engine.FormatJSONFG),
		})
	case engine.FormatJSONFG:
		links = append(links, domain.Link{
			Rel:   "self",
			Title: "This document as JSON-FG",
			Type:  engine.MediaTypeJSONFG,
			Href:  url.toSelfURL(collectionID, featureID, engine.FormatJSONFG),
		})
		links = append(links, domain.Link{
			Rel:   "alternate",
			Title: "This document as GeoJSON",
			Type:  engine.MediaTypeGeoJSON,
			Href:  url.toSelfURL(collectionID, featureID, engine.FormatJSON),
		})
	}
	links = append(links, domain.Link{
		Rel:   "alternate",
		Title: "This document as HTML",
		Type:  engine.MediaTypeHTML,
		Href:  url.toSelfURL(collectionID, featureID, engine.FormatHTML),
	})
	links = append(links, domain.Link{
		Rel:   "collection",
		Title: "The collection to which this feature belongs",
		Type:  engine.MediaTypeJSON,
		Href:  url.toCollectionURL(collectionID, engine.FormatJSON),
	})
	return links
}

// serveAndValidateJSON serves JSON after performing OpenAPI response validation.
// Note: this requires reading first marshalling to the result to JSON in-memory.
func (jf *jsonFeatures) serveAndValidateJSON(input any, contentType string, r *http.Request, w http.ResponseWriter) {
	json := &bytes.Buffer{}
	if err := getEncoder(json).Encode(input); err != nil {
		handleJSONEncodingFailure(err, w)
		return
	}
	jf.engine.ServeResponse(w, r, false /* performed earlier */, jf.validateResponse, contentType, json.Bytes())
}

// serveJSON serves JSON *WITHOUT* OpenAPI validation by writing directly to the response output stream
func serveJSON(input any, contentType string, w http.ResponseWriter) {
	w.Header().Set(engine.HeaderContentType, contentType)

	if err := getEncoder(w).Encode(input); err != nil {
		handleJSONEncodingFailure(err, w)
		return
	}
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
	http.Error(w, "Failed to write JSON response", http.StatusInternalServerError)
}

func setGeom(crs ContentCrs, jsonfgFeature *domain.JSONFGFeature, feature *domain.Feature) {
	if crs.IsWGS84() {
		jsonfgFeature.Geometry = feature.Geometry
	} else {
		jsonfgFeature.Place = feature.Geometry
	}
}
