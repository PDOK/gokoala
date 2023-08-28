package geovolumes

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

	"github.com/PDOK/gokoala/engine"

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

func TestThreeDimensionalGeoVolume_Tile(t *testing.T) {
	type fields struct {
		configFile       string
		url              string
		containerID      string
		tilePathPrefix   string
		tileMatrix       string
		tileRow          string
		tileColAndSuffix string
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
			name: "container_1/0/0/0/0",
			fields: fields{
				configFile:       "ogc/geovolumes/testdata/config_minimal_3d.yaml",
				url:              "http://localhost:8080/collections/:3dContainerId/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol",
				containerID:      "container_1",
				tilePathPrefix:   "0",
				tileMatrix:       "0",
				tileRow:          "0",
				tileColAndSuffix: "0",
			},
			want: want{
				body:       "/container_1/0/0/0/0",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "container_2/1/2/3/4",
			fields: fields{
				configFile:       "ogc/geovolumes/testdata/config_minimal_3d.yaml",
				url:              "http://localhost:8080/collections/:3dContainerId/:tileMatrixSetId/:tileMatrix/:tileRow/:tileCol",
				containerID:      "container_2",
				tilePathPrefix:   "1",
				tileMatrix:       "2",
				tileRow:          "3",
				tileColAndSuffix: "4",
			},
			want: want{
				body:       "/container_2/1/2/3/4",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createTileRequest(tt.fields.url, tt.fields.containerID, tt.fields.tilePathPrefix, tt.fields.tileMatrix, tt.fields.tileRow, tt.fields.tileColAndSuffix)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine := engine.NewEngine(tt.fields.configFile, "")
			threeDimensionalGeoVolume := NewThreeDimensionalGeoVolumes(newEngine, chi.NewRouter())
			handler := threeDimensionalGeoVolume.Tile()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Equal(t, tt.want.body, rr.Body.String())
		})
	}
}

func TestThreeDimensionalGeoVolume_CollectionContent(t *testing.T) {
	type fields struct {
		configFile  string
		url         string
		containerID string
		tileSet     string
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
			name: "container_1/tileset.json",
			fields: fields{
				configFile:  "ogc/geovolumes/testdata/config_minimal_3d.yaml",
				url:         "http://localhost:8080/collections/:3dContainerId/tileset.json",
				containerID: "container_1",
				tileSet:     "tileset.json",
			},
			want: want{
				body:       "/container_1/tileset.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "container_2/tileset.json",
			fields: fields{
				configFile:  "ogc/geovolumes/testdata/config_minimal_3d.yaml",
				url:         "http://localhost:8080/collections/:3dContainerId/tileset.json",
				containerID: "container_2",
				tileSet:     "tileset.json",
			},
			want: want{
				body:       "/container_2/tileset.json",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createTileSetRequest(tt.fields.url, tt.fields.containerID, tt.fields.tileSet)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine := engine.NewEngine(tt.fields.configFile, "")
			threeDimensionalGeoVolume := NewThreeDimensionalGeoVolumes(newEngine, chi.NewRouter())
			handler := threeDimensionalGeoVolume.CollectionContent()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Equal(t, tt.want.body, rr.Body.String())
		})
	}
}

func TestThreeDimensionalGeoVolume_ExplicitTileSet(t *testing.T) {
	type fields struct {
		configFile  string
		url         string
		containerID string
		tileSet     string
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
			name: "container_2/tileset-5-768-896.json",
			fields: fields{
				configFile:  "ogc/geovolumes/testdata/config_minimal_3d.yaml",
				url:         "http://localhost:8080/collections/:3dContainerId/:explicitTileSet.json",
				containerID: "container_2",
				tileSet:     "tileset-5-768-896",
			},
			want: want{
				body:       "/container_2/tileset-5-768-896.json",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createTileSetRequest(tt.fields.url, tt.fields.containerID, tt.fields.tileSet)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine := engine.NewEngine(tt.fields.configFile, "")
			threeDimensionalGeoVolume := NewThreeDimensionalGeoVolumes(newEngine, chi.NewRouter())
			handler := threeDimensionalGeoVolume.ExplicitTileset()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Equal(t, tt.want.body, rr.Body.String())
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

func createTileRequest(url string, containerID string, tilePathPrefix string, tileMatrix string, tileRow string, tileColAndSuffix string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()

	rctx.URLParams.Add("3dContainerId", containerID)
	rctx.URLParams.Add("tilePathPrefix", tilePathPrefix)
	rctx.URLParams.Add("tileMatrix", tileMatrix)
	rctx.URLParams.Add("tileRow", tileRow)
	rctx.URLParams.Add("tileColAndSuffix", tileColAndSuffix)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}
func createTileSetRequest(url string, containerID string, tileSet string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()

	rctx.URLParams.Add("3dContainerId", containerID)
	rctx.URLParams.Add("explicitTileSet", tileSet)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}
