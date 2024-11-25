package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	gomagpieEngine "github.com/PDOK/gomagpie/internal/engine"
	"github.com/PDOK/gomagpie/internal/ogc"
	"github.com/stretchr/testify/assert"
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func Test_newRouter(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		apiCall    string
		wantBody   string
	}{
		{
			name:       "Check conformance with OGC APIs enabled in JSON",
			configFile: "examples/config.yaml",
			apiCall:    "http://localhost:8181/conformance?f=json",
			wantBody:   "internal/engine/testdata/expected_conformance.json",
		},
		{
			name:       "Check conformance with OGC APIs enabled in HTML",
			configFile: "examples/config.yaml",
			apiCall:    "http://localhost:8181/conformance?f=html",
			wantBody:   "internal/engine/testdata/expected_conformance.html",
		},
		{
			name:       "Should have valid sitemap XML",
			configFile: "examples/config.yaml",
			apiCall:    "http://localhost:8181/sitemap.xml",
			wantBody:   "internal/engine/testdata/expected_sitemap.xml",
		},
		{
			name:       "Should have valid structured data of type 'Dataset' on landing page",
			configFile: "examples/config.yaml",
			apiCall:    "http://localhost:8181?f=html",
			wantBody:   "internal/engine/testdata/expected_dataset_landingpage.json",
		},
		{
			name:       "Should have valid structured data of type 'Dataset' on (each) collection page",
			configFile: "examples/config.yaml",
			apiCall:    "http://localhost:8181/collections/addresses?f=html",
			wantBody:   "internal/engine/testdata/expected_dataset_collection.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			eng, err := gomagpieEngine.NewEngine(tt.configFile, false, true)
			assert.NoError(t, err)
			ogc.SetupBuildingBlocks(eng, "PLACEHOLDER DB CONNECTION STRING")

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, tt.apiCall, nil)
			if err != nil {
				t.Fatal(err)
			}

			// when
			eng.Router.ServeHTTP(recorder, req)

			// then
			assert.Equal(t, http.StatusOK, recorder.Code)
			expectedBody, err := os.ReadFile(tt.wantBody)
			if err != nil {
				log.Fatal(err)
			}
			log.Print(recorder.Body.String()) // to ease debugging
			switch {
			case strings.HasSuffix(tt.apiCall, "json"):
				assert.JSONEq(t, recorder.Body.String(), string(expectedBody))
			case strings.HasSuffix(tt.apiCall, "html") || strings.HasSuffix(tt.apiCall, "xml"):
				assert.Contains(t, normalize(recorder.Body.String()), normalize(string(expectedBody)))
			default:
				log.Fatalf("implement support to test format: %s", tt.apiCall)
			}
		})
	}
}

func normalize(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), ""))
}
