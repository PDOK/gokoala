package features

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/google/uuid"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/geopackage"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/postgis"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-spatial/geom"
)

const (
	templatesDir = "internal/ogc/features/templates/"
)

var (
	collections            map[string]*config.GeoSpatialCollectionMetadata
	emptyFeatureCollection = &domain.FeatureCollection{Features: make([]*domain.Feature, 0)}
)

type DatasourceKey struct {
	srid         int
	collectionID string
}

type DatasourceConfig struct {
	collections config.GeoSpatialCollections
	ds          config.Datasource
}

type Features struct {
	engine                    *engine.Engine
	datasources               map[DatasourceKey]ds.Datasource
	configuredPropertyFilters map[string]ds.PropertyFiltersWithAllowedValues
	defaultProfile            domain.Profile

	html *htmlFeatures
	json *jsonFeatures
}

func NewFeatures(e *engine.Engine) *Features {
	collections = cacheCollectionsMetadata(e)
	datasources := createDatasources(e)
	configuredPropertyFilters := configurePropertyFiltersWithAllowedValues(datasources)

	rebuildOpenAPIForFeatures(e, datasources, configuredPropertyFilters)

	f := &Features{
		engine:                    e,
		datasources:               datasources,
		configuredPropertyFilters: configuredPropertyFilters,
		defaultProfile:            domain.NewProfile(domain.RelAsLink, *e.Config.BaseURL.URL),
		html:                      newHTMLFeatures(e),
		json:                      newJSONFeatures(e),
	}

	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/items", f.Features())
	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/items/{featureId}", f.Feature())
	return f
}

// Features serve a FeatureCollection with the given collectionId
//
// Beware: this is one of the most performance sensitive pieces of code in the system.
// Try to do as much initialization work outside the hot path, and only do essential
// operations inside this method.
func (f *Features) Features() http.HandlerFunc {
	cfg := f.engine.Config

	return func(w http.ResponseWriter, r *http.Request) {
		if err := f.engine.OpenAPI.ValidateRequest(r); err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}

		collectionID := chi.URLParam(r, "collectionId")
		collection, ok := collections[collectionID]
		if !ok {
			handleCollectionNotFound(w, collectionID)
			return
		}
		url := featureCollectionURL{*cfg.BaseURL.URL, r.URL.Query(), cfg.OgcAPI.Features.Limit,
			cfg.OgcAPI.Features.PropertyFiltersForCollection(collectionID), hasDateTime(collection)}
		encodedCursor, limit, inputSRID, outputSRID, contentCrs, bbox, referenceDate, propertyFilters, err := url.parse()
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}
		w.Header().Add(engine.HeaderContentCrs, contentCrs.ToLink())

		var newCursor domain.Cursors
		var fc *domain.FeatureCollection
		if querySingleDatasource(inputSRID, outputSRID, bbox) {
			// fast path
			datasource := f.datasources[DatasourceKey{srid: outputSRID.GetOrDefault(), collectionID: collectionID}]
			fc, newCursor, err = datasource.GetFeatures(r.Context(), collectionID, ds.FeaturesCriteria{
				Cursor:           encodedCursor.Decode(url.checksum()),
				Limit:            limit,
				InputSRID:        inputSRID.GetOrDefault(),
				OutputSRID:       outputSRID.GetOrDefault(),
				Bbox:             bbox,
				TemporalCriteria: getTemporalCriteria(collection, referenceDate),
				PropertyFilters:  propertyFilters,
				// Add filter, filter-lang
			}, f.defaultProfile)
			if err != nil {
				handleFeaturesQueryError(w, collectionID, err)
				return
			}
		} else {
			// slower path: get feature ids by input CRS (step 1), then the actual features in output CRS (step 2)
			var fids []int64
			datasource := f.datasources[DatasourceKey{srid: inputSRID.GetOrDefault(), collectionID: collectionID}]
			fids, newCursor, err = datasource.GetFeatureIDs(r.Context(), collectionID, ds.FeaturesCriteria{
				Cursor:           encodedCursor.Decode(url.checksum()),
				Limit:            limit,
				InputSRID:        inputSRID.GetOrDefault(),
				OutputSRID:       outputSRID.GetOrDefault(),
				Bbox:             bbox,
				TemporalCriteria: getTemporalCriteria(collection, referenceDate),
				PropertyFilters:  propertyFilters,
				// Add filter, filter-lang
			})
			if err == nil && fids != nil {
				datasource = f.datasources[DatasourceKey{srid: outputSRID.GetOrDefault(), collectionID: collectionID}]
				fc, err = datasource.GetFeaturesByID(r.Context(), collectionID, fids, f.defaultProfile)
			}
			if err != nil {
				handleFeaturesQueryError(w, collectionID, err)
				return
			}
		}
		if fc == nil {
			fc = emptyFeatureCollection
		}

		format := f.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatHTML:
			f.html.features(w, r, collectionID, newCursor, url, limit, &referenceDate,
				propertyFilters, f.configuredPropertyFilters[collectionID],
				cfg.OgcAPI.Features.MapSheetPropertiesForCollection(collectionID), fc)
		case engine.FormatGeoJSON, engine.FormatJSON:
			f.json.featuresAsGeoJSON(w, r, collectionID, newCursor, url, fc)
		case engine.FormatJSONFG:
			f.json.featuresAsJSONFG(w, r, collectionID, newCursor, url, fc, contentCrs)
		default:
			engine.RenderProblem(engine.ProblemNotAcceptable, w, fmt.Sprintf("format '%s' is not supported", format))
			return
		}
	}
}

