package search

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/PDOK/gokoala/internal/engine"
)

// log error, but send generic message to client to prevent possible information leakage from datasource
func handleQueryError(w http.ResponseWriter, err error) {
	msg := "failed to fulfill search request"
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		// provide more context when user hits the query timeout
		msg += ": querying took too long (timeout encountered). Simplify your request and try again, or contact support"
	}
	log.Printf("%s, error: %v\n", msg, err)
	engine.RenderProblem(engine.ProblemServerError, w, msg) // don't include sensitive information in details msg
}
