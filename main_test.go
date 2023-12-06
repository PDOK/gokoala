package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gokoalaEngine "github.com/PDOK/gokoala/engine"
	"github.com/stretchr/testify/assert"
)

func Test_newRouter(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		apiCall    string
		wantBody   string
	}{
		{
			name:       "multiple_ogc_apis_single_collection",
			configFile: "engine/testdata/config_multiple_ogc_apis_single_collection.yaml",
			apiCall:    "http://localhost:8180/collections/NewYork?f=json",
			wantBody:   "engine/testdata/expected_multiple_ogc_apis_single_collection.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			eng := gokoalaEngine.NewEngine(tt.configFile, "")
			router := newRouter(eng, false)

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, tt.apiCall, nil)
			if err != nil {
				t.Fatal(err)
			}

			// when
			router.ServeHTTP(recorder, req)

			// then
			assert.Equal(t, http.StatusOK, recorder.Code)
			expectedBody, err := os.ReadFile(tt.wantBody)
			if err != nil {
				log.Fatal(err)
			}
			assert.JSONEq(t, recorder.Body.String(), string(expectedBody))
		})
	}
}
