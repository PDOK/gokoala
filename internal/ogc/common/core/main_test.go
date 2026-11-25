package core

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/stretchr/testify/require"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestLandingPageDiscoversRecordsCatalogs(t *testing.T) {
	theme, err := config.NewTheme("internal/engine/testdata/test_theme.yaml")
	require.NoError(t, err)
	serverConfig := &config.Config{
		Version:            "1.0.0",
		Title:              "Records API",
		ServiceIdentifier:  "records",
		Abstract:           "Dataset records",
		AvailableLanguages: []config.Language{{Tag: language.Dutch}},
		License: config.License{
			Name: "CC0",
			URL:  config.URL{URL: &url.URL{Scheme: "https", Host: "creativecommons.org", Path: "/publicdomain/zero/1.0/"}},
		},
		BaseURL: config.URL{URL: &url.URL{Scheme: "https", Host: "api.example.org"}},
		OgcAPI: config.OgcAPI{Records: &config.OgcAPIRecords{
			Collections: config.RecordsCollections{{ID: "datasets"}},
		}},
	}
	newEngine := engine.NewEngineWithConfig(serverConfig, theme, "", false, true)
	core := NewCommonCore(newEngine, ExtraConformanceClasses{})
	request := httptest.NewRequest(http.MethodGet, "https://api.example.org/?f=json", nil)
	recorder := httptest.NewRecorder()
	core.LandingPage().ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "http://www.opengis.net/def/rel/ogc/1.0/ogc-catalog")
	assert.Contains(t, recorder.Body.String(), "https://api.example.org/collections")
	assert.Contains(t, recorder.Body.String(), "https://api.example.org/collections/datasets/items")

	for _, format := range []string{"html", "md"} {
		request := httptest.NewRequest(http.MethodGet, "https://api.example.org/?f="+format, nil)
		recorder := httptest.NewRecorder()
		core.LandingPage().ServeHTTP(recorder, request)
		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "rel=\"http://www.opengis.net/def/rel/ogc/1.0/ogc-catalog\"")
		assert.Contains(t, recorder.Body.String(), "collections/datasets/items")
	}
}

func TestCommonCore_LandingPage(t *testing.T) {
	type fields struct {
		configFile string
		url        string
	}
	type want struct {
		body       string
		statusCode int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "landing page as JSON",
			fields: fields{
				configFile: "internal/engine/testdata/config_minimal.yaml",
				url:        "http://localhost:8080/?f=json",
			},
			want: want{
				body:       "Landing page as JSON",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "landing page as JSON with Thumbnail",
			fields: fields{
				configFile: "internal/engine/testdata/config_resources_dir.yaml",
				url:        "http://localhost:8080/?f=json",
			},
			want: want{
				body:       "\"rel\": \"preview\",\n   \"type\": \"image/jpeg\",\n   \"title\": \"Thumbnail for the dataset shared via this API\"",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "landing page as HTML",
			fields: fields{
				configFile: "internal/engine/testdata/config_minimal.yaml",
				url:        "http://localhost:8080/?f=html",
			},
			want: want{
				body:       "<title>Minimal OGC API (OGC API)</title>",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "landing page as Markdown",
			fields: fields{
				configFile: "internal/engine/testdata/config_minimal.yaml",
				url:        "http://localhost:8080/?f=md",
			},
			want: want{
				body:       "# Minimal OGC API (OGC API)",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "landing page as HTML with Thumbnail",
			fields: fields{
				configFile: "internal/engine/testdata/config_resources_dir.yaml",
				url:        "http://localhost:8080/?f=html",
			},
			want: want{
				body:       "<img src=\"resources/thumbnail.jpg\"",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createRequest(tt.fields.url)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "internal/engine/testdata/test_theme.yaml", "", false, true)
			require.NoError(t, err)
			core := NewCommonCore(newEngine, ExtraConformanceClasses{})
			handler := core.LandingPage()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.body)
		})
	}
}

func TestCommonCore_Conformance(t *testing.T) {
	type fields struct {
		configFile         string
		url                string
		supportsAttributes bool
	}
	type want struct {
		body       string
		statusCode int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "conformance as JSON",
			fields: fields{
				configFile:         "internal/engine/testdata/config_multiple_ogc_apis_single_collection.yaml",
				url:                "http://localhost:8080/conformance?f=json",
				supportsAttributes: false,
			},
			want: want{
				body:       "conformsTo",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "conformance as HTML",
			fields: fields{
				configFile:         "internal/engine/testdata/config_multiple_ogc_apis_single_collection.yaml",
				url:                "http://localhost:8080/conformance?f=html",
				supportsAttributes: false,
			},
			want: want{
				body:       "conformiteitsklassen",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "conformance as Markdown",
			fields: fields{
				configFile:         "internal/engine/testdata/config_multiple_ogc_apis_single_collection.yaml",
				url:                "http://localhost:8080/conformance?f=md",
				supportsAttributes: false,
			},
			want: want{
				body:       "| http://www.opengis.net/spec/ogcapi-common-1/1.0/conf/core",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createRequest(tt.fields.url)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "internal/engine/testdata/test_theme.yaml", "", false, true)
			require.NoError(t, err)
			core := NewCommonCore(newEngine, ExtraConformanceClasses{})
			handler := core.Conformance()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.body)
		})
	}
}