// Feature serves a single Feature
func (f *Features) Feature() http.HandlerFunc {
	cfg := f.engine.Config

	return func(w http.ResponseWriter, r *http.Request) {
		if err := f.engine.OpenAPI.ValidateRequest(r); err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}

		collectionID := chi.URLParam(r, "collectionId")
		if _, ok := collections[collectionID]; !ok {
			handleCollectionNotFound(w, collectionID)
			return
		}
		featureID, err := parseFeatureID(r)
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}
		url := featureURL{*f.engine.Config.BaseURL.URL, r.URL.Query()}
		outputSRID, contentCrs, err := url.parse()
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}
		w.Header().Add(engine.HeaderContentCrs, contentCrs.ToLink())
		mapSheetProperties := cfg.OgcAPI.Features.MapSheetPropertiesForCollection(collectionID)

		datasource := f.datasources[DatasourceKey{srid: outputSRID.GetOrDefault(), collectionID: collectionID}]
		feat, err := datasource.GetFeature(r.Context(), collectionID, featureID, f.defaultProfile)
		if err != nil {
			handleFeatureQueryError(w, collectionID, featureID, err)
			return
		}
		if feat == nil {
			handleFeatureNotFound(w, collectionID, featureID)
			return
		}

		format := f.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatHTML:
			f.html.feature(w, r, collectionID, mapSheetProperties, feat)
		case engine.FormatGeoJSON, engine.FormatJSON:
			f.json.featureAsGeoJSON(w, r, collectionID, feat, url)
		case engine.FormatJSONFG:
			f.json.featureAsJSONFG(w, r, collectionID, feat, url, contentCrs)
		default:
			engine.RenderProblem(engine.ProblemNotAcceptable, w, fmt.Sprintf("format '%s' is not supported", format))
			return
		}
	}
}

func parseFeatureID(r *http.Request) (any, error) {
	var featureID any
	featureID, err := uuid.Parse(chi.URLParam(r, "featureId"))
	if err != nil {
		// fallback to numerical feature id
		featureID, err = strconv.ParseInt(chi.URLParam(r, "featureId"), 10, 0)
		if err != nil {
			return nil, errors.New("feature ID must be a UUID or number")
		}
	}
	return featureID, nil
}

func cacheCollectionsMetadata(e *engine.Engine) map[string]*config.GeoSpatialCollectionMetadata {
	result := make(map[string]*config.GeoSpatialCollectionMetadata)
	for _, collection := range e.Config.OgcAPI.Features.Collections {
		result[collection.ID] = collection.Metadata
	}
	return result
}

func createDatasources(e *engine.Engine) map[DatasourceKey]ds.Datasource {
	configured := make(map[DatasourceKey]*DatasourceConfig, len(e.Config.OgcAPI.Features.Collections))

	// configure collection specific datasources first
	configureCollectionDatasources(e, configured)
	// now configure top-level datasources, for the whole dataset. But only when
	// there's no collection specific datasource already configured
	configureTopLevelDatasources(e, configured)

	if len(configured) == 0 {
		log.Fatal("no datasource(s) configured for OGC API Features, check config")
	}

	// now we have a mapping from collection+projection => desired datasource (the 'configured' map).
	// but the actual datasource connection still needs to be CREATED and associated with these collections.
	// this is what we'll going to do now, but in the process we need to make sure no duplicate datasources
	// are instantiated, since multiple collection can point to the same datasource and we only what to have a single
	// datasource/connection-pool serving those collections.
	createdDatasources := make(map[config.Datasource]ds.Datasource)
	result := make(map[DatasourceKey]ds.Datasource, len(configured))
	for k, cfg := range configured {
		if cfg == nil {
			continue
		}
		existing, ok := createdDatasources[cfg.ds]
		if !ok {
			// make sure to only create a new datasource when it hasn't already been done before (for another collection)
			created := newDatasource(e, cfg.collections, cfg.ds)
			createdDatasources[cfg.ds] = created
			result[k] = created
		} else {
			result[k] = existing
		}
	}
	return result
}

func configurePropertyFiltersWithAllowedValues(datasources map[DatasourceKey]ds.Datasource) map[string]ds.PropertyFiltersWithAllowedValues {
	result := make(map[string]ds.PropertyFiltersWithAllowedValues)
	for k, datasource := range datasources {
		result[k.collectionID] = datasource.GetPropertyFiltersWithAllowedValues(k.collectionID)
	}
	return result
}

