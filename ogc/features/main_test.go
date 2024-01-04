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
	"time"

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

func TestFeatures_CollectionContent(t *testing.T) {
	type fields struct {
		configFile   string
		url          string
		contentCrs   string
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
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items",
				collectionID: "foo",
				contentCrs:   "<" + wgs84CrsURI + ">",
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
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=2",
				collectionID: "foo",
				contentCrs:   "<" + wgs84CrsURI + ">",
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
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/tunneldelen/items?f=json&cursor=Dv58Nwyr1Q%3D%3D&limit=2",
				collectionID: "foo",
				contentCrs:   "<" + wgs84CrsURI + ">",
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
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
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
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
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
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
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
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
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
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=1",
				collectionID: "foo",
				contentCrs:   "<" + wgs84CrsURI + ">",
				format:       "html",
			},
			want: want{
				body:       "ogc/features/testdata/expected_foo_collection_snippet.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output with property filter 'straatnaam' set to 'Silodam'",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?straatnaam=Silodam",
				collectionID: "foo",
				contentCrs:   "<" + wgs84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_straatnaam_silodam.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request HTML output with property filter (validate 2 form fields present, with only straatnaam filled)",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?straatnaam=Silodam",
				collectionID: "foo",
				contentCrs:   "<" + wgs84CrsURI + ">",
				format:       "html",
			},
			want: want{
				body:       "ogc/features/testdata/expected_straatnaam_silodam.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output with two property filters set (straatnaam and postcode)'",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?straatnaam=Zandhoek&postcode=1104MM",
				collectionID: "foo",
				contentCrs:   "<" + wgs84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_straatnaam_and_postcode.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in WGS84 explicitly",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?crs=http://www.opengis.net/def/crs/OGC/1.3/CRS84&limit=2",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + wgs84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_multiple_gpkgs_wgs84.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in RD explicitly",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&limit=2",
				collectionID: "dutch-addresses",
				contentCrs:   "<http://www.opengis.net/def/crs/EPSG/0/28992>",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_multiple_gpkgs_rd.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in default (WGS84) and bbox in default (WGS84)",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=4.86958187578342017%2C53.07965667574639212%2C4.88167082216529113%2C53.09197323827352477&cursor=Wl989YRHSw%3D%3D&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + wgs84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_multiple_gpkgs_bbox_wgs84.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in RD and bbox in default (WGS84)",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=4.86%2C53.07%2C4.88%2C53.09&crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<http://www.opengis.net/def/crs/EPSG/0/28992>",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_multiple_gpkgs_bbox_wgs84_output_rd.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in default (WGS84) and bbox in RD",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=120379.69%2C566718.72%2C120396.30%2C566734.62&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + wgs84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_multiple_gpkgs_bbox_rd.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in default (WGS84) and bbox in RD, with format JSON-FG",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=120379.69%2C566718.72%2C120396.30%2C566734.62&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=jsonfg&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + wgs84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_multiple_gpkgs_bbox_rd_jsonfg.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in RD and bbox in RD",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=120379.69%2C566718.72%2C120396.30%2C566734.62&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<http://www.opengis.net/def/crs/EPSG/0/28992>",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_multiple_gpkgs_bbox_rd_output_also_rd.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in RD and bbox in RD, with format JSON-FG",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=120379.69%2C566718.72%2C120396.30%2C566734.62&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=jsonfg&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<http://www.opengis.net/def/crs/EPSG/0/28992>",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_multiple_gpkgs_bbox_rd_output_also_rd_jsonfg.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in default (WGS84) and bbox explicitly in WGS84",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=4.86%2C53.07%2C4.88%2C53.09&bbox-crs=http://www.opengis.net/def/crs/OGC/1.3/CRS84&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + wgs84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_multiple_gpkgs_bbox_explicit_wgs84.json",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock time
			now = func() time.Time { return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) }

			req, err := createRequest(tt.fields.url, tt.fields.collectionID, "", tt.fields.format)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine := engine.NewEngine(tt.fields.configFile, "", false, true)
			features := NewFeatures(newEngine)
			handler := features.CollectionContent()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.fields.contentCrs, rr.Header().Get("Content-Crs"))
			assert.Equal(t, tt.want.statusCode, rr.Code)
			if tt.want.body != "" {
				expectedBody, err := os.ReadFile(tt.want.body)
				if err != nil {
					log.Fatal(err)
				}

				printActual(rr)
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
			name: "Request GeoJSON for feature 4030",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "foo",
				featureID:    "4030",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_feature_4030.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request non existing feature",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
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
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
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
			name: "Request HTML for feature 4030",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "foo",
				featureID:    "4030",
				format:       "html",
			},
			want: want{
				body:       "ogc/features/testdata/expected_feature_4030.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in WGS84 explicitly",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId?crs=http://www.opengis.net/def/crs/OGC/1.3/CRS84",
				collectionID: "dutch-addresses",
				featureID:    "10",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_multiple_gpkgs_feature_10_wgs84.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in RD",
			fields: fields{
				configFile:   "ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId?crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992",
				collectionID: "dutch-addresses",
				featureID:    "10",
				format:       "json",
			},
			want: want{
				body:       "ogc/features/testdata/expected_multiple_gpkgs_feature_10_rd.json",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createRequest(tt.fields.url, tt.fields.collectionID, tt.fields.featureID, tt.fields.format)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine := engine.NewEngine(tt.fields.configFile, "", false, true)
			features := NewFeatures(newEngine)
			handler := features.Feature()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			if tt.want.body != "" {
				expectedBody, err := os.ReadFile(tt.want.body)
				if err != nil {
					log.Fatal(err)
				}

				printActual(rr)
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
	queryString.Add(engine.FormatParam, format)
	req.URL.RawQuery = queryString.Encode()

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}

func normalize(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), ""))
}

func printActual(rr *httptest.ResponseRecorder) {
	log.Print("\n==> ACTUAL:")
	log.Print(rr.Body.String()) // to ease debugging & updating expected results
	log.Print("=========\n")
}
