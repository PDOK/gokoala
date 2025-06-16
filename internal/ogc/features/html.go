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
	WebConfig          *config.WebConfig

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
	WebConfig          *config.WebConfig
}

func (hf *htmlFeatures) features(w http.ResponseWriter, r *http.Request,
	collection config.GeoSpatialCollection, cursor domain.Cursors,
	featuresURL featureCollectionURL, limit int, referenceDate *time.Time,
	propertyFilters map[string]string,
	configuredPropertyFilters datasources.PropertyFiltersWithAllowedValues,
	fc *domain.FeatureCollection) {

	breadcrumbs := collectionsBreadcrumb
	breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
		{
			Name: getCollectionTitle(collection.ID, collection.Metadata),
			Path: collectionsCrumb + collection.ID,
		},
		{
			Name: "Items",
			Path: collectionsCrumb + collection.ID + "/items",
		},
	}...)

	if referenceDate.IsZero() {
		referenceDate = nil
	}
	var mapSheetProps *config.MapSheetDownloadProperties
	var wc *config.WebConfig
	if collection.Features != nil {
		if collection.Features.MapSheetDownloads != nil {
			mapSheetProps = &collection.Features.MapSheetDownloads.Properties
		}
		wc = collection.Features.Web
	}

	pageContent := &featureCollectionPage{
		*fc,
		collection.ID,
		collection.Metadata,
		cursor,
		featuresURL.toPrevNextURL(collection.ID, cursor.Prev, engine.FormatHTML),
		featuresURL.toPrevNextURL(collection.ID, cursor.Next, engine.FormatHTML),
		limit,
		referenceDate,
		mapSheetProps,
		wc,
		propertyFilters,
		configuredPropertyFilters,
	}

	lang := hf.engine.CN.NegotiateLanguage(w, r)
	hf.engine.RenderAndServe(w, r, engine.ExpandTemplateKey(featuresKey, lang), pageContent, breadcrumbs)
}

func (hf *htmlFeatures) feature(w http.ResponseWriter, r *http.Request,
	collection config.GeoSpatialCollection, feat *domain.Feature) {

	breadcrumbs := collectionsBreadcrumb
	breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
		{
			Name: getCollectionTitle(collection.ID, collection.Metadata),
			Path: collectionsCrumb + collection.ID,
		},
		{
			Name: "Items",
			Path: collectionsCrumb + collection.ID + "/items",
		},
		{
			Name: feat.ID,
			Path: collectionsCrumb + collection.ID + "/items/" + feat.ID,
		},
	}...)

	var mapSheetProps *config.MapSheetDownloadProperties
	var wc *config.WebConfig
	if collection.Features != nil {
		if collection.Features.MapSheetDownloads != nil {
			mapSheetProps = &collection.Features.MapSheetDownloads.Properties
		}
		wc = collection.Features.Web
	}

	pageContent := &featurePage{
		*feat,
		collection.ID,
		feat.ID,
		collection.Metadata,
		mapSheetProps,
		wc,
	}

	lang := hf.engine.CN.NegotiateLanguage(w, r)
	hf.engine.RenderAndServe(w, r, engine.ExpandTemplateKey(featureKey, lang), pageContent, breadcrumbs)
}

func getCollectionTitle(collectionID string, metadata *config.GeoSpatialCollectionMetadata) string {
	if metadata != nil && metadata.Title != nil {
		return *metadata.Title
	}
	return collectionID
}