func TestRecordsConformanceClasses(t *testing.T) {
	theme, err := config.NewTheme("internal/engine/testdata/test_theme.yaml")
	require.NoError(t, err)
	enableCQL := true
	serverConfig := &config.Config{
		Version:            "1.0.0",
		Title:              "Records API",
		ServiceIdentifier:  "records",
		Abstract:           "Dataset records",
		AvailableLanguages: []config.Language{{Tag: language.Dutch}},
		License: config.License{
			Name: "CC0",
			URL:  config.URL{URL: &url.URL{Scheme: "https", Host: "creativecommons.org", Path: "/publicdomain/zero/1.0/"}},
		},
		BaseURL: config.URL{URL: &url.URL{Scheme: "https", Host: "api.example.org"}},
		OgcAPI: config.OgcAPI{Records: &config.OgcAPIRecords{
			Collections: config.RecordsCollections{{
				ID: "datasets",
				Filters: config.FeatureFilters{
					Properties: []config.Queryable{{Name: "title"}},
					CQL:        config.CQL{Enable: &enableCQL},
				},
			}},
		}},
	}
	newEngine := engine.NewEngineWithConfig(serverConfig, theme, "", false, true)
	core := NewCommonCore(newEngine, ExtraConformanceClasses{})
	for _, format := range []string{"json", "html", "md"} {
		request := httptest.NewRequest(http.MethodGet, "https://api.example.org/conformance?f="+format, nil)
		recorder := httptest.NewRecorder()
		core.Conformance().ServeHTTP(recorder, request)
		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "http://www.opengis.net/spec/ogcapi-records-1/1.0/conf/searchable-catalog")
		assert.Contains(t, recorder.Body.String(), "https://www.opengis.net/spec/ogcapi-records-5/1.0/conf/stac-extensions")
		assert.Contains(t, recorder.Body.String(), "https://www.opengis.net/spec/ogcapi-records-5/1.0/conf/stac-items")
		assert.Contains(t, recorder.Body.String(), "http://www.opengis.net/spec/ogcapi-records-1/1.0/conf/filtering")
		assert.Contains(t, recorder.Body.String(), "http://www.opengis.net/spec/cql2/1.0/conf/cql2-text")
	}
}

func TestCommonCore_API(t *testing.T) {
	type fields struct {
		configFile string
		url        string
	}
	type want struct {
		body       string
		statusCode int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "OpenAPI as JSON",
			fields: fields{
				configFile: "internal/engine/testdata/config_multiple_ogc_apis_single_collection.yaml",
				url:        "http://localhost:8080/api?f=json",
			},
			want: want{
				body:       "\"openapi\":",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "OpenAPI as HTML",
			fields: fields{
				configFile: "internal/engine/testdata/config_multiple_ogc_apis_single_collection.yaml",
				url:        "http://localhost:8080/api?f=html",
			},
			want: want{
				body:       "GoKoalaLayoutPlugin", // exists on swagger page, this to make sure we get HTML
				statusCode: http.StatusOK,
			},
		},
		{
			name: "OpenAPI as Markdown",
			fields: fields{
				configFile: "internal/engine/testdata/config_multiple_ogc_apis_single_collection.yaml",
				url:        "http://localhost:8080/api?f=md",
			},
			want: want{
				body:       "[JSON](http://localhost:8180/api?f=json)", // link to the full OpenAPI spec in JSON
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createRequest(tt.fields.url)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "internal/engine/testdata/test_theme.yaml", "", false, true)
			require.NoError(t, err)
			core := NewCommonCore(newEngine, ExtraConformanceClasses{})
			handler := core.API()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			actualBody := rr.Body.String()
			log.Print("ACTUAL:\n" + actualBody)
			assert.Contains(t, actualBody, tt.want.body)
		})
	}
}

func createMockServer() (*httptest.ResponseRecorder, *httptest.Server) {
	rr := httptest.NewRecorder()
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engine.SafeWrite(w.Write, []byte(r.URL.String()))
	}))
	defer ts.Listener.Close()
	ts.Listener = l
	ts.Start()

	return rr, ts
}

func createRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	return req, err
}
