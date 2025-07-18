package features

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestSchema(t *testing.T) {
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
			name: "Request schema in HTML format",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/geopackage/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/schema",
				collectionID: "foo",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_schema_snippet.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request schema in JSON format",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/geopackage/config_features_bag.yaml",
				url:          "http://localhost:8080/collections/:collectionId/schema",
				collectionID: "foo",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_schema.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request schema in HTML format with temporal fields",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/geopackage/config_features_bag_temporal.yaml",
				url:          "http://localhost:8080/collections/:collectionId/schema",
				collectionID: "standplaatsen",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_schema_temporal_snippet.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request schema in JSON format with temporal fields",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/geopackage/config_features_bag_temporal.yaml",
				url:          "http://localhost:8080/collections/:collectionId/schema",
				collectionID: "standplaatsen",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_schema_temporal.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request schema in HTML format with external FID",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/geopackage/config_features_external_fid.yaml",
				url:          "http://localhost:8080/collections/:collectionId/schema",
				collectionID: "ligplaatsen",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_schema_external_fid_snippet.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request schema in JSON format with external FID",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/geopackage/config_features_external_fid.yaml",
				url:          "http://localhost:8080/collections/:collectionId/schema",
				collectionID: "ligplaatsen",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_schema_external_fid.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request schema in JSON format with external FID for collection with x-ogc-role=reference",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/geopackage/config_features_external_fid.yaml",
				url:          "http://localhost:8080/collections/:collectionId/schema",
				collectionID: "standplaatsen",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_schema_external_fid_reference.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request schema in JSON format with 3D geoms",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/geopackage/config_features_3d_geoms.yaml",
				url:          "http://localhost:8080/collections/:collectionId/schema",
				collectionID: "foo",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_schema_3d_geoms.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request schema in HTML format with descriptions from gpkg_data_columns table",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/geopackage/config_features_bag_schema_extension.yaml",
				url:          "http://localhost:8080/collections/:collectionId/schema",
				collectionID: "foo",
				format:       "html",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_schema_descr_from_db_snippet.html",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Request schema in JSON format with descriptions from gpkg_data_columns table",
			fields: fields{
				configFile:   "internal/ogc/features/testdata/geopackage/config_features_bag_schema_extension.yaml",
				url:          "http://localhost:8080/collections/:collectionId/schema",
				collectionID: "foo",
				format:       "json",
			},
			want: want{
				body:       "internal/ogc/features/testdata/expected_schema_descr_from_db.json",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createRequest(tt.fields.url, tt.fields.collectionID, "", tt.fields.format)
			assert.NoError(t, err)
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "internal/engine/testdata/test_theme.yaml", "", false, true)
			assert.NoError(t, err)
			features := NewFeatures(newEngine)
			handler := features.Schema()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			if tt.want.body != "" {
				expectedBody, err := os.ReadFile(tt.want.body)
				if err != nil {
					assert.Fail(t, "failed to read expected body", err)
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
