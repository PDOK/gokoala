package geospatial

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
	"golang.org/x/text/language"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
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

func TestNewCollections(t *testing.T) {
	type args struct {
		e *engine.Engine
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test render templates with Collections (using OGC GeoVolumes config, since that contains collections)",
			args: args{
				e: engine.NewEngineWithConfig(&config.Config{
					Version:            "1.0.0",
					Title:              "Test API",
					Abstract:           "Test API description",
					AvailableLanguages: []config.Language{{Tag: language.Dutch}},
					BaseURL:            config.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: config.OgcAPI{
						GeoVolumes: &config.OgcAPI3dGeoVolumes{
							TileServer: config.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
							Collections: config.GeoSpatialCollections{
								config.GeoSpatialCollection{ID: "buildings"},
								config.GeoSpatialCollection{ID: "obstacles"},
							},
						},
					},
				}, "", false, true),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			collections := NewCollections(test.args.e)
			assert.NotEmpty(t, collections.engine.Templates.RenderedTemplates)
		})
	}
}

func TestNewCollections_Collections(t *testing.T) {
	type fields struct {
		configFile  string
		url         string
		containerID string
	}
	type want struct {
		bodyContains string
		statusCode   int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "container_1",
			fields: fields{
				configFile:  "internal/ogc/geovolumes/testdata/config_minimal_3d.yaml",
				url:         "http://localhost:8080/collections",
				containerID: "container_1",
			},
			want: want{
				bodyContains: "\"title\": \"container_1\"",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "container_2",
			fields: fields{
				configFile:  "internal/ogc/geovolumes/testdata/config_minimal_3d.yaml",
				url:         "http://localhost:8080/collections",
				containerID: "container_1",
			},
			want: want{
				bodyContains: "\"title\": \"container_2\"",
				statusCode:   http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createCollectionsRequest(tt.fields.url)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			collections := NewCollections(newEngine)
			handler := collections.Collections()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.bodyContains)
		})
	}
}

func TestNewCollections_Collection(t *testing.T) {
	type fields struct {
		configFile  string
		url         string
		containerID string
	}
	type want struct {
		bodyContains string
		statusCode   int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "container_1",
			fields: fields{
				configFile:  "internal/ogc/geovolumes/testdata/config_minimal_3d.yaml",
				url:         "http://localhost:8080/collections/:collectionId",
				containerID: "container_1",
			},
			want: want{
				bodyContains: "\"title\": \"container_1\"",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "container_2",
			fields: fields{
				configFile:  "internal/ogc/geovolumes/testdata/config_minimal_3d.yaml",
				url:         "http://localhost:8080/collections/:collectionId",
				containerID: "container_2",
			},
			want: want{
				bodyContains: "\"title\": \"container_2\"",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "container_404",
			fields: fields{
				configFile:  "internal/ogc/geovolumes/testdata/config_minimal_3d.yaml",
				url:         "http://localhost:8080/collections/:collectionId",
				containerID: "container_404",
			},
			want: want{
				bodyContains: "Not Found",
				statusCode:   http.StatusNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createCollectionRequest(tt.fields.url, tt.fields.containerID)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			collections := NewCollections(newEngine)
			handler := collections.Collection()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.bodyContains)
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
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	return rr, ts
}

func createCollectionsRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}

func createCollectionRequest(url string, containerID string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()

	rctx.URLParams.Add("collectionId", containerID)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}
