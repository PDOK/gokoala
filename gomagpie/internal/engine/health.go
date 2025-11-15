package engine

import (
	"net/http"
)

func newHealthEndpoint(e *Engine) {
	e.Router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		SafeWrite(w.Write, []byte("OK"))
	})
}
