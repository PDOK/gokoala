package features

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/PDOK/gokoala/engine"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestFeatures_CollectionContent(t *testing.T) {
	type fields struct {
		configFile   string
		url          string
		collectionID string
		format       string
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
			name: "Request GeoJSON for 'foo' collection using default limit",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items",
				collectionID: "foo",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_foo_collection.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request GeoJSON for 'foo' collection using limit of 2",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=2",
				collectionID: "foo",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_foo_collection_with_limit.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request GeoJSON for 'foo' collection using limit of 2 and cursor to next page",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features.yaml",
				url:          "http://localhost:8080/collections/tunneldelen/items?f=json&cursor=iUMnUmcz&limit=2",
				collectionID: "foo",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_foo_collection_with_cursor.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request non existing feature collection",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?cursor=9&limit=2",
				collectionID: "doesnotexist",
				format:       "json",
			},
			want: want{
				body:       "",
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "Request with unknown query params",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?foo=bar",
				collectionID: "foo",
				format:       "json",
			},
			want: want{
				body:       "",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Request with invalid limit",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=notanumber",
				collectionID: "foo",
				format:       "json",
			},
			want: want{
				body:       "",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Request with negative limit",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=-200",
				collectionID: "foo",
				format:       "json",
			},
			want: want{
				body:       "",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Request HTML for 'foo' collection using limit of 1",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=1",
				collectionID: "foo",
				format:       "html",
			},
			want: want{
				body:       "ogc/features/testdata/expected_foo_collection_snippet.html",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gofakeit.Seed(1) // Uses consistent fake data.

			req, err := createRequest(tt.fields.url, tt.fields.collectionID, "", tt.fields.format)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine := engine.NewEngine(tt.fields.configFile, "")
			features := NewFeatures(newEngine, chi.NewRouter())
			handler := features.CollectionContent()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			if tt.want.body != "" {
				expectedBody, err := os.ReadFile(tt.want.body)
				if err != nil {
					log.Fatal(err)
				}
				switch {
				case tt.fields.format == "json":
					assert.JSONEq(t, string(expectedBody), rr.Body.String())
				case tt.fields.format == "html":
					assert.Contains(t, normalize(rr.Body.String()), normalize(string(expectedBody)))
				default:
					log.Fatalf("implement support to test format: %s", tt.fields.format)
				}
			}
		})
	}
}

func TestFeatures_Feature(t *testing.T) {
	type fields struct {
		configFile   string
		url          string
		collectionID string
		featureID    string
		format       string
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
			name: "Request GeoJSON for feature 19058835",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "foo",
				featureID:    "19058835",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_feature_19058835.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request non existing feature",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "foo",
				featureID:    "9999999999",
				format:       "json",
			},
			want: want{
				body:       "",
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "Request with unknown query params",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId?foo=bar",
				collectionID: "foo",
				featureID:    "19058835",
				format:       "json",
			},
			want: want{
				body:       "",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Request HTML for feature 19058835",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "foo",
				featureID:    "19058835",
				format:       "html",
			},
			want: want{
				body:       "ogc/features/testdata/expected_feature_19058835.html",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gofakeit.Seed(1) // Uses consistent fake data.

			req, err := createRequest(tt.fields.url, tt.fields.collectionID, tt.fields.featureID, tt.fields.format)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine := engine.NewEngine(tt.fields.configFile, "")
			features := NewFeatures(newEngine, chi.NewRouter())
			handler := features.Feature()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			if tt.want.body != "" {
				expectedBody, err := os.ReadFile(tt.want.body)
				if err != nil {
					log.Fatal(err)
				}
				switch {
				case tt.fields.format == "json":
					assert.JSONEq(t, string(expectedBody), rr.Body.String())
				case tt.fields.format == "html":
					assert.Contains(t, normalize(rr.Body.String()), normalize(string(expectedBody)))
				default:
					log.Fatalf("implement support to test format: %s", tt.fields.format)
				}
			}
		})
	}
}

func createMockServer() (*httptest.ResponseRecorder, *httptest.Server) {
	rr := httptest.NewRecorder()
	l, err := net.Listen("tcp", "localhost:9095")
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engine.SafeWrite(w.Write, []byte(r.URL.String()))
	}))
	err = ts.Listener.Close()
	if err != nil {
		log.Fatal(err)
	}
	ts.Listener = l
	ts.Start()
	return rr, ts
}

func createRequest(url string, collectionID string, featureID string, format string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("collectionId", collectionID)
	rctx.URLParams.Add("featureId", featureID)

	queryString := req.URL.Query()
	queryString.Add("f", format)
	req.URL.RawQuery = queryString.Encode()

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}

func normalize(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), ""))
}
