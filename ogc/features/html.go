package features

import (
	"net/http"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/features/domain"
)

var (
	collectionsBreadcrumb = []engine.Breadcrumb{
		{
			Name: "Collections",
			Path: "collections",
		},
	}
	featuresKey = engine.NewTemplateKey(templatesDir + "features.go.html")
	featureKey  = engine.NewTemplateKey(templatesDir + "feature.go.html")
)

type htmlFeatures struct {
	engine            *engine.Engine
	compiledTemplates map[engine.TemplateKey]interface{}
}

func newHTMLFeatures(e *engine.Engine) *htmlFeatures {
	compiledTemplates := make(map[engine.TemplateKey]interface{}, 4)
	compiledTemplates = mergeMaps(compiledTemplates, e.CompileTemplate(featuresKey))
	compiledTemplates = mergeMaps(compiledTemplates, e.CompileTemplate(featureKey))

	return &htmlFeatures{
		engine:            e,
		compiledTemplates: compiledTemplates,
	}
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

func (hf *htmlFeatures) features(w http.ResponseWriter, r *http.Request, collectionID string,
	cursor domain.Cursor, limit int, fc *domain.FeatureCollection, format string) {

	collectionMetadata := collectionsMetadata[collectionID]

	breadcrumbs := collectionsBreadcrumb
	breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
		{
			Name: getCollectionTitle(collectionID, collectionMetadata),
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

	lang := hf.engine.CN.NegotiateLanguage(w, r)
	key := engine.ExpandTemplateKey(featuresKey, lang)
	hf.engine.RenderAndServePage(w, r, key, hf.compiledTemplates[key], pageContent, breadcrumbs)
}

func (hf *htmlFeatures) feature(w http.ResponseWriter, r *http.Request, collectionID string,
	featureID string, feat *domain.Feature, format string) {

	collectionMetadata := collectionsMetadata[collectionID]

	breadcrumbs := collectionsBreadcrumb
	breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
		{
			Name: getCollectionTitle(collectionID, collectionMetadata),
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

	lang := hf.engine.CN.NegotiateLanguage(w, r)
	key := engine.ExpandTemplateKey(featureKey, lang)
	hf.engine.RenderAndServePage(w, r, key, hf.compiledTemplates[key], pageContent, breadcrumbs)
}

func getCollectionTitle(collectionID string, metadata *engine.GeoSpatialCollectionMetadata) string {
	title := collectionID
	if metadata != nil && metadata.Title != nil {
		title = *metadata.Title
	}
	return title
}

func mergeMaps(
	m1 map[engine.TemplateKey]interface{},
	m2 map[engine.TemplateKey]interface{}) map[engine.TemplateKey]interface{} {

	merged := make(map[engine.TemplateKey]interface{})
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}
