package features

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/common/geospatial"
	"github.com/PDOK/gokoala/ogc/features/datasources"
	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/go-chi/chi/v5"
)

const (
	templatesDir = "ogc/features/templates/"
	defaultLimit = 10
)

var (
	collectionsBreadcrumb = []engine.Breadcrumb{
		{
			Name: "Collections",
			Path: "collections",
		},
	}
	collectionsMetadata map[string]*engine.GeoSpatialCollectionMetadata
)

type Features struct {
	engine     *engine.Engine
	datasource datasources.Datasource
}

func NewFeatures(e *engine.Engine, router *chi.Mux) *Features {
	var datasource datasources.Datasource
	if e.Config.OgcAPI.Features.Datasource.FakeDB {
		datasource = datasources.NewFakeDB()
	} else if e.Config.OgcAPI.Features.Datasource.GeoPackage != nil {
		datasource = datasources.NewGeoPackage()
	}
	e.RegisterShutdownHook(datasource.Close)

	features := &Features{
		engine:     e,
		datasource: datasource,
	}
	collectionsMetadata = features.cacheCollectionsMetadata()

	router.Get(geospatial.CollectionsPath+"/{collectionId}/items", features.CollectionContent())
	router.Get(geospatial.CollectionsPath+"/{collectionId}/items/{featureId}", features.Feature())
	return features
}

// CollectionContent serve a FeatureCollection with the given collectionId
func (f *Features) CollectionContent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")
		cursorParam := r.URL.Query().Get("cursor")
		limit, err := f.getLimit(r)
		if err != nil {
			http.Error(w, "limit should be a number", http.StatusBadRequest)
			return
		}

		fc, cursor := f.datasource.GetFeatures(collectionID, cursorParam, limit)
		if fc == nil {
			http.NotFound(w, r)
			return
		}

		format := f.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatHTML:
			f.featuresAsHTML(w, r, collectionID, cursor, limit, fc, format)
		case engine.FormatJSON:
			f.featuresAsJSON(w, collectionID, cursor, limit, fc)
		default:
			http.NotFound(w, r)
		}
	}
}

// Feature serves a single Feature
func (f *Features) Feature() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")
		featureID := chi.URLParam(r, "featureId")

		feat := f.datasource.GetFeature(collectionID, featureID)
		if feat == nil {
			http.NotFound(w, r)
			return
		}

		format := f.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatHTML:
			f.featureAsHTML(w, r, collectionID, featureID, feat, format)
		case engine.FormatJSON:
			f.featureAsJSON(w, feat)
		default:
			http.NotFound(w, r)
		}
	}
}

func (f *Features) featuresAsHTML(w http.ResponseWriter, r *http.Request, collectionID string,
	cursor domain.Cursor, limit int, fc *domain.FeatureCollection, format string) {

	collectionMetadata := collectionsMetadata[collectionID]

	breadcrumbs := collectionsBreadcrumb
	breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
		{
			Name: f.getCollectionTitle(collectionID, collectionMetadata),
			Path: "collections/" + collectionID,
		},
		{
			Name: "Items",
			Path: "collections/" + collectionID + "/items",
		},
	}...)

	pageContent := &featureCollectionPage{
		*fc,
		collectionID,
		collectionMetadata,
		cursor,
		limit,
	}

	lang := f.engine.CN.NegotiateLanguage(w, r)
	key := engine.NewTemplateKeyWithLanguage(templatesDir+"features.go."+format, lang)
	f.engine.RenderAndServePage(w, r, pageContent, breadcrumbs, key, lang)
}

func (f *Features) featureAsHTML(w http.ResponseWriter, r *http.Request, collectionID string,
	featureID string, feat *domain.Feature, format string) {

	collectionMetadata := collectionsMetadata[collectionID]

	breadcrumbs := collectionsBreadcrumb
	breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
		{
			Name: f.getCollectionTitle(collectionID, collectionMetadata),
			Path: "collections/" + collectionID,
		},
		{
			Name: "Items",
			Path: "collections/" + collectionID + "/items",
		},
		{
			Name: featureID,
			Path: "collections/" + collectionID + "/items/" + featureID,
		},
	}...)

	pageContent := &featurePage{
		*feat,
		featureID,
		collectionMetadata,
	}

	lang := f.engine.CN.NegotiateLanguage(w, r)
	key := engine.NewTemplateKeyWithLanguage(templatesDir+"feature.go."+format, lang)
	f.engine.RenderAndServePage(w, r, pageContent, breadcrumbs, key, lang)
}

func (f *Features) featuresAsJSON(w http.ResponseWriter, collectionID string,
	cursor domain.Cursor, limit int, fc *domain.FeatureCollection) {

	fc.Links = f.createJSONLinks(collectionID, cursor, limit)
	fcJSON, err := json.Marshal(&fc)
	if err != nil {
		http.Error(w, "Failed to marshal FeatureCollection to JSON", http.StatusInternalServerError)
		return
	}
	engine.SafeWrite(w.Write, fcJSON)
}

func (f *Features) featureAsJSON(w http.ResponseWriter, feat *domain.Feature) {
	featJSON, err := json.Marshal(feat)
	if err != nil {
		http.Error(w, "Failed to marshal Feature to JSON", http.StatusInternalServerError)
		return
	}
	engine.SafeWrite(w.Write, featJSON)
}

// featureCollectionPage enriched FeatureCollection for HTML representation.
type featureCollectionPage struct {
	domain.FeatureCollection

	CollectionID string
	Metadata     *engine.GeoSpatialCollectionMetadata
	Cursor       domain.Cursor
	Limit        int
}

// featurePage enriched Feature for HTML representation.
type featurePage struct {
	domain.Feature

	FeatureID string
	Metadata  *engine.GeoSpatialCollectionMetadata
}

func (f *Features) cacheCollectionsMetadata() map[string]*engine.GeoSpatialCollectionMetadata {
	result := make(map[string]*engine.GeoSpatialCollectionMetadata)
	for _, collection := range f.engine.Config.OgcAPI.Features.Collections {
		result[collection.ID] = collection.Metadata
	}
	return result
}

func (f *Features) getCollectionTitle(collectionID string, metadata *engine.GeoSpatialCollectionMetadata) string {
	title := collectionID
	if metadata != nil && metadata.Title != nil {
		title = *metadata.Title
	}
	return title
}

func (f *Features) getLimit(r *http.Request) (int, error) {
	limit := defaultLimit
	var err error
	if r.URL.Query().Get("limit") != "" {
		limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
	}
	if limit < 0 {
		limit = 0
	}
	return limit, err
}

func (f *Features) createJSONLinks(collectionID string, cursor domain.Cursor, limit int) []domain.Link {
	featuresBaseURL := fmt.Sprintf("%s/collections/%s/items", f.engine.Config.BaseURL.String(), collectionID)

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
