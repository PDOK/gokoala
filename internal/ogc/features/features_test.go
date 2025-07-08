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
				collectionID: "foo",
				format:       "docx",
			},
			want: want{
				body:       "",
				statusCode: http.StatusBadRequest,
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
		{
			name: "Request features with relation to other feature (URL based on external FID)",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_external_fid.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?profile=rel-as-uri",
				collectionID: "standplaatsen",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_with_rel_as_uri.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request features with relation to other feature (ID/key based on external FID)",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_external_fid.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?profile=rel-as-key",
				collectionID: "standplaatsen",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_with_rel_as_key.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request features for collection with specific viewer configuration, to make sure this is reflected in the HTML output",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_webconfig.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?f=html",
				collectionID: "ligplaatsen",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_webconfig_snippet.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request features for collection with specific web configuration, and make sure URLs are rendered as hyperlinks in HTML output",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_webconfig.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?f=html",
				collectionID: "ligplaatsen",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_webconfig_hyperlink_snippet.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request features where certain properties are excluded",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_properties_exclude.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?f=json",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_properties_exclude.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request features where properties are in a specific order",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_properties_order.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?f=json",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_properties_order.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request features where properties are in a specific order and certain properties are excluded",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_properties_order_exclude.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?f=json",
				collectionID: "dutch-addresses",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_properties_order_exclude.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request features of collection with a long description",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_bag_long_description.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=1",
				collectionID: "bar",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_bar_collection_snippet.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request 3D geoms (LINESTRING Z) as features",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_3d_geoms.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=5",
				collectionID: "foo",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_3d_geoms.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request 3D geoms (LINESTRING Z) as features as JSON-FG",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_3d_geoms.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=5&f=jsonfg",
				collectionID: "foo",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_3d_geoms_jsonfg.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request 3D geoms (MULTIPOINT Z) as features",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_3d_geoms.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=5",
				collectionID: "bar",
				contentCrs:   "<" + domain.WGS84CrsURI + ">", // Geoms are actually in RD in gpkg, but not important for this test
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_3d_geoms_multipoint.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request 3D geoms (MULTIPOINT Z) as features as JSON-FG",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_3d_geoms.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=5&f=jsonfg",
				collectionID: "bar",
				contentCrs:   "<" + domain.WGS84CrsURI + ">", // Geoms are actually in RD in gpkg, but not important for this test
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_3d_geoms_multipoint_jsonfg.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request road polygons as features",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_roads.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=10",
				collectionID: "road",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_roads.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request road polygons as features in JSON-FG",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/config_features_roads.yaml",
				url:          "http://localhost:8080/collections/:collectionId/items?limit=10&f=jsonfg",
				collectionID: "road",
				contentCrs:   "<" + domain.WGS84CrsURI + ">",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_features_roads_jsonfg.json",
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

			newEngine, err := engine.NewEngine(tt.fields.configFile, "internal/engine/testdata/test_theme.yaml", "", false, true)
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
				case tt.fields.format == engine.FormatJSON:
					assert.JSONEq(t, string(expectedBody), rr.Body.String())
				case tt.fields.format == engine.FormatHTML:
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
	log.Print("\n===========================\n")
	log.Print("\n==> ACTUAL RESPONSE BELOW. Copy/paste and compare with response in file. " +
		"Note that In the case of HTML output we only compare a relevant snippet instead of the whole file.")
	log.Print("\n===========================\n")
	log.Print(rr.Body.String()) // to ease debugging & updating expected results
	log.Print("\n===========================\n")
}
