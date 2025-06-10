package features

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Feature endpoint serves a single Feature
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

		datasource := f.datasources[datasourceKey{srid: outputSRID.GetOrDefault(), collectionID: collectionID}]
		feat, err := datasource.GetFeature(r.Context(), collectionID, featureID, f.swapXY[outputSRID], f.defaultProfile)
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

func handleFeatureNotFound(w http.ResponseWriter, collectionID string, featureID any) {
	msg := fmt.Sprintf("the requested feature with id: %v does not exist in collection '%v'", featureID, collectionID)
	log.Println(msg)
	engine.RenderProblem(engine.ProblemNotFound, w, msg)
}
