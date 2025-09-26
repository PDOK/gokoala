package features

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Feature endpoint serves a single Feature by ID
//
//nolint:cyclop
func (f *Features) Feature() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f.engine.OpenAPI.ValidateRequest(r); err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}

		collectionID := chi.URLParam(r, "collectionId")
		collection, ok := f.configuredCollections[collectionID]
		if !ok {
			handleCollectionNotFound(w, collection.ID)
			return
		}
		featureID, err := parseFeatureID(r)
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}
		url := featureURL{*f.engine.Config.BaseURL.URL,
			r.URL.Query(),
			f.schemas[collection.ID],
		}
		outputSRID, contentCrs, profile, err := url.parse()
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}
		w.Header().Add(engine.HeaderContentCrs, contentCrs.ToLink())

		// validation completed, now get the feature
		datasource := f.datasources[datasourceKey{srid: outputSRID.GetOrDefault(), collectionID: collection.ID}]
		feat, err := datasource.GetFeature(r.Context(), collection.ID, featureID,
			outputSRID, f.axisOrderBySRID[outputSRID.GetOrDefault()], profile)
		if err != nil {
			handleFeatureQueryError(w, collection.ID, featureID, err)
			return
		}
		if feat == nil {
			handleFeatureNotFound(w, collection.ID, featureID)
			return
		}

		// render output
		format := f.engine.CN.NegotiateFormat(r)
		switch datasource.GetCollectionType(collection.ID) {
		case domain.Features:
			switch format {
			case engine.FormatHTML:
				f.html.feature(w, r, collection, feat)
			case engine.FormatGeoJSON, engine.FormatJSON:
				f.json.featureAsGeoJSON(w, r, collectionID, collection.Features, feat, url)
			case engine.FormatJSONFG:
				f.json.featureAsJSONFG(w, r, collectionID, collection.Features, feat, url, contentCrs)
			default:
				engine.RenderProblem(engine.ProblemNotAcceptable, w, fmt.Sprintf("format '%s' is not supported", format))
				return
			}
		case domain.Attributes:
			switch format {
			case engine.FormatHTML:
				f.html.attribute(w, r, collection, feat)
			case engine.FormatJSON:
				f.json.featureAsAttributeJSON(w, r, collectionID, feat, url)
			default:
				engine.RenderProblem(engine.ProblemNotAcceptable, w, fmt.Sprintf("format '%s' is not supported", format))
				return
			}
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

// log error but send a generic message to the client to prevent possible information leakage from datasource
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