func configureTopLevelDatasources(e *engine.Engine, result map[DatasourceKey]*DatasourceConfig) {
	cfg := e.Config.OgcAPI.Features
	if cfg.Datasources == nil {
		return
	}
	var defaultDS *DatasourceConfig
	for _, coll := range cfg.Collections {
		key := DatasourceKey{srid: domain.WGS84SRID, collectionID: coll.ID}
		if result[key] == nil {
			if defaultDS == nil {
				defaultDS = &DatasourceConfig{cfg.Collections, cfg.Datasources.DefaultWGS84}
			}
			result[key] = defaultDS
		}
	}

	for _, additional := range cfg.Datasources.Additional {
		for _, coll := range cfg.Collections {
			srid, err := domain.EpsgToSrid(additional.Srs)
			if err != nil {
				log.Fatal(err)
			}
			key := DatasourceKey{srid: srid.GetOrDefault(), collectionID: coll.ID}
			if result[key] == nil {
				result[key] = &DatasourceConfig{cfg.Collections, additional.Datasource}
			}
		}
	}
}

func configureCollectionDatasources(e *engine.Engine, result map[DatasourceKey]*DatasourceConfig) {
	cfg := e.Config.OgcAPI.Features
	for _, coll := range cfg.Collections {
		if coll.Features == nil || coll.Features.Datasources == nil {
			continue
		}
		defaultDS := &DatasourceConfig{cfg.Collections, coll.Features.Datasources.DefaultWGS84}
		result[DatasourceKey{srid: domain.WGS84SRID, collectionID: coll.ID}] = defaultDS

		for _, additional := range coll.Features.Datasources.Additional {
			srid, err := domain.EpsgToSrid(additional.Srs)
			if err != nil {
				log.Fatal(err)
			}
			additionalDS := &DatasourceConfig{cfg.Collections, additional.Datasource}
			result[DatasourceKey{srid: srid.GetOrDefault(), collectionID: coll.ID}] = additionalDS
		}
	}
}

func newDatasource(e *engine.Engine, coll config.GeoSpatialCollections, dsConfig config.Datasource) ds.Datasource {
	var datasource ds.Datasource
	if dsConfig.GeoPackage != nil {
		datasource = geopackage.NewGeoPackage(coll, *dsConfig.GeoPackage)
	} else if dsConfig.PostGIS != nil {
		datasource = postgis.NewPostGIS()
	}
	e.RegisterShutdownHook(datasource.Close)
	return datasource
}

func handleCollectionNotFound(w http.ResponseWriter, collectionID string) {
	msg := fmt.Sprintf("collection %s doesn't exist in this features service", collectionID)
	log.Println(msg)
	engine.RenderProblem(engine.ProblemNotFound, w, msg)
}

func handleFeatureNotFound(w http.ResponseWriter, collectionID string, featureID any) {
	msg := fmt.Sprintf("the requested feature with id: %v does not exist in collection '%v'", featureID, collectionID)
	log.Println(msg)
	engine.RenderProblem(engine.ProblemNotFound, w, msg)
}

// log error, but send generic message to client to prevent possible information leakage from datasource
func handleFeaturesQueryError(w http.ResponseWriter, collectionID string, err error) {
	msg := "failed to retrieve feature collection " + collectionID
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		// provide more context when user hits the query timeout
		msg += ": querying the features took too long (timeout encountered). Simplify your request and try again, or contact support"
	}
	log.Printf("%s, error: %v\n", msg, err)
	engine.RenderProblem(engine.ProblemServerError, w, msg) // don't include sensitive information in details msg
}

// log error, but sent generic message to client to prevent possible information leakage from datasource
func handleFeatureQueryError(w http.ResponseWriter, collectionID string, featureID any, err error) {
	msg := fmt.Sprintf("failed to retrieve feature %v in collection %s", featureID, collectionID)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		// provide more context when user hits the query timeout
		msg += ": querying the feature took too long (timeout encountered). Try again, or contact support"
	}
	log.Printf("%s, error: %v\n", msg, err)
	engine.RenderProblem(engine.ProblemServerError, w, msg) // don't include sensitive information in details msg
}

func querySingleDatasource(input domain.SRID, output domain.SRID, bbox *geom.Extent) bool {
	return bbox == nil ||
		int(input) == int(output) ||
		(int(input) == domain.UndefinedSRID && int(output) == domain.WGS84SRID) ||
		(int(input) == domain.WGS84SRID && int(output) == domain.UndefinedSRID)
}

func getTemporalCriteria(collection *config.GeoSpatialCollectionMetadata, referenceDate time.Time) ds.TemporalCriteria {
	var temporalCriteria ds.TemporalCriteria
	if hasDateTime(collection) {
		temporalCriteria = ds.TemporalCriteria{
			ReferenceDate:     referenceDate,
			StartDateProperty: collection.TemporalProperties.StartDate,
			EndDateProperty:   collection.TemporalProperties.EndDate}
	}
	return temporalCriteria
}

func hasDateTime(collection *config.GeoSpatialCollectionMetadata) bool {
	return collection != nil && collection.TemporalProperties != nil
}
