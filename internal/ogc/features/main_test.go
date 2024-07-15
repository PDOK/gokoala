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

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/go-chi/chi/v5"
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

func TestFeatures(t *testing.T) {
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
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items",
				collectionID: "foo",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_foo_collection.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request GeoJSON for 'foo' collection using limit of 2",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=2",
				collectionID: "foo",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_foo_collection_with_limit.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request GeoJSON for 'foo' collection using limit of 2 and cursor to next page",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/tunneldelen/items?cursor=Dv4%7CNwyr1Q&limit=2",
				collectionID: "foo",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_foo_collection_with_cursor.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request non existing feature collection",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
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
			name: "Request unsupported format (DOCX)",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				collectionID: "foo",
				format:       "docx",
			},
			want: want{
				body:       "",
				statusCode: http.StatusNotAcceptable,
			},
		},
		{
			name: "Request with unknown query params",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
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
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
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
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
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
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=1",
				collectionID: "foo",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_foo_collection_snippet.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output with property filter 'straatnaam' set to 'Silodam'",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?straatnaam=Silodam",
				collectionID: "foo",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_straatnaam_silodam.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request HTML output with property filter (validate 2 form fields present, with only straatnaam filled)",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?straatnaam=Silodam",
				collectionID: "foo",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_straatnaam_silodam.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output with two property filters set (straatnaam and postcode)'",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?straatnaam=Zandhoek&postcode=1104MM",
				collectionID: "foo",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_straatnaam_and_postcode.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request empty feature collection (zero results)'",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?straatnaam=doesnotexist",
				collectionID: "foo",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_empty_feature_collection.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output with property filters with allowed values restriction, using allowed 'straatname' value",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag_allowed_values.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?straatnaam=Silodam",
				collectionID: "foo",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_straatnaam_silodam.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output with property filters with allowed values restriction, using not allowed 'straatnaam' value",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag_allowed_values.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?straatnaam=StreetNotInAllowedValues",
				collectionID: "foo",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_straatnaam_not_allowed_value.json",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Request output with property filters with allowed values restriction, using allowed 'type' value",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag_allowed_values.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?type=Ligplaats&straatnaam=Westerdok&limit=3",
				collectionID: "foo",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_type_ligplaats.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in WGS84 explicitly",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?crs=http://www.opengis.net/def/crs/OGC/1.3/CRS84&limit=2",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_wgs84.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in RD explicitly",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&limit=2",
				collectionID: "dutch-addresses",
				contentCrs:   "<http://www.opengis.net/def/crs/EPSG/0/28992>",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_rd.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in default (WGS84) and bbox in default (WGS84)",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=4.86958187578342017%2C53.07965667574639212%2C4.88167082216529113%2C53.09197323827352477&cursor=Wl8%7C9YRHSw&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_bbox_wgs84.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in default (WGS84) and bbox in default (WGS84) in JSON-FG",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=4.86958187578342017%2C53.07965667574639212%2C4.88167082216529113%2C53.09197323827352477&cursor=Wl8%7C9YRHSw&f=jsonfg&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_bbox_wgs84_jsonfg.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in RD and bbox in default (WGS84)",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=4.86%2C53.07%2C4.88%2C53.09&crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<http://www.opengis.net/def/crs/EPSG/0/28992>",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_bbox_wgs84_output_rd.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in default (WGS84) and bbox in RD",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=120379.69%2C566718.72%2C120396.30%2C566734.62&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_bbox_rd.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in default (WGS84) and bbox in RD, with GeoPackages configured on different levels (top-level and collection-level)",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs_multiple_levels.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=120379.69%2C566718.72%2C120396.30%2C566734.62&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_bbox_rd.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in default (WGS84) and bbox in RD, with format JSON-FG",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=120379.69%2C566718.72%2C120396.30%2C566734.62&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=jsonfg&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_bbox_rd_jsonfg.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in RD and bbox in RD",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=120379.69%2C566718.72%2C120396.30%2C566734.62&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<http://www.opengis.net/def/crs/EPSG/0/28992>",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_bbox_rd_output_also_rd.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in RD and bbox in RD, with format JSON-FG",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=120379.69%2C566718.72%2C120396.30%2C566734.62&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=jsonfg&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<http://www.opengis.net/def/crs/EPSG/0/28992>",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_bbox_rd_output_also_rd_jsonfg.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in default (WGS84) and bbox explicitly in WGS84",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=4.86%2C53.07%2C4.88%2C53.09&bbox-crs=http://www.opengis.net/def/crs/OGC/1.3/CRS84&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_bbox_explicit_wgs84.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in default (WGS84) and bbox explicitly in WGS84 - with JSON response validation disabled",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_validation_disabled.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=4.86%2C53.07%2C4.88%2C53.09&bbox-crs=http://www.opengis.net/def/crs/OGC/1.3/CRS84&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_bbox_explicit_wgs84.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request WGS84 for collections with same backing feature table",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_collection_single_table.yaml",
				url:          "http://localhost:8080/collections/dutch-addresses/items?bbox=4.86%2C53.07%2C4.88%2C53.09&bbox-crs=http://www.opengis.net/def/crs/OGC/1.3/CRS84&f=json&limit=10",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_bbox_explicit_wgs84.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request temporal collection",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag_temporal.yaml",
				url:          "http://localhost:8080/collections/standplaatsen/items?datetime=2020-05-20T00:00:00Z&limit=10",
				collectionID: "standplaatsen",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_temporal.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request mapsheets as JSON",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_mapsheets.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=2",
				collectionID: "example_mapsheets",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_mapsheets.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request mapsheets as JSON-FG",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_mapsheets.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=2&f=jsonfg",
				collectionID: "example_mapsheets",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_mapsheets_jsonfg.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request mapsheets as HTML",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_mapsheets.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=2",
				collectionID: "example_mapsheets",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_mapsheets.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request slow response, hitting query timeout",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_short_query_timeout.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_query_timeout_features.json",
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "Request slow response, hitting query timeout with different. With bbox in WGS84 and output in RD",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_short_query_timeout.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?bbox=120379.69%2C566718.72%2C120396.30%2C566734.62&crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&limit=2",
				collectionID: "dutch-addresses",
				contentCrs:   "<http://www.opengis.net/def/crs/EPSG/0/28992>",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_query_timeout_features.json",
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "Request features with relation to other feature (href based on external FID)",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_external_fid.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items",
				collectionID: "standplaatsen",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_with_rel_as_link.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request features with relation to other feature (href based on external FID) as HTML hyperlink",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_external_fid.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items",
				collectionID: "standplaatsen",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_with_rel_as_link_snippet.html",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock time
			now = func() time.Time { return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) }
			engine.Now = now

			req, err := createRequest(tt.fields.url, tt.fields.collectionID, "", tt.fields.format)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
			features := NewFeatures(newEngine)
			handler := features.Features()
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

