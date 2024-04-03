package engine

import (
	"log"
	"net/http"
	"time"

	"schneider.vip/problem"
)

const (
	timestampKey             = "timestamp"
	defaultMessageServerErr  = "An unexpected error has occurred, try again or contact support if the problem persists"
	defaultMessageBadGateway = "Failed to proxy request, try again or contact support if the problem persists"
)

var Now = time.Now // allow mocking

// The following problems should be added to openapi/problems.go.json
var (
	ProblemBadRequest    = problem.Of(http.StatusBadRequest)
	ProblemNotFound      = problem.Of(http.StatusNotFound)
	ProblemNotAcceptable = problem.Of(http.StatusNotAcceptable)
	ProblemServerError   = problem.Of(http.StatusInternalServerError).Append(problem.Detail(defaultMessageServerErr))
	ProblemBadGateway    = problem.Of(http.StatusBadGateway).Append(problem.Detail(defaultMessageBadGateway))
)

func RenderProblem(p *problem.Problem, w http.ResponseWriter, details ...string) {
	for _, detail := range details {
		p = p.Append(problem.Detail(detail))
	}
	p = p.Append(problem.Custom(timestampKey, Now().UTC().Format(time.RFC3339)))
	_, err := p.WriteTo(w)
	if err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
