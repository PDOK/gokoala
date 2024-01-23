package features

import (
	"bytes"
	stdjson "encoding/json"
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
	validateResponse *bool
}

func newJSONFeatures(e *engine.Engine) *jsonFeatures {
	return &jsonFeatures{
		engine:           e,
		validateResponse: e.Config.OgcAPI.Features.ValidateResponses,
	}
}

func (jf *jsonFeatures) featuresAsGeoJSON(w http.ResponseWriter, r *http.Request, collectionID string,
	cursor domain.Cursors, featuresURL featureCollectionURL, fc *domain.FeatureCollection) {

	fc.Timestamp = now().Format(time.RFC3339)
	fc.Links = jf.createFeatureCollectionLinks(engine.FormatGeoJSON, collectionID, cursor, featuresURL)
	fcJSON, err := toJSON(&fc)
	if err != nil {
		http.Error(w, "Failed to marshal FeatureCollection to JSON", http.StatusInternalServerError)
		return
	}
	jf.engine.ServeResponse(w, r, false /* performed earlier */, *jf.validateResponse, engine.MediaTypeGeoJSON, fcJSON)
}

func (jf *jsonFeatures) featureAsGeoJSON(w http.ResponseWriter, r *http.Request, collectionID string,
	feat *domain.Feature, url featureURL) {

	feat.Links = jf.createFeatureLinks(engine.FormatGeoJSON, url, collectionID, feat.ID)
	featJSON, err := toJSON(feat)
	if err != nil {
		http.Error(w, "Failed to marshal Feature to JSON", http.StatusInternalServerError)
		return
	}
	jf.engine.ServeResponse(w, r, false /* performed earlier */, *jf.validateResponse, engine.MediaTypeGeoJSON, featJSON)
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

	featJSON, err := toJSON(&fgFC)
	if err != nil {
		http.Error(w, "Failed to marshal Feature to JSON", http.StatusInternalServerError)
		return
	}
	jf.engine.ServeResponse(w, r, false /* performed earlier */, *jf.validateResponse, engine.MediaTypeJSONFG, featJSON)
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

	featJSON, err := toJSON(&fgF)
	if err != nil {
		http.Error(w, "Failed to marshal Feature to JSON", http.StatusInternalServerError)
		return
	}
	jf.engine.ServeResponse(w, r, false /* performed earlier */, *jf.validateResponse, engine.MediaTypeJSONFG, featJSON)
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

// toJSON performs the equivalent of json.Marshal but without escaping '<', '>' and '&'.
// Especially the '&' is important since we use this character in the next/prev links.
func toJSON(input any) ([]byte, error) {
	buffer := &bytes.Buffer{}
	if disableJSONPerfOptimization {
		// use Go stdlib JSON encoder
		encoder := stdjson.NewEncoder(buffer)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(input)
		return buffer.Bytes(), err
	}
	// use ~7% faster 3rd party JSON encoder
	encoder := perfjson.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(input)
	return buffer.Bytes(), err
}

func setGeom(crs ContentCrs, jsonfgFeature *domain.JSONFGFeature, feature *domain.Feature) {
	if crs.IsWGS84() {
		jsonfgFeature.Geometry = feature.Geometry
	} else {
		jsonfgFeature.Place = feature.Geometry
	}
}