// Run the benchmark with the following command:
//
//	go test -bench=BenchmarkFeatures -run=^# -benchmem -count=10 > bench1_run1.txt
//
// Install "benchstat": go install golang.org/x/perf/cmd/benchstat@latest
// Now compare the results for each benchmark before and after making a change, e.g:
//
//	benchstat bench1_run1.txt bench1_run2.txt
//
// This will summarize the difference in performance between the runs.
// To profile CPU and Memory usage run as:
//
//	go test -bench=BenchmarkFeatures -run=^# -benchmem -count=10 -cpuprofile cpu.pprof -memprofile mem.pprof
//
// Now analyse the pprof files using:
//
//	go tool pprof -web cpu.pprof
//	go tool pprof -web mem.pprof
//
// ----
func BenchmarkFeatures(b *testing.B) {
	type fields struct {
		configFile string
		url        string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "1", // output is WGS84 json, input is WGS84 bbox
			fields: fields{
				configFile: "internal/ogc/features/testdata/config_benchmark.yaml",
				url:        "http://localhost:8080/collections/dutch-addresses/items?bbox=4.651476%2C52.962408%2C4.979398%2C53.074282&f=json&limit=1000",
			},
		},
		{
			name: "2", // same as benchmark 1 above, but now the next page
			fields: fields{
				configFile: "internal/ogc/features/testdata/config_benchmark.yaml",
				url:        "http://localhost:8080/collections/dutch-addresses/items?bbox=4.651476%2C52.962408%2C4.979398%2C53.074282&cursor=Cpc%7CwXkQbQ&f=json&limit=1000",
			},
		},
		{
			name: "3", // output is WGS84 json, input is RD bbox
			fields: fields{
				configFile: "internal/ogc/features/testdata/config_benchmark.yaml",
				url:        "http://localhost:8080/collections/dutch-addresses/items?bbox=105564.79055389616405591%2C553072.85584054281935096%2C127668.63754775881534442%2C565347.87356295716017485&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=json&limit=1000",
			},
		},
		{
			name: "4", // same as benchmark 3 above, but now the next page
			fields: fields{
				configFile: "internal/ogc/features/testdata/config_benchmark.yaml",
				url:        "http://localhost:8080/collections/dutch-addresses/items?bbox=105564.79055389616405591%2C553072.85584054281935096%2C127668.63754775881534442%2C565347.87356295716017485&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&cursor=Cyo%7CiLD6Iw&f=json&limit=1000",
			},
		},
	}
	for _, tt := range tests {
		req, err := createRequest(tt.fields.url, "dutch-addresses", "", "json")
		if err != nil {
			log.Fatal(err)
		}
		rr, ts := createMockServer()

		newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
		assert.NoError(b, err)
		features := NewFeatures(newEngine)
		handler := features.Features()

		// Start benchmark
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				handler.ServeHTTP(rr, req)

				assert.Equal(b, 200, rr.Code)
			}
		})

		ts.Close()
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
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "foo",
				featureID:    "4030",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_feature_4030.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request JSON-FG for feature 4030",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId?f=jsonfg",
				collectionID: "foo",
				featureID:    "4030",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_feature_4030_jsonfg.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request non existing feature",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
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
			name: "Request non existing collection",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "does-not-exist",
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
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
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
			name: "Request with invalid FID",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "foo",
				featureID:    "thisisnotaUUIDorINTEGER",
				format:       "json",
			},
			want: want{
				body:       "",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Request unsupported format (DOCX) for feature 4030",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "foo",
				featureID:    "4030",
				format:       "docx",
			},
			want: want{
				body:       "",
				statusCode: http.StatusNotAcceptable,
			},
		},
		{
			name: "Request HTML for feature 4030",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "foo",
				featureID:    "4030",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_feature_4030.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in WGS84 explicitly",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId?crs=http://www.opengis.net/def/crs/OGC/1.3/CRS84",
				collectionID: "dutch-addresses",
				featureID:    "b29c12b1-21a9-5e63-83b4-0ff9122ef80f",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_feature_b29c12b1_wgs84.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in WGS84 explicitly - with validation disabled",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_validation_disabled.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId?crs=http://www.opengis.net/def/crs/OGC/1.3/CRS84",
				collectionID: "dutch-addresses",
				featureID:    "b29c12b1-21a9-5e63-83b4-0ff9122ef80f",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_feature_b29c12b1_wgs84.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in RD",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId?crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992",
				collectionID: "dutch-addresses",
				featureID:    "b29c12b1-21a9-5e63-83b4-0ff9122ef80f",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_multiple_gpkgs_feature_b29c12b1_rd.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request output in unsupported CRS explicitly",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId?crs=http://www.opengis.net/def/crs/OGC/EPSG:3812131313131314141",
				collectionID: "dutch-addresses",
				featureID:    "b29c12b1-21a9-5e63-83b4-0ff9122ef80f",
				format:       "json",
			},
			want: want{
				body:       "",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Request non existing feature",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_multiple_gpkgs.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "dutch-addresses",
				featureID:    "999999",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_feature_404.json",
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "Request slow response, hitting query timeout",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_short_query_timeout.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "dutch-addresses",
				featureID:    "4030",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_query_timeout_feature.json",
				statusCode: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock time
			engine.Now = func() time.Time { return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) }

			req, err := createRequest(tt.fields.url, tt.fields.collectionID, tt.fields.featureID, tt.fields.format)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "", false, true)
			assert.NoError(t, err)
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
	url = strings.ReplaceAll(url, ":collectionId", collectionID)
	url = strings.ReplaceAll(url, ":featureId", featureID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if req == nil || err != nil {
		return req, err
	}
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
	log.Print("\n==> ACTUAL JSON RESPONSE (copy/paste and compare with response in file):")
	log.Print(rr.Body.String()) // to ease debugging & updating expected results
	log.Print("\n=========\n")
}
