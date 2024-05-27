package features

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/config"

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
	templatesDir  = "internal/ogc/features/templates/"
	crsURIPrefix  = "http://www.opengis.net/def/crs/"
	undefinedSRID = 0
	wgs84SRID     = 100000 // We use the SRID for CRS84 (WGS84) as defined in the GeoPackage, instead of EPSG:4326 (due to axis order). In time, we may need to read this value dynamically from the GeoPackage.
	wgs84CodeOGC  = "CRS84"
	wgs84CrsURI   = crsURIPrefix + "OGC/1.3/" + wgs84CodeOGC
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

type DownloadPeriodConfig struct {
	paramName   string
	paramValues []string
}

type Features struct {
	engine          *engine.Engine
	datasources     map[DatasourceKey]ds.Datasource
	downloadPeriods map[string]DownloadPeriodConfig

	html *htmlFeatures
	json *jsonFeatures
}

func NewFeatures(e *engine.Engine) *Features {
	collections = cacheCollectionsMetadata(e)
	datasources := createDatasources(e)
	downloadPeriods := createDownloadPeriods(e, datasources)

	rebuildOpenAPIForFeatures(e, datasources)

	f := &Features{
		engine:          e,
		datasources:     datasources,
		downloadPeriods: downloadPeriods,
		html:            newHTMLFeatures(e),
		json:            newJSONFeatures(e),
	}

	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/items", f.Features())
	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/items/{featureId}", f.Feature())
	return f
}

// Features serve a FeatureCollection with the given collectionId
//
//nolint:cyclop
func (f *Features) Features() http.HandlerFunc {
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
		downloadPeriod := f.downloadPeriods[collectionID]
		mapSheetProperties := cfg.OgcAPI.Features.MapSheetPropertiesForCollection(collectionID)
		url := featureCollectionURL{*cfg.BaseURL.URL, r.URL.Query(), cfg.OgcAPI.Features.Limit,
			cfg.OgcAPI.Features.PropertyFiltersForCollection(collectionID), false, downloadPeriod.paramName}
		if collection := collections[collectionID]; collection != nil && collection.TemporalProperties != nil {
			url.supportsDatetime = true
		}
		encodedCursor, limit, inputSRID, outputSRID, contentCrs, bbox, referenceDate, propertyFilters, err := url.parse()
		var temporalCriteria ds.TemporalCriteria
		if collection := collections[collectionID]; collection != nil && collection.TemporalProperties != nil {
			temporalCriteria = ds.TemporalCriteria{
				ReferenceDate:     referenceDate,
				StartDateProperty: collection.TemporalProperties.StartDate,
				EndDateProperty:   collection.TemporalProperties.EndDate}
		}
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
				TemporalCriteria: temporalCriteria,
				PropertyFilters:  propertyFilters,
				// Add filter, filter-lang
			})
			if err != nil {
				handleFeatureCollectionError(w, collectionID, err)
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
				TemporalCriteria: temporalCriteria,
				PropertyFilters:  propertyFilters,
				// Add filter, filter-lang
			})
			if err == nil && fids != nil {
				datasource = f.datasources[DatasourceKey{srid: outputSRID.GetOrDefault(), collectionID: collectionID}]
				fc, err = datasource.GetFeaturesByID(r.Context(), collectionID, fids)
			}
			if err != nil {
				handleFeatureCollectionError(w, collectionID, err)
				return
			}
		}
		if fc == nil {
			fc = emptyFeatureCollection
		}

		format := f.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatHTML:
			f.html.features(w, r, collectionID, newCursor, url, limit, &referenceDate, propertyFilters, mapSheetProperties, downloadPeriod, fc)
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
		featureID, err := strconv.Atoi(chi.URLParam(r, "featureId"))
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, "feature ID must be a number")
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
		feat, err := datasource.GetFeature(r.Context(), collectionID, int64(featureID))
		if err != nil {
			// log error, but sent generic message to client to prevent possible information leakage from datasource
			msg := fmt.Sprintf("failed to retrieve feature %d in collection %s", featureID, collectionID)
			log.Printf("%s, error: %v\n", msg, err)
			engine.RenderProblem(engine.ProblemServerError, w, msg)
			return
		}
		if feat == nil {
			msg := fmt.Sprintf("the requested feature with id: %d does not exist in collection '%s'", featureID, collectionID)
			log.Println(msg)
			engine.RenderProblem(engine.ProblemNotFound, w, msg)
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

	created := make(map[DatasourceKey]ds.Datasource, len(configured))
	for k, cfg := range configured {
		if cfg == nil {
			continue
		}
		created[k] = newDatasource(e, cfg.collections, cfg.ds)
	}
	return created
}

