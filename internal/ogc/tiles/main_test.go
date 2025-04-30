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

	"github.com/PDOK/gokoala/config"

	"github.com/PDOK/gokoala/internal/engine"
	"golang.org/x/text/language"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

var tilePath = "/foo/12/34/56.pbf"

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../")
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
				e: engine.NewEngineWithConfig(&config.Config{
					Version:            "3.3.0",
					Title:              "Test API",
					Abstract:           "Test API description",
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
											Start: 0,
											End:   6,
										},
									},
								},
								HealthCheck: config.HealthCheck{Enabled: ptrTo(true), Srs: "EPSG:28992", TilePath: &tilePath},
							},
						},
						Styles: &config.OgcAPIStyles{
							Default:         "foo",
							SupportedStyles: nil,
						},
					},
				}, "", false, true),
			},
		},
		{
			name: "Test render templates with OGC Tiles config and one SRS",
			args: args{
				e: engine.NewEngineWithConfig(&config.Config{
					Version:            "3.3.0",
					Title:              "Test API",
					Abstract:           "Test API description",
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
											Start: 0,
											End:   6,
										},
									},
								},
								HealthCheck: config.HealthCheck{Enabled: ptrTo(false), Srs: "EPSG:28992", TilePath: &tilePath},
							},
						},
						Styles: &config.OgcAPIStyles{
							Default:         "foo",
							SupportedStyles: nil,
						},
					},
				}, "", false, true),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tiles := NewTiles(test.args.e)
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
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
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
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
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
			name: "NetherlandsRDNewQuad/5/10/15?f=pbf",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol?f=pbf",
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
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol",
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
			name: "NetherlandsRDNewQuad/5/10/15.pbf",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
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
			name: "NetherlandsRDNewQuad/5/10/15?f=tilejson",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol?f=tilejson",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15",
			},
			want: want{
				body:       "specify tile format. Currently only Mapbox Vector Tiles (?f=mvt) tiles are supported",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "different uriTemplateTiles",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_urltemplate.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol?f=mvt",
				tileMatrixSetID: "EuropeanETRS89_LAEAQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15",
			},
			want: want{
				body:       "/foo/EuropeanETRS89_LAEAQuad/5/10/15",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "request unknown tileMatrixSet",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol?f=mvt",
				tileMatrixSetID: "EuropeanETRS89_LAEAQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15",
			},
			want: want{
				body:       "unknown tileMatrixSet 'EuropeanETRS89_LAEAQuad'",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "request tile in unsupported tileMatrix",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "13",
				tileRow:         "10",
				tileCol:         "15.pbf",
			},
			want: want{
				body:       "tileMatrix 13 is out of range",
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "request tile beyond tileMatrixSetLimits",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "32",
				tileCol:         "32.pbf",
			},
			want: want{
				body:       "tileRow/tileCol 32/32 is out of range",
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "invalid request parameter",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
				url:             "http://localhost:8080/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "3A",
				tileCol:         "E2.pbf",
			},
			want: want{
				body:       "invalid syntax",
				statusCode: http.StatusBadRequest,
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

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			tiles := NewTiles(newEngine)
			handler := tiles.Tile(*newEngine.Config.OgcAPI.Tiles.DatasetTiles)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.body)
		})
	}
}

func TestTiles_TileForCollection(t *testing.T) {
	type fields struct {
		configFile      string
		url             string
		tileMatrixSetID string
		tileMatrix      string
		tileRow         string
		tileCol         string
		collection      string
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
			name: "example/NetherlandsRDNewQuad/0/0/0?f=mvt",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_collectionlevel.yaml",
				url:             "http://localhost:8080/example/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol?f=mvt",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "0",
				tileRow:         "0",
				tileCol:         "0",
				collection:      "example",
			},
			want: want{
				body:       "/NetherlandsRDNewQuad/0/0/0.pbf",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "example/NetherlandsRDNewQuad/5/10/15?f=mvt",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_collectionlevel.yaml",
				url:             "http://localhost:8080/example/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol?f=mvt",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15",
				collection:      "example",
			},
			want: want{
				body:       "/NetherlandsRDNewQuad/5/15/10.pbf",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "example/NetherlandsRDNewQuad/5/10/15?f=pbf",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_collectionlevel.yaml",
				url:             "http://localhost:8080/example/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol?f=pbf",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15",
				collection:      "example",
			},
			want: want{
				body:       "/NetherlandsRDNewQuad/5/15/10.pbf",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "example/NetherlandsRDNewQuad/5/10/15",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_collectionlevel.yaml",
				url:             "http://localhost:8080/example/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15",
				collection:      "example",
			},
			want: want{
				body:       "/NetherlandsRDNewQuad/5/15/10.pbf",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "example/NetherlandsRDNewQuad/5/10/15.pbf",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_collectionlevel.yaml",
				url:             "http://localhost:8080/example/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15.pbf",
				collection:      "example",
			},
			want: want{
				body:       "/NetherlandsRDNewQuad/5/15/10.pbf",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "example/NetherlandsRDNewQuad/5/10/15?f=tilejson",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_collectionlevel.yaml",
				url:             "http://localhost:8080/example/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol?f=tilejson",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15",
				collection:      "example",
			},
			want: want{
				body:       "specify tile format. Currently only Mapbox Vector Tiles (?f=mvt) tiles are supported",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid/NetherlandsRDNewQuad/5/10/15?=pbf",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_collectionlevel.yaml",
				url:             "http://localhost:8080/invalid/tiles/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol?f=pbf",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				tileMatrix:      "5",
				tileRow:         "10",
				tileCol:         "15",
				collection:      "invalid",
			},
			want: want{
				body:       "no tiles available for collection: invalid",
				statusCode: http.StatusNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createTileRequest(tt.fields.url, tt.fields.tileMatrixSetID, tt.fields.tileMatrix, tt.fields.tileRow, tt.fields.tileCol, tt.fields.collection)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			tiles := NewTiles(newEngine)
			geoDataTiles := map[string]config.Tiles{newEngine.Config.OgcAPI.Tiles.Collections[0].ID: newEngine.Config.OgcAPI.Tiles.Collections[0].Tiles.GeoDataTiles}
			handler := tiles.TileForCollection(geoDataTiles)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.body)
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
				configFile: "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
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
				configFile: "internal/ogc/tiles/testdata/config_tiles_urltemplate.yaml",
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

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			tiles := NewTiles(newEngine)
			handler := tiles.TilesetsList()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.bodyContains)
		})
	}
}

