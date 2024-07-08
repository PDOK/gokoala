package features

import (
	"net/http"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

const (
	collectionsCrumb = "collections/"
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
	engine *engine.Engine
}

func newHTMLFeatures(e *engine.Engine) *htmlFeatures {
	e.ParseTemplate(featuresKey)
	e.ParseTemplate(featureKey)

	return &htmlFeatures{
		engine: e,
	}
}

// featureCollectionPage enriched FeatureCollection for HTML representation.
type featureCollectionPage struct {
	domain.FeatureCollection

	CollectionID       string
	Metadata           *config.GeoSpatialCollectionMetadata
	Cursor             domain.Cursors
	PrevLink           string
	NextLink           string
	Limit              int
	ReferenceDate      *time.Time
	MapSheetProperties *config.MapSheetDownloadProperties

	// Property filters as supplied by the user in the URL: filter name + value(s)
	PropertyFilters map[string]string
	// Property filters as specified in the (YAML) config, enriched with allowed values. Does not contain user supplied values
	ConfiguredPropertyFilters map[string]datasources.PropertyFilterWithAllowedValues
}

// featurePage enriched Feature for HTML representation.
type featurePage struct {
	domain.Feature

	CollectionID       string
	FeatureID          string
	Metadata           *config.GeoSpatialCollectionMetadata
	MapSheetProperties *config.MapSheetDownloadProperties
}

func (hf *htmlFeatures) features(w http.ResponseWriter, r *http.Request, collectionID string, cursor domain.Cursors,
	featuresURL featureCollectionURL, limit int, referenceDate *time.Time,
	propertyFilters map[string]string, configuredPropertyFilters datasources.PropertyFiltersWithAllowedValues,
	mapSheetProperties *config.MapSheetDownloadProperties, fc *domain.FeatureCollection) {

	collectionMetadata := collections[collectionID]

	breadcrumbs := collectionsBreadcrumb
	breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
		{
			Name: getCollectionTitle(collectionID, collectionMetadata),
			Path: collectionsCrumb + collectionID,
		},
		{
			Name: "Items",
			Path: collectionsCrumb + collectionID + "/items",
		},
	}...)

	if referenceDate.IsZero() {
		referenceDate = nil
	}

	pageContent := &featureCollectionPage{
		*fc,
		collectionID,
		collectionMetadata,
		cursor,
		featuresURL.toPrevNextURL(collectionID, cursor.Prev, engine.FormatHTML),
		featuresURL.toPrevNextURL(collectionID, cursor.Next, engine.FormatHTML),
		limit,
		referenceDate,
		mapSheetProperties,
		propertyFilters,
		configuredPropertyFilters,
	}

	lang := hf.engine.CN.NegotiateLanguage(w, r)
	hf.engine.RenderAndServePage(w, r, engine.ExpandTemplateKey(featuresKey, lang), pageContent, breadcrumbs)
}

func (hf *htmlFeatures) feature(w http.ResponseWriter, r *http.Request, collectionID string,
	mapSheetProperties *config.MapSheetDownloadProperties, feat *domain.Feature) {
	collectionMetadata := collections[collectionID]

	breadcrumbs := collectionsBreadcrumb
	breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
		{
			Name: getCollectionTitle(collectionID, collectionMetadata),
			Path: collectionsCrumb + collectionID,
		},
		{
			Name: "Items",
			Path: collectionsCrumb + collectionID + "/items",
		},
		{
			Name: feat.ID,
			Path: collectionsCrumb + collectionID + "/items/" + feat.ID,
		},
	}...)

	pageContent := &featurePage{
		*feat,
		collectionID,
		feat.ID,
		collectionMetadata,
		mapSheetProperties,
	}

	lang := hf.engine.CN.NegotiateLanguage(w, r)
	hf.engine.RenderAndServePage(w, r, engine.ExpandTemplateKey(featureKey, lang), pageContent, breadcrumbs)
}

func getCollectionTitle(collectionID string, metadata *config.GeoSpatialCollectionMetadata) string {
	title := collectionID
	if metadata != nil && metadata.Title != nil {
		title = *metadata.Title
	}
	return title
}
