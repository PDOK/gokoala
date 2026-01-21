package features

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

var (
	now = time.Now // allow mocking
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

// GeoJSON.
func (jf *jsonFeatures) featuresAsGeoJSON(w http.ResponseWriter, r *http.Request, collectionID string, cursor domain.Cursors,
	featuresURL featureCollectionURL, configuredFC *config.CollectionFeatures, fc *domain.FeatureCollection) {

	fc.Timestamp = now().Format(time.RFC3339)
	fc.Links = jf.createFeatureCollectionLinks(engine.FormatGeoJSON, collectionID, cursor, featuresURL)

	jf.createFeatureDownloadLinks(configuredFC, fc)

	jf.serve(&fc, engine.MediaTypeGeoJSON, r, w)
}

// GeoJSON.
func (jf *jsonFeatures) featureAsGeoJSON(w http.ResponseWriter, r *http.Request, collectionID string,
	configuredFC *config.CollectionFeatures, feat *domain.Feature, url featureURL) {

	feat.Links = jf.createFeatureLinks(engine.FormatGeoJSON, url, collectionID, feat.ID)
	if mapSheetProperties := getMapSheetProperties(configuredFC); mapSheetProperties != nil {
		feat.Links = append(feat.Links, domain.Link{
			Rel:   "enclosure",
			Title: "Download feature",
			Type:  mapSheetProperties.MediaType.String(),
			Href:  fmt.Sprintf("%v", feat.Properties.Value(mapSheetProperties.AssetURL)),
		})
	}

	jf.serve(&feat, engine.MediaTypeGeoJSON, r, w)
}

// GeoJSON for non-spatial data ("attribute JSON").
func (jf *jsonFeatures) featuresAsAttributeJSON(w http.ResponseWriter, r *http.Request, collectionID string, cursor domain.Cursors,
	featuresURL featureCollectionURL, fc *domain.FeatureCollection) {

	fgFC := domain.AttributeCollection{}
	if len(fc.Features) == 0 {
		fgFC.Features = make([]*domain.Attribute, 0)
	} else {
		for _, f := range fc.Features {
			fgF := domain.Attribute{
				ID:         f.ID,
				Links:      f.Links,
				Properties: f.Properties,
			}
			fgFC.Features = append(fgFC.Features, &fgF)
		}
	}
	fgFC.NumberReturned = fc.NumberReturned
	fgFC.Timestamp = now().Format(time.RFC3339)
	fgFC.Links = jf.createFeatureCollectionLinks(engine.FormatJSON, collectionID, cursor, featuresURL)

	jf.serve(&fgFC, engine.MediaTypeJSON, r, w)
}

// GeoJSON for non-spatial data ("attribute JSON").
func (jf *jsonFeatures) featureAsAttributeJSON(w http.ResponseWriter, r *http.Request, collectionID string,
	f *domain.Feature, url featureURL) {

	fgF := domain.Attribute{
		ID:         f.ID,
		Links:      f.Links,
		Properties: f.Properties,
	}
	fgF.Links = jf.createFeatureLinks(engine.FormatJSON, url, collectionID, fgF.ID)

	jf.serve(&fgF, engine.MediaTypeJSON, r, w)
}

// JSON-FG.
func (jf *jsonFeatures) featuresAsJSONFG(w http.ResponseWriter, r *http.Request, collectionID string, cursor domain.Cursors,
	featuresURL featureCollectionURL, configuredFC *config.CollectionFeatures, fc *domain.FeatureCollection, crs domain.ContentCrs) {

	fgFC := domain.FeatureCollectionToJSONFG(*fc, crs)
	fgFC.Timestamp = now().Format(time.RFC3339)
	fgFC.Links = jf.createFeatureCollectionLinks(engine.FormatJSONFG, collectionID, cursor, featuresURL)

	jf.createJSONFGFeatureDownloadLinks(configuredFC, &fgFC)

	jf.serve(&fgFC, engine.MediaTypeJSONFG, r, w)
}

// JSON-FG.
func (jf *jsonFeatures) featureAsJSONFG(w http.ResponseWriter, r *http.Request, collectionID string,
	configuredFC *config.CollectionFeatures, f *domain.Feature, url featureURL, crs domain.ContentCrs) {

	fgF := domain.JSONFGFeature{
		ID:          f.ID,
		Links:       f.Links,
		ConformsTo:  []string{domain.ConformanceJSONFGCore},
		CoordRefSys: string(crs),
		Properties:  f.Properties,
	}
	fgF.SetGeom(crs, f.Geometry)
	fgF.Links = jf.createFeatureLinks(engine.FormatJSONFG, url, collectionID, fgF.ID)
	if mapSheetProperties := getMapSheetProperties(configuredFC); mapSheetProperties != nil {
		fgF.Links = append(fgF.Links, domain.Link{
			Rel:   "enclosure",
			Title: "Download feature",
			Type:  mapSheetProperties.MediaType.String(),
			Href:  fmt.Sprintf("%v", fgF.Properties.Value(mapSheetProperties.AssetURL)),
		})
	}

	jf.serve(&fgF, engine.MediaTypeJSONFG, r, w)
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
	case engine.FormatJSON:
		links = append(links, domain.Link{
			Rel:   "self",
			Title: "This document as JSON",
			Type:  engine.MediaTypeJSON,
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
	collectionID string, featureID string) []domain.Link {

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
	case engine.FormatJSON:
		links = append(links, domain.Link{
			Rel:   "self",
			Title: "This document as JSON",
			Type:  engine.MediaTypeJSON,
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

func (jf *jsonFeatures) createFeatureDownloadLinks(configuredFC *config.CollectionFeatures, fc *domain.FeatureCollection) {
	if mapSheetProperties := getMapSheetProperties(configuredFC); mapSheetProperties != nil {
		for _, feature := range fc.Features {
			links := make([]domain.Link, 0)
			links = append(links, domain.Link{
				Rel:   "enclosure",
				Title: "Download feature",
				Type:  mapSheetProperties.MediaType.String(),
				Href:  fmt.Sprintf("%v", feature.Properties.Value(mapSheetProperties.AssetURL)),
			})
			feature.Links = links
		}
	}
}

func (jf *jsonFeatures) createJSONFGFeatureDownloadLinks(configuredFC *config.CollectionFeatures, fc *domain.JSONFGFeatureCollection) {
	if mapSheetProperties := getMapSheetProperties(configuredFC); mapSheetProperties != nil {
		for _, feature := range fc.Features {
			links := make([]domain.Link, 0)
			links = append(links, domain.Link{
				Rel:   "enclosure",
				Title: "Download feature",
				Type:  mapSheetProperties.MediaType.String(),
				Href:  fmt.Sprintf("%v", feature.Properties.Value(mapSheetProperties.AssetURL)),
			})
			feature.Links = links
		}
	}
}

func (jf *jsonFeatures) serve(input any, contentType string, r *http.Request, w http.ResponseWriter) {
	jf.engine.Serve(w, r,
		engine.ServeJSON(input),
		engine.ServeValidation(false /* performed earlier */, jf.validateResponse),
		engine.ServeContentType(contentType))
}

func getMapSheetProperties(configuredFC *config.CollectionFeatures) *config.MapSheetDownloadProperties {
	if configuredFC != nil && configuredFC.MapSheetDownloads != nil {
		return &configuredFC.MapSheetDownloads.Properties
	}
	return nil
}