func TestTile_TilesetsListForCollection(t *testing.T) {
	type fields struct {
		configFile string
		url        string
		collection string
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
				configFile: "internal/ogc/tiles/testdata/config_tiles_collectionlevel.yaml",
				url:        "http://localhost:8080/collections/:collection/tiles",
				collection: "example",
			},
			want: want{
				bodyContains: "NetherlandsRDNewQuad",
				statusCode:   http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createTilesetsListRequest(tt.fields.url, tt.fields.collection)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			tiles := NewTiles(newEngine)
			handler := tiles.TilesetsListForCollection()
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
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
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
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
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
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
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
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
				url:             "http://localhost:8080/tiles/Invalid",
				tileMatrixSetID: "Invalid",
			},
			want: want{
				bodyContains: "request doesn't conform to OpenAPI spec: parameter \\\"tileMatrixSetId\\\" in path has an error: value is not one of the allowed values",
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

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			tiles := NewTiles(newEngine)
			handler := tiles.Tileset()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.bodyContains)
		})
	}
}

func TestTile_TilesetForCollection(t *testing.T) {
	type fields struct {
		configFile      string
		url             string
		tileMatrixSetID string
		collection      string
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
				configFile:      "internal/ogc/tiles/testdata/config_tiles_collectionlevel.yaml",
				url:             "http://localhost:8080/collections/example/tiles/NetherlandsRDNewQuad",
				tileMatrixSetID: "NetherlandsRDNewQuad",
				collection:      "example",
			},
			want: want{
				bodyContains: "NetherlandsRDNewQuad",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "EuropeanETRS89_LAEAQuad",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_collectionlevel.yaml",
				url:             "http://localhost:8080/collections/example/tiles/EuropeanETRS89_LAEAQuad",
				tileMatrixSetID: "EuropeanETRS89_LAEAQuad",
				collection:      "example",
			},
			want: want{
				bodyContains: "EuropeanETRS89_LAEAQuad",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "WebMercatorQuad",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_collectionlevel.yaml",
				url:             "http://localhost:8080/collections/example/tiles/WebMercatorQuad",
				tileMatrixSetID: "WebMercatorQuad",
				collection:      "example",
			},
			want: want{
				bodyContains: "WebMercatorQuad",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "Invalid",
			fields: fields{
				configFile:      "internal/ogc/tiles/testdata/config_tiles_collectionlevel.yaml",
				url:             "http://localhost:8080/collections/example/tiles/Invalid",
				tileMatrixSetID: "Invalid",
				collection:      "example",
			},
			want: want{
				bodyContains: "request doesn't conform to OpenAPI spec: parameter \\\"tileMatrixSetId\\\" in path has an error: value is not one of the allowed values",
				statusCode:   http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createTilesetRequest(tt.fields.url, tt.fields.tileMatrixSetID, tt.fields.collection)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			tiles := NewTiles(newEngine)
			handler := tiles.TilesetForCollection()
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
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
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
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
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
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
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
				configFile:      "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
				url:             "http://localhost:8080/tileMatrixSets/Invalid",
				tileMatrixSetID: "Invalid",
			},
			want: want{
				bodyContains: "request doesn't conform to OpenAPI spec: parameter \\\"tileMatrixSetId\\\" in path has an error: value is not one of the allowed values",
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

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			tiles := NewTiles(newEngine)
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
				configFile: "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
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
				configFile: "internal/ogc/tiles/testdata/config_tiles_urltemplate.yaml",
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
				configFile: "internal/ogc/tiles/testdata/config_tiles_toplevel.yaml",
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

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			tiles := NewTiles(newEngine)
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

func createTileRequest(url string, tileMatrixSetID string, tileMatrix string, tileRow string, tileCol string, collectionID ...string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()
	for _, id := range collectionID {
		rctx.URLParams.Add("collectionId", id)
	}

	rctx.URLParams.Add("tileMatrixSetId", tileMatrixSetID)
	rctx.URLParams.Add("tileMatrix", tileMatrix)
	rctx.URLParams.Add("tileRow", tileRow)
	rctx.URLParams.Add("tileCol", tileCol)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}

func createTilesetsListRequest(url string, collectionID ...string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()
	for _, id := range collectionID {
		rctx.URLParams.Add("collectionId", id)
	}

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}

func createTilesetRequest(url string, tileMatrixSetID string, collectionID ...string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("tileMatrixSetId", tileMatrixSetID)
	for _, id := range collectionID {
		rctx.URLParams.Add("collectionId", id)
	}

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

func ptrTo[T any](val T) *T {
	return &val
}
