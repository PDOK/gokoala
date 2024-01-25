package engine

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template/config files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestEngine_ServePage_LandingPage(t *testing.T) {
	// given
	engine, err := NewEngine("engine/testdata/config_minimal.yaml", "", false, true)
	assert.NoError(t, err)

	templateKey := NewTemplateKey("ogc/common/core/templates/landing-page.go.json")
	engine.RenderTemplates("/", nil, templateKey)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engine.ServePage(w, r, templateKey)
	})

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// when
	handler.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get(HeaderContentType))
	assert.Contains(t, recorder.Body.String(), "This is a minimal OGC API, offering only OGC API Common")
}

func TestEngine_ReverseProxy(t *testing.T) {
	// given
	mockTargetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(fmt.Sprintf("Mock response, received header %s", r.Header.Get(HeaderBaseURL))))
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer mockTargetServer.Close()

	engine, targetURL := makeEngine(mockTargetServer)
	rec, req := makeAPICall(t, mockTargetServer)

	// when
	engine.ReverseProxy(rec, req, targetURL, false, "")

	// then
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, rec.Body.String(), "Mock response, received header https://api.foobar.example/")
}

func TestEngine_ReverseProxy_Status204(t *testing.T) {
	// given
	mockTargetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockTargetServer.Close()

	engine, targetURL := makeEngine(mockTargetServer)
	rec, req := makeAPICall(t, mockTargetServer)

	// when
	engine.ReverseProxy(rec, req, targetURL, true, "audio/wav")

	// then
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "audio/wav", rec.Header().Get(HeaderContentType))
}

func makeEngine(mockTargetServer *httptest.Server) (*Engine, *url.URL) {
	engine := &Engine{
		Config: &Config{
			BaseURL: YAMLURL{&url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
		},
	}
	targetURL, _ := url.Parse(mockTargetServer.URL)
	return engine, targetURL
}

func makeAPICall(t *testing.T, mockTargetServer *httptest.Server) (*httptest.ResponseRecorder, *http.Request) {
	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, mockTargetServer.URL+"/some/path", nil)
	if err != nil {
		t.Fatal(err)
	}
	return rec, req
}
