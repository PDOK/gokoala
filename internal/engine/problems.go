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

type ProblemKind int

var Now = time.Now // allow mocking

// The following problems should be added to openapi/problems.go.json
var (
	ProblemBadRequest    = ProblemKind(http.StatusBadRequest)
	ProblemNotFound      = ProblemKind(http.StatusNotFound)
	ProblemNotAcceptable = ProblemKind(http.StatusNotAcceptable)
	ProblemServerError   = ProblemKind(http.StatusInternalServerError)
	ProblemBadGateway    = ProblemKind(http.StatusBadGateway)
)

// RenderProblem writes FC 7807 (https://tools.ietf.org/html/rfc7807) problem to response output.
// Only the listed problem kinds are supported since they should be advertised in the OpenAPI spec.
// Optionally a caller may add a details (single string) about the problem. Warning: Be sure to not
// include sensitive information in the details string!
func RenderProblem(kind ProblemKind, w http.ResponseWriter, details ...string) {
	p := problem.Of(int(kind))
	if kind == ProblemServerError {
		p = p.Append(problem.Detail(defaultMessageServerErr))
	} else if kind == ProblemBadGateway {
		p = p.Append(problem.Detail(defaultMessageBadGateway))
	} else if len(details) > 0 {
		p = p.Append(problem.Detail(details[0]))
	}
	p = p.Append(problem.Custom(timestampKey, Now().UTC().Format(time.RFC3339)))
	_, err := p.WriteTo(w)
	if err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
