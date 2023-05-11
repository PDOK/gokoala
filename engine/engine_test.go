package engine

import (
	"net/http"
	"net/http/httptest"
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
	engine := NewEngine("engine/testdata/config_minimal.yaml", "", "")

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
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.Contains(t, recorder.Body.String(), "This is a minimal OGC API, offering only OGC API Common")
}
