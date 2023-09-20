package features

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/features/domain"
)

type JSONFeatures struct {
	engine *engine.Engine
}

func (jf *JSONFeatures) featuresAsGeoJSON(w http.ResponseWriter, collectionID string,
	cursor domain.Cursor, limit int, fc *domain.FeatureCollection) {

	fc.Links = jf.createJSONLinks(collectionID, cursor, limit)
	fcJSON, err := json.Marshal(&fc)
	if err != nil {
		http.Error(w, "Failed to marshal FeatureCollection to JSON", http.StatusInternalServerError)
		return
	}
	engine.SafeWrite(w.Write, fcJSON)
}

func (jf *JSONFeatures) featureAsGeoJSON(w http.ResponseWriter, feat *domain.Feature) {
	featJSON, err := json.Marshal(feat)
	if err != nil {
		http.Error(w, "Failed to marshal Feature to JSON", http.StatusInternalServerError)
		return
	}
	engine.SafeWrite(w.Write, featJSON)
}

func (jf *JSONFeatures) featuresAsJSONFG() {
	// TODO: not implemented yet
}

func (jf *JSONFeatures) featureAsJSONFG() {
	// TODO: not implemented yet
}

func (jf *JSONFeatures) createJSONLinks(collectionID string, cursor domain.Cursor, limit int) []domain.Link {
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
	if !cursor.IsLast {
		links = append(links, domain.Link{
			Rel:   "next",
			Title: "Next page",
			Type:  engine.MediaTypeGeoJSON,
			Href:  fmt.Sprintf("%s?f=json&cursor=%d&limit=%d", featuresBaseURL, cursor.Next, limit),
		})
	}
	if !cursor.IsFirst {
		links = append(links, domain.Link{
			Rel:   "prev",
			Title: "Previous page",
			Type:  engine.MediaTypeGeoJSON,
			Href:  fmt.Sprintf("%s?f=json&cursor=%d&limit=%d", featuresBaseURL, cursor.Prev, limit),
		})
	}
	return links
}
