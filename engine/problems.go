package engine

import (
	"log"
	"net/http"

	"schneider.vip/problem"
)

const (
	messageInternalServer = "An unexpected error has occurred, try again or contact support if the problem persists"
)

var (
	ProblemServerError   = problem.Of(http.StatusInternalServerError).Append(problem.Detail(messageInternalServer))
	ProblemBadRequest    = problem.Of(http.StatusBadRequest)
	ProblemNotFound      = problem.Of(http.StatusNotFound)
	ProblemNotAcceptable = problem.Of(http.StatusNotAcceptable)
)

func RenderProblem(p *problem.Problem, w http.ResponseWriter, details ...string) {
	for _, detail := range details {
		p = p.Append(problem.Detail(detail))
	}
	_, err := p.WriteTo(w)
	if err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
