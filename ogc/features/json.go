package features

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func (jf *jsonFeatures) featuresAsGeoJSON(w http.ResponseWriter, collectionID string,
	cursor domain.Cursors, filters string, fc *domain.FeatureCollection) {

	fc.Links = jf.createFeatureCollectionLinks(collectionID, cursor, filters)
	fcJSON, err := toJSON(&fc)
	if err != nil {
		http.Error(w, "Failed to marshal FeatureCollection to JSON", http.StatusInternalServerError)
		return
	}
	engine.SafeWrite(w.Write, fcJSON)
}

func (jf *jsonFeatures) featureAsGeoJSON(w http.ResponseWriter, collectionID string, feat *domain.Feature) {
	feat.Links = jf.createFeatureLinks(collectionID, feat.ID)
	featJSON, err := toJSON(feat)
	if err != nil {
		http.Error(w, "Failed to marshal Feature to JSON", http.StatusInternalServerError)
		return
	}
	engine.SafeWrite(w.Write, featJSON)
}

func (jf *jsonFeatures) featuresAsJSONFG() {
	// TODO: not implemented yet
}

func (jf *jsonFeatures) featureAsJSONFG() {
	// TODO: not implemented yet
}

func (jf *jsonFeatures) createFeatureCollectionLinks(collectionID string, cursor domain.Cursors, filters string) []domain.Link {
	featuresBaseURL := fmt.Sprintf("%s/collections/%s/items", jf.engine.Config.BaseURL.String(), collectionID)

	links := make([]domain.Link, 0)
	links = append(links, domain.Link{
		Rel:   "self",
		Title: "This document as GeoJSON",
		Type:  engine.MediaTypeGeoJSON,
		Href:  featuresBaseURL + "?f=json",
	})
	links = append(links, domain.Link{
		Rel:   "alternate",
		Title: "This document as HTML",
		Type:  engine.MediaTypeHTML,
		Href:  featuresBaseURL + "?f=html",
	})
	if cursor.HasNext {
		links = append(links, domain.Link{
			Rel:   "next",
			Title: "Next page",
			Type:  engine.MediaTypeGeoJSON,
			Href:  fmt.Sprintf("%s?f=json&cursor=%s&%s", featuresBaseURL, cursor.Next, filters),
		})
	}
	if cursor.HasPrev {
		links = append(links, domain.Link{
			Rel:   "prev",
			Title: "Previous page",
			Type:  engine.MediaTypeGeoJSON,
			Href:  fmt.Sprintf("%s?f=json&cursor=%s&%s", featuresBaseURL, cursor.Prev, filters),
		})
	}
	return links
}

func (jf *jsonFeatures) createFeatureLinks(collectionID string, featureID int64) []domain.Link {
	featureBaseURL := fmt.Sprintf("%s/collections/%s/items/%d", jf.engine.Config.BaseURL.String(), collectionID, featureID)

	links := make([]domain.Link, 0)
	links = append(links, domain.Link{
		Rel:   "self",
		Title: "This document as GeoJSON",
		Type:  engine.MediaTypeGeoJSON,
		Href:  featureBaseURL + "?f=json",
	})
	links = append(links, domain.Link{
		Rel:   "alternate",
		Title: "This document as HTML",
		Type:  engine.MediaTypeHTML,
		Href:  featureBaseURL + "?f=html",
	})
	links = append(links, domain.Link{
		Rel:   "collection",
		Title: "The collection to which this feature belongs",
		Type:  engine.MediaTypeJSON,
		Href:  fmt.Sprintf("%s/collections/%s?f=json", jf.engine.Config.BaseURL.String(), collectionID),
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
