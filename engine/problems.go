package engine

import (
	"log"
	"net/http"
	"time"

	"schneider.vip/problem"
)

const (
	timestampKey          = "timestamp"
	messageInternalServer = "An unexpected error has occurred, try again or contact support if the problem persists"
)

var (
	Now                  = time.Now // allow mocking
	ProblemServerError   = problem.Of(http.StatusInternalServerError).Append(problem.Detail(messageInternalServer))
	ProblemBadRequest    = problem.Of(http.StatusBadRequest)
	ProblemNotFound      = problem.Of(http.StatusNotFound)
	ProblemNotAcceptable = problem.Of(http.StatusNotAcceptable)
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
