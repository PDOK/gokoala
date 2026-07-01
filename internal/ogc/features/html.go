package features

import (
	"net/http"

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
	DateTime           domain.DateTime
	MapSheetProperties *config.MapSheetDownloadProperties
	WebConfig          *config.WebConfig
	ShowViewer         bool

	// Property filters as supplied by the user in the URL: filter name + value(s)
	PropertyFilters map[string]string
	// Property filters as specified in the (YAML) config, enriched with allowed values. Does not contain user supplied values
	ConfiguredPropertyFilters map[string]datasources.QueryableWithAllowedValues
}

// featurePage enriched Feature for HTML representation.
type featurePage struct {
	domain.Feature

	CollectionID       string
	FeatureID          string
	Metadata           *config.GeoSpatialCollectionMetadata
	MapSheetProperties *config.MapSheetDownloadProperties
	WebConfig          *config.WebConfig
	ShowViewer         bool
}

func (hf *htmlFeatures) features(w http.ResponseWriter, r *http.Request,
	collection config.FeaturesCollection, cursor domain.Cursors,
	featuresURL featureCollectionURL, limit int, dateTime domain.DateTime,
	propertyFilters map[string]string, queryables datasources.Queryables,
	fc *domain.FeatureCollection, outputFormats []engine.OutputFormat) {

	breadcrumbs, pageContent := hf.toItemsPage(collection, dateTime, fc, cursor,
		featuresURL, limit, propertyFilters, queryables)

	hf.engine.RenderAndServe(w, r,
		engine.ExpandTemplateKey(featuresKey, hf.engine.CN.NegotiateLanguage(w, r)),
		pageContent, breadcrumbs, outputFormats)
}

func (hf *htmlFeatures) attributes(w http.ResponseWriter, r *http.Request, collection config.FeaturesCollection,
	cursor domain.Cursors, featuresURL featureCollectionURL, limit int, dateTime domain.DateTime,
	propertyFilters map[string]string, queryables datasources.Queryables,
	fc *domain.FeatureCollection, outputFormats []engine.OutputFormat) {

	breadcrumbs, pageContent := hf.toItemsPage(collection, dateTime, fc, cursor,
		featuresURL, limit, propertyFilters, queryables)
	pageContent.ShowViewer = false // since items have no geometry

	hf.engine.RenderAndServe(w, r,
		engine.ExpandTemplateKey(featuresKey, hf.engine.CN.NegotiateLanguage(w, r)),
		pageContent, breadcrumbs, outputFormats)
}

func (hf *htmlFeatures) toItemsPage(collection config.FeaturesCollection, dateTime domain.DateTime,
	fc *domain.FeatureCollection, cursor domain.Cursors, featuresURL featureCollectionURL, limit int,
	propertyFilters map[string]string, queryables datasources.Queryables) ([]engine.Breadcrumb, *featureCollectionPage) {

	breadcrumbs := collectionsBreadcrumb
	breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
		{
			Name: getCollectionTitle(collection.GetID(), collection.GetMetadata()),
			Path: collectionsCrumb + collection.GetID(),
		},
		{
			Name: "Items",
			Path: collectionsCrumb + collection.GetID() + "/items",
		},
	}...)

	var mapSheetProps *config.MapSheetDownloadProperties
	var wc *config.WebConfig
	if collection.MapSheetDownloads != nil {
		mapSheetProps = &collection.MapSheetDownloads.Properties
	}
	wc = collection.Web

	configuredPropertyFilters := make(map[string]datasources.QueryableWithAllowedValues, len(queryables))
	for name, queryable := range queryables {
		if queryable.IsPrimaryGeometry {
			// no need to expose geometry as a property filter (but can be used in CQL)
			continue
		}
		configuredPropertyFilters[name] = queryable
	}

	pageContent := &featureCollectionPage{
		FeatureCollection:         *fc,
		CollectionID:              collection.GetID(),
		Metadata:                  collection.GetMetadata(),
		Cursor:                    cursor,
		PrevLink:                  featuresURL.toPrevNextURL(collection.GetID(), cursor.Prev, engine.FormatHTML),
		NextLink:                  featuresURL.toPrevNextURL(collection.GetID(), cursor.Next, engine.FormatHTML),
		Limit:                     limit,
		DateTime:                  dateTime,
		MapSheetProperties:        mapSheetProps,
		WebConfig:                 wc,
		ShowViewer:                true,
		PropertyFilters:           propertyFilters,
		ConfiguredPropertyFilters: configuredPropertyFilters,
	}

	return breadcrumbs, pageContent
}

func (hf *htmlFeatures) feature(w http.ResponseWriter, r *http.Request,
	collection config.FeaturesCollection, feat *domain.Feature, outputFormats []engine.OutputFormat) {

	breadcrumbs, pageContent := hf.toItemPage(collection, feat)

	hf.engine.RenderAndServe(w, r,
		engine.ExpandTemplateKey(featureKey, hf.engine.CN.NegotiateLanguage(w, r)),
		pageContent, breadcrumbs, outputFormats)
}

func (hf *htmlFeatures) attribute(w http.ResponseWriter, r *http.Request,
	collection config.FeaturesCollection, feat *domain.Feature, outputFormats []engine.OutputFormat) {

	breadcrumbs, pageContent := hf.toItemPage(collection, feat)
	pageContent.ShowViewer = false // since items have no geometry

	hf.engine.RenderAndServe(w, r,
		engine.ExpandTemplateKey(featureKey, hf.engine.CN.NegotiateLanguage(w, r)),
		pageContent, breadcrumbs, outputFormats)
}

func (hf *htmlFeatures) toItemPage(collection config.FeaturesCollection, feat *domain.Feature) ([]engine.Breadcrumb, *featurePage) {
	breadcrumbs := collectionsBreadcrumb
	breadcrumbs = append(breadcrumbs, []engine.Breadcrumb{
		{
			Name: getCollectionTitle(collection.GetID(), collection.GetMetadata()),
			Path: collectionsCrumb + collection.GetID(),
		},
		{
			Name: "Items",
			Path: collectionsCrumb + collection.GetID() + "/items",
		},
		{
			Name: feat.ID,
			Path: collectionsCrumb + collection.GetID() + "/items/" + feat.ID,
		},
	}...)

	var mapSheetProps *config.MapSheetDownloadProperties
	var wc *config.WebConfig

	if collection.MapSheetDownloads != nil {
		mapSheetProps = &collection.MapSheetDownloads.Properties
	}
	wc = collection.Web

	pageContent := &featurePage{
		*feat,
		collection.GetID(),
		feat.ID,
		collection.GetMetadata(),
		mapSheetProps,
		wc,
		true,
	}

	return breadcrumbs, pageContent
}

func getCollectionTitle(collectionID string, metadata *config.GeoSpatialCollectionMetadata) string {
	if metadata != nil && metadata.Title != nil {
		return *metadata.Title
	}

	return collectionID
}

func getCollectionTitleAndDesc(collection config.GeoSpatialCollection) (string, *string) {
	var description *string
	if collection.GetMetadata() != nil {
		description = collection.GetMetadata().Description
	}

	return getCollectionTitle(collection.GetID(), collection.GetMetadata()), description
}
