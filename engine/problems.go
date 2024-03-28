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
	ProblemInternalServer = problem.Of(http.StatusInternalServerError).Append(problem.Detail(messageInternalServer))
	ProblemBadRequest     = problem.Of(http.StatusBadRequest)
)

func HandleProblem(p *problem.Problem, w http.ResponseWriter, details ...string) {
	for _, detail := range details {
		p = p.Append(problem.Detail(detail))
	}
	_, err := p.WriteTo(w)
	if err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
