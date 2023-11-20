package features

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/common/geospatial"
	ds "github.com/PDOK/gokoala/ogc/features/datasources"
	"github.com/PDOK/gokoala/ogc/features/datasources/geopackage"
	"github.com/PDOK/gokoala/ogc/features/datasources/postgis"
	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-spatial/geom"
)

const (
	templatesDir = "ogc/features/templates/"
	wgs84SRID    = 4326
	wgs84CodeOGC = "CRS84"
	crsURLPrefix = "http://www.opengis.net/def/crs/"
)

type DataSourceKey struct {
	srid         int
	collectionID string
}

var (
	collections map[string]*engine.GeoSpatialCollectionMetadata
)

type Features struct {
	engine      *engine.Engine
	datasources map[DataSourceKey]ds.Datasource

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
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID, encodedCursor, limit, crs, bbox, bboxCrs, err := f.parseFeatureCollectionRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		url := featureCollectionURL{*f.engine.Config.BaseURL.URL, r.URL.Query()}
		if err = url.validateNoUnknownParams(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if _, ok := collections[collectionID]; !ok {
			log.Printf("collection %s doesn't exist in this features service", collectionID)
			http.NotFound(w, r)
			return
		}

		datasource := f.datasources[DataSourceKey{srid: crs, collectionID: collectionID}]
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
		crs, err := f.parseSRID(r.URL.Query(), crsParam)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		url := featureURL{*f.engine.Config.BaseURL.URL, r.URL.Query()}
		if err = url.validateNoUnknownParams(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if _, ok := collections[collectionID]; !ok {
			log.Printf("collection %s doesn't exist in this features service", collectionID)
			http.NotFound(w, r)
			return
		}

		datasource := f.datasources[DataSourceKey{srid: crs, collectionID: collectionID}]
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

func (f *Features) parseFeatureCollectionRequest(r *http.Request) (string, domain.EncodedCursor, int, int, *geom.Extent, int, error) {
	collectionID := chi.URLParam(r, "collectionId")
	encodedCursor := domain.EncodedCursor(r.URL.Query().Get(cursorParam))
	limit, limitErr := f.parseLimit(r.URL.Query())
	crs, crsErr := f.parseSRID(r.URL.Query(), crsParam)
	bbox, bboxCrs, bboxErr := f.parseBbox(r.URL.Query())
	dateTimeErr := f.parseDateTime(r.URL.Query())
	filterErr := f.parseFilter(r.URL.Query())

	err := errors.Join(limitErr, crsErr, bboxErr, dateTimeErr, filterErr)
	return collectionID, encodedCursor, limit, crs, bbox, bboxCrs, err
}

func (f *Features) parseLimit(params neturl.Values) (int, error) {
	limit := f.engine.Config.OgcAPI.Features.Limit.Default
	var err error
	if params.Get(limitParam) != "" {
		limit, err = strconv.Atoi(params.Get(limitParam))
		if err != nil {
			err = fmt.Errorf("limit must be numeric")
		}
		// OpenAPI validation already guards against exceeding max limit, this is just a defense in-depth measure.
		if limit > f.engine.Config.OgcAPI.Features.Limit.Max {
			limit = f.engine.Config.OgcAPI.Features.Limit.Max
		}
	}
	if limit < 0 {
		err = fmt.Errorf("limit can't be negative")
	}
	return limit, err
}

func (f *Features) parseBbox(params neturl.Values) (*geom.Extent, int, error) {
	bboxCrs, err := f.parseSRID(params, bboxCrsParam)
	if err != nil {
		return nil, -1, err
	}

	if params.Get(bboxParam) == "" {
		return nil, bboxCrs, nil
	}
	bboxValues := strings.Split(params.Get(bboxParam), ",")
	if len(bboxValues) != 4 {
		return nil, bboxCrs, fmt.Errorf("bbox should contain exactly 4 values " +
			"separated by commas: minx,miny,maxx,maxy")
	}

	var extent geom.Extent
	for i, v := range bboxValues {
		extent[i], err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, bboxCrs, fmt.Errorf("failed to parse value %s in bbox, error: %w", v, err)
		}
	}

	return &extent, bboxCrs, nil
}

func (f *Features) parseSRID(params neturl.Values, paramName string) (int, error) {
	srid := wgs84SRID
	param := params.Get(paramName)
	if param == "" {
		return srid, nil
	}
	param = strings.TrimSpace(param)
	if !strings.HasPrefix(param, crsURLPrefix) {
		return srid, fmt.Errorf("%s param should start with %s, got: %s", paramName, crsURLPrefix, param)
	}
	lastIndex := strings.LastIndex(param, "/")
	if lastIndex != -1 {
		crsCode := param[lastIndex+1:]
		if crsCode == wgs84CodeOGC {
			return srid, nil // CRS84 is WGS84, just like EPSG:4326 (only axis order differs but SRID is the same)
		}
		var err error
		srid, err = strconv.Atoi(crsCode)
		if err != nil {
			return 0, fmt.Errorf("expected numerical CRS code, received: %s", crsCode)
		}
	}
	return srid, nil
}

func (f *Features) parseDateTime(params neturl.Values) error {
	if params.Get(dateTimeParam) != "" {
		return fmt.Errorf("datetime param is currently not supported")
	}
	return nil
}

func (f *Features) parseFilter(params neturl.Values) error {
	if params.Get(filterParam) != "" {
		return fmt.Errorf("CQL filter param is currently not supported")
	}
	if params.Get(filterCrsParam) != "" {
		return fmt.Errorf("CQL filter-crs param is currently not supported")
	}
	return nil
}

func epsgToSrid(srs string) (int, error) {
	srsCode, found := strings.CutPrefix(srs, "EPSG:")
	if !found {
		return -1, fmt.Errorf("expected configured SRS to start with EPSG, got %s", srs)
	}
	srid, err := strconv.Atoi(srsCode)
	if err != nil {
		return -1, fmt.Errorf("expected EPSG code to have numeric value, got %s", srsCode)
	}
	return srid, nil
}

func configureDatasources(e *engine.Engine) map[DataSourceKey]ds.Datasource {
	cfg := e.Config.OgcAPI.Features
	result := make(map[DataSourceKey]ds.Datasource, len(cfg.Collections))

	if cfg.Datasources != nil {
		defaultDS := newDatasource(e, cfg.Collections, cfg.Datasources.DefaultWGS84)
		for _, coll := range cfg.Collections {
			result[DataSourceKey{srid: wgs84SRID, collectionID: coll.ID}] = defaultDS
		}

		for _, additional := range cfg.Datasources.Additional {
			additionalDS := newDatasource(e, cfg.Collections, additional.Datasource)
			for _, coll := range cfg.Collections {
				srid, err := epsgToSrid(additional.Srs)
				if err != nil {
					log.Fatal(err)
				}
				result[DataSourceKey{srid: srid, collectionID: coll.ID}] = additionalDS
			}
		}
	} else {
		for _, coll := range cfg.Collections {
			defaultDS := newDatasource(e, cfg.Collections, coll.Features.Datasources.DefaultWGS84)
			result[DataSourceKey{srid: wgs84SRID, collectionID: coll.ID}] = defaultDS

			for _, additional := range coll.Features.Datasources.Additional {
				additionalDS := newDatasource(e, cfg.Collections, additional.Datasource)
				srid, err := epsgToSrid(additional.Srs)
				if err != nil {
					log.Fatal(err)
				}
				result[DataSourceKey{srid: srid, collectionID: coll.ID}] = additionalDS
			}
		}
	}
	if len(result) == 0 {
		log.Fatal("no datasource(s) configured for OGC API Features, check config")
	}
	return result
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