func createDownloadPeriods(e *engine.Engine, datasources map[DatasourceKey]ds.Datasource) map[string]DownloadPeriodConfig {
	result := make(map[string]DownloadPeriodConfig)
	for k, datasource := range datasources {
		mapSheetConfig := e.Config.OgcAPI.Features.MapSheetPropertiesForCollection(k.collectionID)
		if mapSheetConfig == nil || mapSheetConfig.Temporal == nil {
			continue
		}
		downloadPeriods, err := datasource.GetDownloadPeriods(k.collectionID)
		if err != nil {
			continue
		}
		result[k.collectionID] = DownloadPeriodConfig{mapSheetConfig.Temporal.Name, downloadPeriods}
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
		key := DatasourceKey{srid: wgs84SRID, collectionID: coll.ID}
		if result[key] == nil {
			if defaultDS == nil {
				defaultDS = &DatasourceConfig{cfg.Collections, cfg.Datasources.DefaultWGS84}
			}
			result[key] = defaultDS
		}
	}

	for _, additional := range cfg.Datasources.Additional {
		for _, coll := range cfg.Collections {
			srid, err := epsgToSrid(additional.Srs)
			if err != nil {
				log.Fatal(err)
			}
			key := DatasourceKey{srid: srid, collectionID: coll.ID}
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
		result[DatasourceKey{srid: wgs84SRID, collectionID: coll.ID}] = defaultDS

		for _, additional := range coll.Features.Datasources.Additional {
			srid, err := epsgToSrid(additional.Srs)
			if err != nil {
				log.Fatal(err)
			}
			additionalDS := &DatasourceConfig{cfg.Collections, additional.Datasource}
			result[DatasourceKey{srid: srid, collectionID: coll.ID}] = additionalDS
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

func epsgToSrid(srs string) (int, error) {
	prefix := "EPSG:"
	srsCode, found := strings.CutPrefix(srs, prefix)
	if !found {
		return -1, fmt.Errorf("expected configured SRS to start with '%s', got %s", prefix, srs)
	}
	srid, err := strconv.Atoi(srsCode)
	if err != nil {
		return -1, fmt.Errorf("expected EPSG code to have numeric value, got %s", srsCode)
	}
	return srid, nil
}

func handleCollectionNotFound(w http.ResponseWriter, collectionID string) {
	msg := fmt.Sprintf("collection %s doesn't exist in this features service", collectionID)
	log.Println(msg)
	engine.RenderProblem(engine.ProblemNotFound, w, msg)
}

// log error, but send generic message to client to prevent possible information leakage from datasource
func handleFeatureCollectionError(w http.ResponseWriter, collectionID string, err error) {
	msg := "failed to retrieve feature collection " + collectionID
	log.Printf("%s, error: %v\n", msg, err)
	engine.RenderProblem(engine.ProblemServerError, w, msg)
}

func querySingleDatasource(input SRID, output SRID, bbox *geom.Extent) bool {
	return bbox == nil ||
		int(input) == int(output) ||
		(int(input) == undefinedSRID && int(output) == wgs84SRID) ||
		(int(input) == wgs84SRID && int(output) == undefinedSRID)
}
