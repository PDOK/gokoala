package core

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/PDOK/gomagpie/internal/engine"
	"github.com/stretchr/testify/require"

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
			name: "landing page as HTML",
			fields: fields{
				configFile: "internal/engine/testdata/config_minimal.yaml",
				url:        "http://localhost:8080/?f=html",
			},
			want: want{
				body:       "<title>Minimal OGC API</title>",
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

			newEngine, err := engine.NewEngine(tt.fields.configFile, false, true)
			require.NoError(t, err)
			core := NewCommonCore(newEngine)
			handler := core.LandingPage()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.body)
		})
	}
}

func TestCommonCore_Conformance(t *testing.T) {
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
			name: "conformance as JSON",
			fields: fields{
				configFile: "internal/engine/testdata/config_minimal.yaml",
				url:        "http://localhost:8080/conformance?f=json",
			},
			want: want{
				body:       "conformsTo",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "conformance as HTML",
			fields: fields{
				configFile: "internal/engine/testdata/config_minimal.yaml",
				url:        "http://localhost:8080/conformance?f=html",
			},
			want: want{
				body:       "conformiteitsklassen",
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

			newEngine, err := engine.NewEngine(tt.fields.configFile, false, true)
			require.NoError(t, err)
			core := NewCommonCore(newEngine)
			handler := core.Conformance()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.body)
		})
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
				configFile: "internal/engine/testdata/config_minimal.yaml",
				url:        "http://localhost:8080/api?f=json",
			},
			want: want{
				body:       "OpenAPI",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "OpenAPI as HTML",
			fields: fields{
				configFile: "internal/engine/testdata/config_minimal.yaml",
				url:        "http://localhost:8080/api?f=html",
			},
			want: want{
				body:       "GomagpieLayoutPlugin", // exists on swagger page, this to make sure we get HTML
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

			newEngine, err := engine.NewEngine(tt.fields.configFile, false, true)
			require.NoError(t, err)
			core := NewCommonCore(newEngine)
			handler := core.API()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.body)
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
