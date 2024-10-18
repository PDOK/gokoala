package features

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/PDOK/gokoala/internal/engine"
)

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
