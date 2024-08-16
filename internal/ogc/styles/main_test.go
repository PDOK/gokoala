package styles

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
	"github.com/go-chi/chi/v5"

	"github.com/PDOK/gokoala/internal/engine"
	"golang.org/x/text/language"

	"github.com/stretchr/testify/assert"
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestNewStyles(t *testing.T) {
	type args struct {
		e *engine.Engine
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test render templates with OGC Styles config",
			args: args{
				e: engine.NewEngineWithConfig(&config.Config{
					Version:  "0.4.0",
					Title:    "Test API",
					Abstract: "Test API description",
					Resources: &config.Resources{
						Directory: ptrTo("/fakedirectory"),
					},
					AvailableLanguages: []config.Language{{Tag: language.Dutch}},
					BaseURL:            config.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: config.OgcAPI{
						GeoVolumes: nil,
						Tiles: &config.OgcAPITiles{
							DatasetTiles: &config.Tiles{
								TileServer: config.URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset"}},
								Types:      []config.TilesType{config.TilesTypeVector},
								SupportedSrs: []config.SupportedSrs{
									{
										Srs: "EPSG:28992",
										ZoomLevelRange: config.ZoomLevelRange{
											Start: 12,
											End:   12,
										},
									},
								},
							},
						},
						Styles: &config.OgcAPIStyles{
							Default: "foo",
							SupportedStyles: []config.Style{
								{
									ID:    "foo",
									Title: "bar",
								},
							},
						},
					},
				}, "", false, true),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			styles := NewStyles(test.args.e)
			assert.NotEmpty(t, styles.engine.Templates.RenderedTemplates)
		})
	}
}

func TestStyles_Style(t *testing.T) {
	type fields struct {
		configFile string
		url        string
		style      string
	}
	type want struct {
		bodyContains []string
		statusCode   int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "styles/default__netherlandsrdnewquad",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style",
				style:      "default__netherlandsrdnewquad",
			},
			want: want{
				bodyContains: []string{"\"id\": \"default\"", "tiles/NetherlandsRDNewQuad/{z}/{y}/{x}?f=mvt"},
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "styles/default__webmercatorquad",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style",
				style:      "default__webmercatorquad",
			},
			want: want{
				bodyContains: []string{"\"id\": \"default\"", "tiles/WebMercatorQuad/{z}/{y}/{x}?f=mvt"},
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "styles/default (backwards comp)",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style",
				style:      "default",
			},
			want: want{
				bodyContains: []string{"\"id\": \"default\"", "tiles/NetherlandsRDNewQuad/{z}/{y}/{x}?f=mvt"},
				statusCode:   http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createStyleRequest(tt.fields.url, tt.fields.style)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			styles := NewStyles(newEngine)
			handler := styles.Style()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			for _, c := range tt.want.bodyContains {
				assert.Contains(t, rr.Body.String(), c)
			}
		})
	}
}

func TestStyles_StyleMetadata(t *testing.T) {
	type fields struct {
		configFile string
		url        string
		style      string
	}
	type want struct {
		bodyContains []string
		statusCode   int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "styles/default__netherlandsrdnewquad",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style/metadata",
				style:      "default__netherlandsrdnewquad",
			},
			want: want{
				bodyContains: []string{"\"id\": \"default\"", "\"title\": \"Mapbox Style\"", "default__netherlandsrdnewquad"},
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "styles/default__webmercatorquad",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style/metadata",
				style:      "default__webmercatorquad",
			},
			want: want{
				bodyContains: []string{"\"id\": \"default\"", "\"title\": \"Mapbox Style\"", "default__webmercatorquad"},
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "styles/default (backwards comp)",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style/metadata",
				style:      "default",
			},
			want: want{
				bodyContains: []string{"\"id\": \"default\"", "\"title\": \"Mapbox Style\"", "default__netherlandsrdnewquad"},
				statusCode:   http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createStyleRequest(tt.fields.url, tt.fields.style)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			styles := NewStyles(newEngine)
			handler := styles.StyleMetadata()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			for _, c := range tt.want.bodyContains {
				assert.Contains(t, rr.Body.String(), c)
			}
		})
	}
}

func TestTile_Styles(t *testing.T) {
	type fields struct {
		configFile string
		url        string
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
			name: "test NetherlandsRDNewQuad present",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles",
			},
			want: want{
				bodyContains: "Test style (NetherlandsRDNewQuad)",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "test WebMercatorQuad present",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles",
			},
			want: want{
				bodyContains: "Test style (WebMercatorQuad)",
				statusCode:   http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createStylesRequest(tt.fields.url)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			styles := NewStyles(newEngine)
			handler := styles.Styles()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.bodyContains)
		})
	}
}

func createMockServer() (*httptest.ResponseRecorder, *httptest.Server) {
	rr := httptest.NewRecorder()
	l, err := net.Listen("tcp", "localhost:10090")
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

func createStyleRequest(url string, style string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("style", style)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}

func createStylesRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}

func ptrTo[T any](val T) *T {
	return &val
}
