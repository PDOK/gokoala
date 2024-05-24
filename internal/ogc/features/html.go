package features

import (
	"net/http"
	"strconv"
	"time"

	"github.com/PDOK/gokoala/config"

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

	CollectionID         string
	Metadata             *config.GeoSpatialCollectionMetadata
	Cursor               domain.Cursors
	PrevLink             string
	NextLink             string
	Limit                int
	ReferenceDate        *time.Time
	PropertyFilters      map[string]string
	MapSheetProperties   *config.MapSheetDownloadProperties
	DownloadPeriodValues []string
}

// featurePage enriched Feature for HTML representation.
type featurePage struct {
	domain.Feature

	CollectionID       string
	FeatureID          int64
	Metadata           *config.GeoSpatialCollectionMetadata
	MapSheetProperties *config.MapSheetDownloadProperties
}

func (hf *htmlFeatures) features(w http.ResponseWriter, r *http.Request, collectionID string,
	cursor domain.Cursors, featuresURL featureCollectionURL, limit int, referenceDate *time.Time,
	propertyFilters map[string]string, mapSheetProperties *config.MapSheetDownloadProperties, downloadPeriod DownloadPeriodConfig, fc *domain.FeatureCollection) {

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
		propertyFilters,
		mapSheetProperties,
		downloadPeriod.paramValues,
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
			Name: strconv.FormatInt(feat.ID, 10),
			Path: collectionsCrumb + collectionID + "/items/" + strconv.FormatInt(feat.ID, 10),
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
