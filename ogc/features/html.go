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
)

type HTMLFeatures struct {
	engine *engine.Engine
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

func (hf *HTMLFeatures) features(w http.ResponseWriter, r *http.Request, collectionID string,
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
	key := engine.NewTemplateKeyWithLanguage(templatesDir+"features.go."+format, lang)
	hf.engine.RenderAndServePage(w, r, pageContent, breadcrumbs, key, lang)
}

func (hf *HTMLFeatures) feature(w http.ResponseWriter, r *http.Request, collectionID string,
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
	key := engine.NewTemplateKeyWithLanguage(templatesDir+"feature.go."+format, lang)
	hf.engine.RenderAndServePage(w, r, pageContent, breadcrumbs, key, lang)
}

func getCollectionTitle(collectionID string, metadata *engine.GeoSpatialCollectionMetadata) string {
	title := collectionID
	if metadata != nil && metadata.Title != nil {
		title = *metadata.Title
	}
	return title
}
