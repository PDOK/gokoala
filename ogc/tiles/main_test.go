package tiles

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

	"github.com/PDOK/gokoala/engine"
	"golang.org/x/text/language"

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

func TestNewTiles(t *testing.T) {
	type args struct {
		e *engine.Engine
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test render templates with OGC Tiles config",
			args: args{
				e: engine.NewEngineWithConfig(&engine.Config{
					Version:            "3.3.0",
					Title:              "Test API",
					Abstract:           "Test API description",
					AvailableLanguages: []language.Tag{language.Dutch},
					BaseURL:            engine.YAMLURL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: engine.OgcAPI{
						GeoVolumes: nil,
						Tiles: &engine.OgcAPITiles{
							TileServer: engine.YAMLURL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset"}},
							Types:      []string{"vector"},
							SupportedSrs: []engine.SupportedSrs{
								{
									Srs: "EPSG:28992",
									ZoomLevelRange: engine.ZoomLevelRange{
										Start: 0,
										End:   6,
									},
								},
							},
						},
						Styles: &engine.OgcAPIStyles{
							Default:         "foo",
							SupportedStyles: nil,
						},
					},
				}, ""),
			},
		},
		{
			name: "Test render templates with OGC Tiles config and one SRS",
			args: args{
				e: engine.NewEngineWithConfig(&engine.Config{
					Version:            "3.3.0",
					Title:              "Test API",
					Abstract:           "Test API description",
					AvailableLanguages: []language.Tag{language.Dutch},
					BaseURL:            engine.YAMLURL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: engine.OgcAPI{
						GeoVolumes: nil,
						Tiles: &engine.OgcAPITiles{
							TileServer: engine.YAMLURL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset"}},
							Types:      []string{"vector"},
							SupportedSrs: []engine.SupportedSrs{
								{
									Srs: "EPSG:28992",
									ZoomLevelRange: engine.ZoomLevelRange{
										Start: 0,
										End:   6,
									},
								},
							},
						},
						Styles: &engine.OgcAPIStyles{
							Default:         "foo",
							SupportedStyles: nil,
						},
					},
				}, ""),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tiles := NewTiles(test.args.e, chi.NewRouter())
			assert.NotEmpty(t, tiles.engine.Templates.RenderedTemplates)
		})
	}
}

func TestTiles_Tile(t *testing.T) {
	type fields struct {
		configFile      string
		url             string
		tileMatrixSetID string
		tileMatrix      string
		tileRow         string
		tileCol         string
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
			name: "NetherlandsRDNewQuad/0/0/0?f=mvt",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol?f=mvt",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "0",
				tileRow:         "0",
				tileCol:         "0",
			},
			want: want{
				body:       "/NetherlandsRDNewQuad/0/0/0.pbf",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "NetherlandsRDNewQuad/5/10/15?f=mvt",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol?f=mvt",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15",
			},
			want: want{
				body:       "/NetherlandsRDNewQuad/5/15/10.pbf",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "NetherlandsRDNewQuad/5/10/15",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15",
			},
			want: want{
				body:       "Specify tile format. Currently only Mapbox Vector Tiles (?f=mvt) tiles are supported\n",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "NetherlandsRDNewQuad/5/10/15.pbf",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15.pbf",
			},
			want: want{
				body:       "/NetherlandsRDNewQuad/5/15/10.pbf",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "different uriTemplateTiles",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles_2.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol?f=mvt",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15",
			},
			want: want{
				body:       "/foo/NetherlandsRDNewQuad/5/10/15",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createTileRequest(tt.fields.url, tt.fields.tileMatrixSetID, tt.fields.tileMatrix, tt.fields.tileRow, tt.fields.tileCol)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine := engine.NewEngine(tt.fields.configFile, "")
			tiles := NewTiles(newEngine, chi.NewRouter())
			handler := tiles.Tile()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Equal(t, tt.want.body, rr.Body.String())
		})
	}
}

func TestTile_TilesetsList(t *testing.T) {
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
				configFile: "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:        "http://localhost:8080/tiles",
			},
			want: want{
				bodyContains: "NetherlandsRDNewQuad",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "test EuropeanETRS89_LAEAQuad present",
			fields: fields{
				configFile: "ogc/tiles/testdata/config_minimal_tiles_2.yaml",
				url:        "http://localhost:8080/tiles",
			},
			want: want{
				bodyContains: "EuropeanETRS89_LAEAQuad",
				statusCode:   http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createTilesetsListRequest(tt.fields.url)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine := engine.NewEngine(tt.fields.configFile, "")
			tiles := NewTiles(newEngine, chi.NewRouter())
			handler := tiles.TilesetsList()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.bodyContains)
		})
	}
}

