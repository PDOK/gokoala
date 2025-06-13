package features

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/engine/util"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/go-chi/chi/v5"
	"github.com/twpayne/go-geom"
)

var emptyFeatureCollection = &domain.FeatureCollection{Features: make([]*domain.Feature, 0)}

// Features endpoint serves a FeatureCollection with the given collectionId
//
// Beware: this is one of the most performance-sensitive pieces of code in the system.
// Try to do as much initialization work outside the hot path, and only do essential
// operations inside this method.
func (f *Features) Features() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f.engine.OpenAPI.ValidateRequest(r); err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}

		collectionID := chi.URLParam(r, "collectionId")
		collection, ok := f.configuredCollections[collectionID]
		if !ok {
			handleCollectionNotFound(w, collectionID)
			return
		}
		url := featureCollectionURL{
			*f.engine.Config.BaseURL.URL,
			r.URL.Query(),
			f.engine.Config.OgcAPI.Features.Limit,
			f.configuredPropertyFilters[collection.ID],
			util.Keys(f.configuredCollections),
			collection.HasDateTime(),
		}
		encodedCursor, limit, inputSRID, outputSRID, contentCrs, bbox, referenceDate, propertyFilters, profile, err := url.parse()
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}
		w.Header().Add(engine.HeaderContentCrs, contentCrs.ToLink())

		var newCursor domain.Cursors
		var fc *domain.FeatureCollection
		if querySingleDatasource(inputSRID, outputSRID, bbox) {
			// fast path
			datasource := f.datasources[datasourceKey{srid: outputSRID.GetOrDefault(), collectionID: collectionID}]
			fc, newCursor, err = datasource.GetFeatures(r.Context(), collectionID, ds.FeaturesCriteria{
				Cursor:           encodedCursor.Decode(url.checksum()),
				Limit:            limit,
				InputSRID:        inputSRID,
				OutputSRID:       outputSRID,
				Bbox:             bbox,
				TemporalCriteria: createTemporalCriteria(collection, referenceDate),
				PropertyFilters:  propertyFilters,
				// Add filter, filter-lang
			}, f.axisOrderBySRID[outputSRID.GetOrDefault()], profile)
			if err != nil {
				handleFeaturesQueryError(w, collectionID, err)
				return
			}
		} else {
			// slower path: get feature ids by input CRS (step 1), then the actual features in output CRS (step 2)
			var fids []int64
			datasource := f.datasources[datasourceKey{srid: inputSRID.GetOrDefault(), collectionID: collectionID}]
			fids, newCursor, err = datasource.GetFeatureIDs(r.Context(), collectionID, ds.FeaturesCriteria{
				Cursor:           encodedCursor.Decode(url.checksum()),
				Limit:            limit,
				InputSRID:        inputSRID,
				OutputSRID:       outputSRID,
				Bbox:             bbox,
				TemporalCriteria: createTemporalCriteria(collection, referenceDate),
				PropertyFilters:  propertyFilters,
				// Add filter, filter-lang
			})
			if err == nil && fids != nil {
				// this is step 2: get the actual features in output CRS by feature ID
				datasource = f.datasources[datasourceKey{srid: outputSRID.GetOrDefault(), collectionID: collectionID}]
				fc, err = datasource.GetFeaturesByID(r.Context(), collectionID, fids, f.axisOrderBySRID[outputSRID.GetOrDefault()], profile)
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
			f.html.features(w, r, collection, newCursor, url, limit, &referenceDate,
				propertyFilters, f.configuredPropertyFilters[collectionID], fc)
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

func querySingleDatasource(input domain.SRID, output domain.SRID, bbox *geom.Bounds) bool {
	return bbox == nil ||
		int(input) == int(output) ||
		(int(input) == domain.UndefinedSRID && int(output) == domain.WGS84SRID) ||
		(int(input) == domain.WGS84SRID && int(output) == domain.UndefinedSRID)
}

func createTemporalCriteria(collection config.GeoSpatialCollection, referenceDate time.Time) ds.TemporalCriteria {
	var temporalCriteria ds.TemporalCriteria
	if collection.HasDateTime() {
		temporalCriteria = ds.TemporalCriteria{
			ReferenceDate:     referenceDate,
			StartDateProperty: collection.Metadata.TemporalProperties.StartDate,
			EndDateProperty:   collection.Metadata.TemporalProperties.EndDate}
	}
	return temporalCriteria
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
