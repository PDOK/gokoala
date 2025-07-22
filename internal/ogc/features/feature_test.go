package features

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestFeature(t *testing.T) {
	type fields struct {
		configFiles  []string
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
				configFiles: []string{
					"internal/ogc/features/testdata/geopackage/config_features_bag.yaml",
					"internal/ogc/features/testdata/postgresql/config_features_bag.yaml",
				},
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
				configFiles: []string{
					"internal/ogc/features/testdata/geopackage/config_features_bag.yaml",
					"internal/ogc/features/testdata/postgresql/config_features_bag.yaml",
				},
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
				configFiles: []string{
					"internal/ogc/features/testdata/geopackage/config_features_bag.yaml",
					"internal/ogc/features/testdata/postgresql/config_features_bag.yaml",
				},
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
				configFiles: []string{
					"internal/ogc/features/testdata/geopackage/config_features_bag.yaml",
					"internal/ogc/features/testdata/postgresql/config_features_bag.yaml",
				},
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
				configFiles: []string{
					"internal/ogc/features/testdata/geopackage/config_features_bag.yaml",
					"internal/ogc/features/testdata/postgresql/config_features_bag.yaml",
				},
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
				configFiles: []string{
					"internal/ogc/features/testdata/geopackage/config_features_bag.yaml",
					"internal/ogc/features/testdata/postgresql/config_features_bag.yaml",
				},
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
				configFiles: []string{
					"internal/ogc/features/testdata/geopackage/config_features_bag.yaml",
					"internal/ogc/features/testdata/postgresql/config_features_bag.yaml",
				},
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "foo",
				featureID:    "4030",
				format:       "docx",
			},
			want: want{
				body:       "",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Request HTML for feature 4030",
			fields: fields{
				configFiles: []string{
					"internal/ogc/features/testdata/geopackage/config_features_bag.yaml",
					"internal/ogc/features/testdata/postgresql/config_features_bag.yaml",
				},
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
				configFiles: []string{
					"internal/ogc/features/testdata/geopackage/config_features_multiple_gpkgs.yaml",
					"internal/ogc/features/testdata/postgresql/config_features_multiple_projections.yaml",
				},
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
				configFiles:  []string{"internal/ogc/features/testdata/geopackage/config_features_validation_disabled.yaml"},
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
				configFiles: []string{
					"internal/ogc/features/testdata/geopackage/config_features_multiple_gpkgs.yaml",
					"internal/ogc/features/testdata/postgresql/config_features_multiple_projections.yaml",
				},
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
				configFiles: []string{
					"internal/ogc/features/testdata/geopackage/config_features_multiple_gpkgs.yaml",
					"internal/ogc/features/testdata/postgresql/config_features_multiple_projections.yaml",
				},
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
				configFiles: []string{
					"internal/ogc/features/testdata/geopackage/config_features_multiple_gpkgs.yaml",
					"internal/ogc/features/testdata/postgresql/config_features_multiple_projections.yaml",
				},
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
				configFiles:  []string{"internal/ogc/features/testdata/geopackage/config_features_short_query_timeout.yaml"},
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
		{
			name: "Request feature with specific web/viewer configuration, and make sure this is reflected in the HTML output",
			fields: fields{
				configFiles:  []string{"internal/ogc/features/testdata/geopackage/config_features_webconfig.yaml"},
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId?f=html",
				collectionID: "ligplaatsen",
				featureID:    "4030",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_feature_webconfig_snippet.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request feature with specific web configuration, and make sure URLs are rendered as hyperlinks in HTML output",
			fields: fields{
				configFiles:  []string{"internal/ogc/features/testdata/geopackage/config_features_webconfig.yaml"},
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId?f=html",
				collectionID: "ligplaatsen",
				featureID:    "4030",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_webconfig_hyperlink_snippet.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request GeoJSON for feature with null geom",
			fields: fields{
				configFiles:  []string{"internal/ogc/features/testdata/geopackage/config_features_geom_null_empty.yaml"},
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "foo",
				featureID:    "6436",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_feature_geom_null.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request JSON-FG for feature with null geom",
			fields: fields{
				configFiles:  []string{"internal/ogc/features/testdata/geopackage/config_features_geom_null_empty.yaml"},
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId?f=jsonfg",
				collectionID: "foo",
				featureID:    "6436",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_feature_geom_null_jsonfg.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request GeoJSON for feature with empty point",
			fields: fields{
				configFiles:  []string{"internal/ogc/features/testdata/geopackage/config_features_geom_null_empty.yaml"},
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId",
				collectionID: "foo",
				featureID:    "3542",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_feature_geom_empty_point.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request JSON-FG for feature with empty point",
			fields: fields{
				configFiles:  []string{"internal/ogc/features/testdata/geopackage/config_features_geom_null_empty.yaml"},
				url:          "http://localhost:8080/collections/:collectionId/items/:featureId?f=jsonfg",
				collectionID: "foo",
				featureID:    "3542",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_feature_geom_empty_point_jsonfg.json",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, configFile := range tt.fields.configFiles {
				dir := filepath.Dir(configFile)
				datasourceName := filepath.Base(dir)

				// nested subtest for each config-file/datasource
				// tip: in JetBrains IDEs you can still jump to failed tests by explicitly selecting "jump to source"
				t.Run(datasourceName, func(t *testing.T) {
					// mock time
					now = func() time.Time { return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) }
					engine.Now = now

					req, err := createRequest(tt.fields.url, tt.fields.collectionID, tt.fields.featureID, tt.fields.format)
					assert.NoError(t, err)
					rr, ts := createMockServer()
					defer ts.Close()

					newEngine, err := engine.NewEngine(configFile, "internal/engine/testdata/test_theme.yaml", "", false, true)
					assert.NoError(t, err)

					// use fixed decimal limit in coordinates and UTC timezone across all tests for
					// stable output between different data sources (postgres, geopackage, etc)
					newEngine.Config.OgcAPI.Features.MaxDecimals = 5
					newEngine.Config.OgcAPI.Features.ForceUTC = true

					features := NewFeatures(newEngine)
					handler := features.Feature()
					handler.ServeHTTP(rr, req)

					assert.Equal(t, tt.want.statusCode, rr.Code)
					if tt.want.body != "" {
						expectedBody, err := os.ReadFile(tt.want.body)
						assert.NoError(t, err)

						printActual(rr)
						switch {
						case tt.fields.format == engine.FormatJSON:
							assert.JSONEq(t, string(expectedBody), rr.Body.String())
						case tt.fields.format == engine.FormatHTML:
							assert.Contains(t, normalize(rr.Body.String()), normalize(string(expectedBody)))
						default:
							assert.Fail(t, "implement support to test format: %s", tt.fields.format)
						}
					}
				})
			}
		})
	}
}