func TestTile_Tileset(t *testing.T) {
	type fields struct {
		configFile      string
		url             string
		tileMatrixSetID string
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
			name: "NetherlandsRDNewQuad",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:             "http://localhost:8080/tiles/NetherlandsRDNewQuad",
				tileMatrixSetID: "NetherlandsRDNewQuad",
			},
			want: want{
				bodyContains: "NetherlandsRDNewQuad",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "EuropeanETRS89_LAEAQuad",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:             "http://localhost:8080/tiles/EuropeanETRS89_LAEAQuad",
				tileMatrixSetID: "EuropeanETRS89_LAEAQuad",
			},
			want: want{
				bodyContains: "EuropeanETRS89_LAEAQuad",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "WebMercatorQuad",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:             "http://localhost:8080/tiles/WebMercatorQuad",
				tileMatrixSetID: "WebMercatorQuad",
			},
			want: want{
				bodyContains: "WebMercatorQuad",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "Invalid",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:             "http://localhost:8080/tiles/Invalid",
				tileMatrixSetID: "Invalid",
			},
			want: want{
				bodyContains: "request doesn't conform to OpenAPI spec: parameter \"tileMatrixSetId\" in path has an error: value is not one of the allowed values",
				statusCode:   http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createTilesetRequest(tt.fields.url, tt.fields.tileMatrixSetID)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine := engine.NewEngine(tt.fields.configFile, "")
			tiles := NewTiles(newEngine, chi.NewRouter())
			handler := tiles.Tileset()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.bodyContains)
		})
	}
}

func TestTile_TilematrixSet(t *testing.T) {
	type fields struct {
		configFile      string
		url             string
		tileMatrixSetID string
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
			name: "NetherlandsRDNewQuad",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:             "http://localhost:8080/tileMatrixSets/NetherlandsRDNewQuad",
				tileMatrixSetID: "NetherlandsRDNewQuad",
			},
			want: want{
				bodyContains: "NetherlandsRDNewQuad",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "EuropeanETRS89_LAEAQuad",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:             "http://localhost:8080/tileMatrixSets/EuropeanETRS89_LAEAQuad",
				tileMatrixSetID: "EuropeanETRS89_LAEAQuad",
			},
			want: want{
				bodyContains: "EuropeanETRS89_LAEAQuad",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "WebMercatorQuad",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:             "http://localhost:8080/tileMatrixSets/WebMercatorQuad",
				tileMatrixSetID: "WebMercatorQuad",
			},
			want: want{
				bodyContains: "WebMercatorQuad",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "Invalid",
			fields: fields{
				configFile:      "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:             "http://localhost:8080/tileMatrixSets/Invalid",
				tileMatrixSetID: "Invalid",
			},
			want: want{
				bodyContains: "request doesn't conform to OpenAPI spec: parameter \"tileMatrixSetId\" in path has an error: value is not one of the allowed values",
				statusCode:   http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createTilematrixSetRequest(tt.fields.url, tt.fields.tileMatrixSetID)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine := engine.NewEngine(tt.fields.configFile, "")
			tiles := NewTiles(newEngine, chi.NewRouter())
			handler := tiles.TileMatrixSet()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.bodyContains)
		})
	}
}

func TestTile_TilematrixSets(t *testing.T) {
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
			name: "NetherlandsRDNewQuad",
			fields: fields{
				configFile: "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:        "http://localhost:8080/tileMatrixSets",
			},
			want: want{
				bodyContains: "NetherlandsRDNewQuad",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "EuropeanETRS89_LAEAQuad",
			fields: fields{
				configFile: "ogc/tiles/testdata/config_minimal_tiles_2.yaml",
				url:        "http://localhost:8080/tileMatrixSets",
			},
			want: want{
				bodyContains: "EuropeanETRS89_LAEAQuad",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "EuropeanETRS89_LAEAQuad",
			fields: fields{
				configFile: "ogc/tiles/testdata/config_minimal_tiles.yaml",
				url:        "http://localhost:8080/tileMatrixSets",
			},
			want: want{
				bodyContains: "WebMercatorQuad",
				statusCode:   http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createTilematrixSetsRequest(tt.fields.url)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine := engine.NewEngine(tt.fields.configFile, "")
			tiles := NewTiles(newEngine, chi.NewRouter())
			handler := tiles.TileMatrixSets()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.bodyContains)
		})
	}
}

func createMockServer() (*httptest.ResponseRecorder, *httptest.Server) {
	rr := httptest.NewRecorder()
	l, err := net.Listen("tcp", "localhost:9090")
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

func createTileRequest(url string, tileMatrixSetID string, tileMatrix string, tileRow string, tileCol string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("tileMatrixSetId", tileMatrixSetID)
	rctx.URLParams.Add("tileMatrix", tileMatrix)
	rctx.URLParams.Add("tileRow", tileRow)
	rctx.URLParams.Add("tileCol", tileCol)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}

func createTilesetsListRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}

func createTilesetRequest(url string, tileMatrixSetID string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("tileMatrixSetId", tileMatrixSetID)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}

func createTilematrixSetRequest(url string, tileMatrixSetID string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("tileMatrixSetId", tileMatrixSetID)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}

func createTilematrixSetsRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}
