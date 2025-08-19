package engine

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnexpectedProblemRecoverer(t *testing.T) {
	// given
	defer func() {
		_ = recover()
	}()
	w := httptest.NewRecorder()

	r := newRouter("1.2.3", true, false)
	r.Get("/panic", func(_ http.ResponseWriter, _ *http.Request) {
		panic("oops")
	})

	req, err := http.NewRequest(http.MethodGet, "/panic", nil)
	if err != nil {
		t.Fatal(err)
	}

	// when
	r.ServeHTTP(w, req)

	// then
	assert.Contains(t, w.Body.String(), "{\"detail\":\"An unexpected error has occurred, try again or contact support if the problem persists\",\"status\":500,")
}

func TestExpectedProblemRecoverer(t *testing.T) {
	// given
	defer func() {
		_ = recover()
	}()
	w := httptest.NewRecorder()

	tests := []struct {
		name         string
		path         string
		handlerFunc  http.HandlerFunc
		bodyContains string
	}{
		{"Bad request", "/panic", func(_ http.ResponseWriter, _ *http.Request) { RenderProblem(ProblemBadRequest, w, "foo bar baz") }, "{\"detail\":\"foo bar baz\",\"status\":400,"},
		{"Bad gateway", "/noproxy", func(_ http.ResponseWriter, _ *http.Request) { RenderProblem(ProblemBadGateway, w) }, "{\"detail\":\"Failed to proxy request, try again or contact support if the problem persists\",\"status\":502,"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRouter("1.2.3", true, false)
			r.Get(tt.path, tt.handlerFunc)

			req, err := http.NewRequest(http.MethodGet, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			// when
			r.ServeHTTP(w, req)

			// then
			assert.Contains(t, w.Body.String(), tt.bodyContains)
		})
	}
}

func TestAbortRecoverer(t *testing.T) {
	defer func() {
		// then
		rcv := recover()
		if rcv != http.ErrAbortHandler { //nolint:errorlint // already so in Chi
			t.Fatalf("http.ErrAbortHandler should not be recovered")
		}
	}()

	// given
	w := httptest.NewRecorder()

	r := newRouter("1.2.3", true, false)
	r.Get("/panic", func(_ http.ResponseWriter, _ *http.Request) {
		panic(http.ErrAbortHandler)
	})

	req, err := http.NewRequest(http.MethodGet, "/panic", nil)
	if err != nil {
		t.Fatal(err)
	}

	// when
	r.ServeHTTP(w, req)
}
