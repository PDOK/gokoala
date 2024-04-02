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
	assert.Equal(t, "{\"detail\":\"An unexpected error has occurred, try again or contact support if the problem persists\",\"status\":500,\"title\":\"Internal Server Error\"}", w.Body.String())
}

func TestExpectedProblemRecoverer(t *testing.T) {
	// given
	defer func() {
		_ = recover()
	}()
	w := httptest.NewRecorder()

	r := newRouter("1.2.3", true, false)
	r.Get("/panic", func(_ http.ResponseWriter, _ *http.Request) {
		RenderProblem(ProblemBadRequest, w, "foo bar baz")
	})

	req, err := http.NewRequest(http.MethodGet, "/panic", nil)
	if err != nil {
		t.Fatal(err)
	}

	// when
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, "{\"detail\":\"foo bar baz\",\"status\":400,\"title\":\"Bad Request\"}", w.Body.String())
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
