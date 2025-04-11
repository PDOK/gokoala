package features

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/google/uuid"
	"github.com/twpayne/go-geom"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/geopackage"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/postgis"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/go-chi/chi/v5"
)

const (
	templatesDir = "internal/ogc/features/templates/"
)

var (
	configuredCollections  map[string]config.GeoSpatialCollection
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
	datasources := createDatasources(e)
	configuredCollections = cacheConfiguredFeatureCollections(e)
	configuredPropertyFilters := configurePropertyFiltersWithAllowedValues(datasources, configuredCollections)

	rebuildOpenAPIForFeatures(e, datasources, configuredPropertyFilters)

	f := &Features{
		engine:                    e,
		datasources:               datasources,
		configuredPropertyFilters: configuredPropertyFilters,
		defaultProfile:            domain.NewProfile(domain.RelAsLink, *e.Config.BaseURL.URL, util.Keys(configuredCollections)),
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
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f.engine.OpenAPI.ValidateRequest(r); err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}

		collectionID := chi.URLParam(r, "collectionId")
		collection, ok := configuredCollections[collectionID]
		if !ok {
			handleCollectionNotFound(w, collectionID)
			return
		}
		url, encodedCursor, limit, inputSRID, outputSRID, contentCrs, bbox,
			referenceDate, propertyFilters, err := f.parseFeaturesURL(r, collection)
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
				propertyFilters, f.configuredPropertyFilters[collectionID], collection.Features, fc)
		case engine.FormatGeoJSON, engine.FormatJSON:
			f.json.featuresAsGeoJSON(w, r, collectionID, newCursor, url, collection.Features, fc)
		case engine.FormatJSONFG:
			f.json.featuresAsJSONFG(w, r, collectionID, newCursor, url, collection.Features, fc, contentCrs)
		default:
			engine.RenderProblem(engine.ProblemNotAcceptable, w, fmt.Sprintf("format '%s' is not supported", format))
			return
		}
	}
}

// Feature serves a single Feature
func (f *Features) Feature() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f.engine.OpenAPI.ValidateRequest(r); err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}

		collectionID := chi.URLParam(r, "collectionId")
		collection, ok := configuredCollections[collectionID]
		if !ok {
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
			f.html.feature(w, r, collectionID, collection.Features, feat)
		case engine.FormatGeoJSON, engine.FormatJSON:
			f.json.featureAsGeoJSON(w, r, collectionID, collection.Features, feat, url)
		case engine.FormatJSONFG:
			f.json.featureAsJSONFG(w, r, collectionID, collection.Features, feat, url, contentCrs)
		default:
			engine.RenderProblem(engine.ProblemNotAcceptable, w, fmt.Sprintf("format '%s' is not supported", format))
			return
		}
	}
}

func (f *Features) parseFeaturesURL(r *http.Request, collection config.GeoSpatialCollection) (featureCollectionURL,
	domain.EncodedCursor, int, domain.SRID, domain.SRID, domain.ContentCrs, *geom.Bounds, time.Time, map[string]string, error) {

	url := featureCollectionURL{
		*f.engine.Config.BaseURL.URL,
		r.URL.Query(),
		f.engine.Config.OgcAPI.Features.Limit,
		f.configuredPropertyFilters[collection.ID],
		collection.HasDateTime(),
	}
	encodedCursor, limit, inputSRID, outputSRID, contentCrs, bbox, referenceDate, propertyFilters, err := url.parse()
	return url, encodedCursor, limit, inputSRID, outputSRID, contentCrs, bbox, referenceDate, propertyFilters, err
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

func cacheConfiguredFeatureCollections(e *engine.Engine) map[string]config.GeoSpatialCollection {
	result := make(map[string]config.GeoSpatialCollection)
	for _, collection := range e.Config.OgcAPI.Features.Collections {
		result[collection.ID] = collection
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

func configurePropertyFiltersWithAllowedValues(datasources map[DatasourceKey]ds.Datasource,
	collections map[string]config.GeoSpatialCollection) map[string]ds.PropertyFiltersWithAllowedValues {

	result := make(map[string]ds.PropertyFiltersWithAllowedValues)
	for k, datasource := range datasources {
		result[k.collectionID] = datasource.GetPropertyFiltersWithAllowedValues(k.collectionID)
	}

	// sanity check to make sure datasources return all configured property filters.
	for _, collection := range collections {
		actual := len(result[collection.ID])
		if collection.Features != nil && collection.Features.Filters.Properties != nil {
			expected := len(collection.Features.Filters.Properties)
			if expected != actual {
				log.Fatalf("number of property filters received from datasource for collection '%s' does not "+
					"match the number of configured property filters. Expected filters: %d, got from datasource: %d",
					collection.ID, expected, actual)
			}
		}
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

func querySingleDatasource(input domain.SRID, output domain.SRID, bbox *geom.Bounds) bool {
	return bbox == nil ||
		int(input) == int(output) ||
		(int(input) == domain.UndefinedSRID && int(output) == domain.WGS84SRID) ||
		(int(input) == domain.WGS84SRID && int(output) == domain.UndefinedSRID)
}

func getTemporalCriteria(collection config.GeoSpatialCollection, referenceDate time.Time) ds.TemporalCriteria {
	var temporalCriteria ds.TemporalCriteria
	if collection.HasDateTime() {
		temporalCriteria = ds.TemporalCriteria{
			ReferenceDate:     referenceDate,
			StartDateProperty: collection.Metadata.TemporalProperties.StartDate,
			EndDateProperty:   collection.Metadata.TemporalProperties.EndDate}
	}
	return temporalCriteria
}
