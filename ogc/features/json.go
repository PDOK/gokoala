package features

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/features/domain"
)

type jsonFeatures struct {
	engine *engine.Engine
}

func newJSONFeatures(e *engine.Engine) *jsonFeatures {
	return &jsonFeatures{
		engine: e,
	}
}

func (jf *jsonFeatures) featuresAsGeoJSON(w http.ResponseWriter, r *http.Request, collectionID string,
	cursor domain.Cursors, featuresURL featureCollectionURL, fc *domain.FeatureCollection) {

	fc.Links = jf.createFeatureCollectionLinks(collectionID, cursor, featuresURL)
	fcJSON, err := toJSON(&fc)
	if err != nil {
		http.Error(w, "Failed to marshal FeatureCollection to JSON", http.StatusInternalServerError)
		return
	}
	jf.engine.ServeResponse(w, r, false /* performed earlier */, engine.MediaTypeGeoJSON, fcJSON)
}

func (jf *jsonFeatures) featureAsGeoJSON(w http.ResponseWriter, r *http.Request, collectionID string,
	feat *domain.Feature, url featureURL) {

	feat.Links = jf.createFeatureLinks(url, collectionID, feat.ID)
	featJSON, err := toJSON(feat)
	if err != nil {
		http.Error(w, "Failed to marshal Feature to JSON", http.StatusInternalServerError)
		return
	}
	jf.engine.ServeResponse(w, r, false /* performed earlier */, engine.MediaTypeGeoJSON, featJSON)
}

func (jf *jsonFeatures) featuresAsJSONFG(w http.ResponseWriter, r *http.Request, collectionID string,
	cursor domain.Cursors, featuresURL featureCollectionURL, fc *domain.FeatureCollection, crs ContentCrs) {

	fgFC := domain.JSONFGFeatureCollection{}
	for _, f := range fc.Features {
		fgF := domain.JSONFGFeature{
			ID:         f.ID,
			Links:      f.Links,
			Place:      f.Geometry,
			Properties: f.Properties,
		}
		fgFC.Features = append(fgFC.Features, &fgF)
		fgFC.CoordRefSys = string(crs)
	}

	fgFC.Links = jf.createFeatureCollectionLinks(collectionID, cursor, featuresURL)
	featJSON, err := toJSON(&fgFC)
	if err != nil {
		http.Error(w, "Failed to marshal Feature to JSON", http.StatusInternalServerError)
		return
	}
	jf.engine.ServeResponse(w, r, false /* performed earlier */, engine.MediaTypeJSONFG, featJSON)
}

func (jf *jsonFeatures) featureAsJSONFG(w http.ResponseWriter, r *http.Request, collectionID string,
	f *domain.Feature, url featureURL, crs ContentCrs) {

	fgF := domain.JSONFGFeature{
		ID:          f.ID,
		Links:       f.Links,
		CoordRefSys: string(crs),
		Place:       f.Geometry,
		Properties:  f.Properties,
	}

	fgF.Links = jf.createFeatureLinks(url, collectionID, fgF.ID)
	featJSON, err := toJSON(&fgF)
	if err != nil {
		http.Error(w, "Failed to marshal Feature to JSON", http.StatusInternalServerError)
		return
	}
	jf.engine.ServeResponse(w, r, false /* performed earlier */, engine.MediaTypeJSONFG, featJSON)
}

func (jf *jsonFeatures) createFeatureCollectionLinks(collectionID string, cursor domain.Cursors, featuresURL featureCollectionURL) []domain.Link {
	links := make([]domain.Link, 0)
	links = append(links, domain.Link{
		Rel:   "self",
		Title: "This document as GeoJSON",
		Type:  engine.MediaTypeGeoJSON,
		Href:  featuresURL.toSelfURL(collectionID, engine.FormatJSON),
	})
	links = append(links, domain.Link{
		Rel:   "alternate",
		Title: "This document as HTML",
		Type:  engine.MediaTypeHTML,
		Href:  featuresURL.toSelfURL(collectionID, engine.FormatHTML),
	})
	if cursor.HasNext {
		links = append(links, domain.Link{
			Rel:   "next",
			Title: "Next page",
			Type:  engine.MediaTypeGeoJSON,
			Href:  featuresURL.toPrevNextURL(collectionID, cursor.Next, engine.FormatJSON),
		})
	}
	if cursor.HasPrev {
		links = append(links, domain.Link{
			Rel:   "prev",
			Title: "Previous page",
			Type:  engine.MediaTypeGeoJSON,
			Href:  featuresURL.toPrevNextURL(collectionID, cursor.Prev, engine.FormatJSON),
		})
	}
	return links
}

func (jf *jsonFeatures) createFeatureLinks(url featureURL, collectionID string, featureID int64) []domain.Link {
	links := make([]domain.Link, 0)
	links = append(links, domain.Link{
		Rel:   "self",
		Title: "This document as GeoJSON",
		Type:  engine.MediaTypeGeoJSON,
		Href:  url.toSelfURL(collectionID, featureID, engine.FormatJSON),
	})
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
func toJSON(input interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(input)
	marshalled := bytes.TrimRight(buffer.Bytes(), "\n")
	return marshalled, err
}
