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

	gokoalaEngine "github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc"
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
			name:       "Check conformance with all OGC APIs enabled in JSON",
			configFile: "examples/config_all.yaml",
			apiCall:    "http://localhost:8181/conformance?f=json",
			wantBody:   "internal/engine/testdata/expected_conformance.json",
		},
		{
			name:       "Check conformance with all OGC APIs enabled in HTML",
			configFile: "examples/config_all.yaml",
			apiCall:    "http://localhost:8181/conformance?f=html",
			wantBody:   "internal/engine/testdata/expected_conformance.html",
		},
		{
			name:       "Serve multiple OGC APIs for single collection in JSON",
			configFile: "internal/engine/testdata/config_multiple_ogc_apis_single_collection.yaml",
			apiCall:    "http://localhost:8180/collections/NewYork?f=json",
			wantBody:   "internal/engine/testdata/expected_multiple_ogc_apis_single_collection.json",
		},
		{
			name:       "Serve multiple OGC APIs for single collection in HTML",
			configFile: "internal/engine/testdata/config_multiple_ogc_apis_single_collection.yaml",
			apiCall:    "http://localhost:8180/collections/NewYork?f=html",
			wantBody:   "internal/engine/testdata/expected_multiple_ogc_apis_single_collection.html",
		},
		{
			name:       "Serve multiple Feature Tables from single GeoPackage",
			configFile: "internal/ogc/features/testdata/config_features_bag_multiple_feature_tables.yaml",
			apiCall:    "http://localhost:8180/collections?f=json",
			wantBody:   "internal/ogc/features/testdata/expected_multiple_feature_tables_single_geopackage.json",
		},
		{
			name:       "Check conformance of OGC API Processes",
			configFile: "internal/engine/testdata/config_processes.yaml",
			apiCall:    "http://localhost:8181/conformance?f=html",
			wantBody:   "internal/engine/testdata/expected_processes_conformance.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			eng, err := gokoalaEngine.NewEngine(tt.configFile, "", false, true)
			assert.NoError(t, err)
			ogc.SetupBuildingBlocks(eng)

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
			case strings.HasSuffix(tt.apiCall, "html"):
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
