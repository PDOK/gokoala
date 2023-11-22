package features

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/common/geospatial"
	ds "github.com/PDOK/gokoala/ogc/features/datasources"
	"github.com/PDOK/gokoala/ogc/features/datasources/geopackage"
	"github.com/PDOK/gokoala/ogc/features/datasources/postgis"
	"github.com/go-chi/chi/v5"
)

const (
	templatesDir = "ogc/features/templates/"
	wgs84SRID    = 4326
	wgs84CodeOGC = "CRS84"
	crsURLPrefix = "http://www.opengis.net/def/crs/"
)

var (
	collections map[string]*engine.GeoSpatialCollectionMetadata
)

type DatasourceKey struct {
	srid         int
	collectionID string
}

type Features struct {
	engine      *engine.Engine
	datasources map[DatasourceKey]ds.Datasource

	html *htmlFeatures
	json *jsonFeatures
}

func NewFeatures(e *engine.Engine, router *chi.Mux) *Features {
	f := &Features{
		engine:      e,
		datasources: configureDatasources(e),
		html:        newHTMLFeatures(e),
		json:        newJSONFeatures(e),
	}
	collections = f.cacheCollectionsMetadata()

	router.Get(geospatial.CollectionsPath+"/{collectionId}/items", f.CollectionContent())
	router.Get(geospatial.CollectionsPath+"/{collectionId}/items/{featureId}", f.Feature())
	return f
}

// CollectionContent serve a FeatureCollection with the given collectionId
func (f *Features) CollectionContent(_ ...any) http.HandlerFunc {
	cfg := f.engine.Config

	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")

		url := featureCollectionURL{*cfg.BaseURL.URL, r.URL.Query(), cfg.OgcAPI.Features.Limit}
		encodedCursor, limit, crs, bbox, bboxCrs, err := url.parseParams()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = url.validateNoUnknownParams(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if _, ok := collections[collectionID]; !ok {
			log.Printf("collection %s doesn't exist in this features service", collectionID)
			http.NotFound(w, r)
			return
		}

		datasource := f.datasources[DatasourceKey{srid: crs, collectionID: collectionID}]
		fc, newCursor, err := datasource.GetFeatures(r.Context(), collectionID, ds.FeatureOptions{
			Cursor:  encodedCursor.Decode(url.checksum()),
			Limit:   limit,
			Crs:     crs,
			Bbox:    bbox,
			BboxCrs: bboxCrs,
			// Add filter, filter-crs, etc
		})
		if err != nil {
			// log error, but sent generic message to client to prevent possible information leakage from datasource
			msg := fmt.Sprintf("failed to retrieve feature collection %s", collectionID)
			log.Printf("%s, error: %v\n", msg, err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		if fc == nil {
			log.Printf("no results found for collection '%s' with params: %s",
				collectionID, r.URL.Query().Encode())
			return // still 200 OK
		}

		switch f.engine.CN.NegotiateFormat(r) {
		case engine.FormatHTML:
			f.html.features(w, r, collectionID, newCursor, url, limit, fc)
		case engine.FormatJSON:
			f.json.featuresAsGeoJSON(w, collectionID, newCursor, url, fc)
		case engine.FormatJSONFG:
			f.json.featuresAsJSONFG()
		default:
			http.NotFound(w, r)
			return
		}
	}
}

// Feature serves a single Feature
func (f *Features) Feature() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")
		featureID, err := strconv.Atoi(chi.URLParam(r, "featureId"))
		if err != nil {
			http.Error(w, "feature ID must be a number", http.StatusBadRequest)
			return
		}

		url := featureURL{*f.engine.Config.BaseURL.URL, r.URL.Query()}
		crs, err := url.parseParams()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = url.validateNoUnknownParams(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if _, ok := collections[collectionID]; !ok {
			log.Printf("collection %s doesn't exist in this features service", collectionID)
			http.NotFound(w, r)
			return
		}

		datasource := f.datasources[DatasourceKey{srid: crs, collectionID: collectionID}]
		feat, err := datasource.GetFeature(r.Context(), collectionID, int64(featureID))
		if err != nil {
			// log error, but sent generic message to client to prevent possible information leakage from datasource
			msg := fmt.Sprintf("failed to retrieve feature %d in collection %s", featureID, collectionID)
			log.Printf("%s, error: %v\n", msg, err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		if feat == nil {
			log.Printf("no result found for collection '%s' and feature id: %d",
				collectionID, featureID)
			http.NotFound(w, r)
			return
		}

		switch f.engine.CN.NegotiateFormat(r) {
		case engine.FormatHTML:
			f.html.feature(w, r, collectionID, feat)
		case engine.FormatJSON:
			f.json.featureAsGeoJSON(w, collectionID, feat, url)
		case engine.FormatJSONFG:
			f.json.featureAsJSONFG()
		default:
			http.NotFound(w, r)
			return
		}
	}
}

func (f *Features) cacheCollectionsMetadata() map[string]*engine.GeoSpatialCollectionMetadata {
	result := make(map[string]*engine.GeoSpatialCollectionMetadata)
	for _, collection := range f.engine.Config.OgcAPI.Features.Collections {
		result[collection.ID] = collection.Metadata
	}
	return result
}

func configureDatasources(e *engine.Engine) map[DatasourceKey]ds.Datasource {
	result := make(map[DatasourceKey]ds.Datasource, len(e.Config.OgcAPI.Features.Collections))

	// configure collection specific datasources first
	configureCollectionDatasources(e, result)
	// now configure top-level datasources, for the whole dataset. But only when
	// there's no collection specific datasource already configured
	configureTopLevelDatasources(e, result)

	if len(result) == 0 {
		log.Fatal("no datasource(s) configured for OGC API Features, check config")
	}
	return result
}

func configureTopLevelDatasources(e *engine.Engine, result map[DatasourceKey]ds.Datasource) {
	cfg := e.Config.OgcAPI.Features
	if cfg.Datasources == nil {
		return
	}
	var defaultDS ds.Datasource
	for _, coll := range cfg.Collections {
		key := DatasourceKey{srid: wgs84SRID, collectionID: coll.ID}
		if result[key] == nil {
			if defaultDS == nil {
				defaultDS = newDatasource(e, cfg.Collections, cfg.Datasources.DefaultWGS84)
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
				result[key] = newDatasource(e, cfg.Collections, additional.Datasource)
			}
		}
	}
}

func configureCollectionDatasources(e *engine.Engine, result map[DatasourceKey]ds.Datasource) {
	cfg := e.Config.OgcAPI.Features
	for _, coll := range cfg.Collections {
		if coll.Features == nil || coll.Features.Datasources == nil {
			continue
		}
		defaultDS := newDatasource(e, cfg.Collections, coll.Features.Datasources.DefaultWGS84)
		result[DatasourceKey{srid: wgs84SRID, collectionID: coll.ID}] = defaultDS

		for _, additional := range coll.Features.Datasources.Additional {
			srid, err := epsgToSrid(additional.Srs)
			if err != nil {
				log.Fatal(err)
			}
			additionalDS := newDatasource(e, cfg.Collections, additional.Datasource)
			result[DatasourceKey{srid: srid, collectionID: coll.ID}] = additionalDS
		}
	}
}

func newDatasource(e *engine.Engine, coll engine.GeoSpatialCollections, dsConfig engine.Datasource) ds.Datasource {
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
