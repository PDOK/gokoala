package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/PDOK/gokoala/internal/engine/types"
	"github.com/stretchr/testify/assert"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

var (
	mockPoint           = geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{1.0, 2.0})
	mockPointGeoJSON, _ = geojson.Encode(mockPoint)
)

func mockMapGeom(data any) (geom.T, error) {
	dataBytes, ok := data.([]byte)
	if !ok {
		assert.Fail(nil, "expected data to be []byte")
	}
	if string(dataBytes) == "mock error" {
		return nil, errors.New(string(dataBytes))
	}
	return mockPoint, nil
}

func TestMapColumnsToFeature(t *testing.T) {
	tests := []struct {
		name             string
		firstRow         bool
		feature          *Feature
		columns          []string
		values           []any
		fidColumn        string
		externalFidCol   string
		geomColumn       string
		schemaFields     []Field
		mapGeom          MapGeom
		expectedFeature  *Feature
		expectedPrevNext *PrevNextFID
		expectedError    error
	}{
		{
			name:             "Test FID",
			feature:          &Feature{Properties: NewFeatureProperties(false)},
			columns:          []string{"id", "name"},
			values:           []any{1, "test"},
			fidColumn:        "id",
			expectedFeature:  &Feature{ID: "1", Properties: NewFeaturePropertiesWithData(false, map[string]any{"name": "test"})},
			expectedPrevNext: &PrevNextFID{},
		},
		{
			name:             "Test Geometry valid",
			feature:          &Feature{Properties: NewFeatureProperties(false)},
			columns:          []string{"id", "geom"},
			values:           []any{1, []byte("good")},
			fidColumn:        "id",
			geomColumn:       "geom",
			mapGeom:          mockMapGeom,
			expectedFeature:  &Feature{ID: "1", Properties: NewFeaturePropertiesWithData(false, map[string]any{}), Geometry: mockPointGeoJSON},
			expectedPrevNext: &PrevNextFID{},
		},
		{
			name:          "Test Geometry invalid",
			feature:       &Feature{Properties: NewFeatureProperties(false)},
			columns:       []string{"id", "geom"},
			values:        []any{1, []byte("mock error")},
			fidColumn:     "id",
			geomColumn:    "geom",
			mapGeom:       mockMapGeom,
			expectedError: errors.New("failed to map/decode geometry from datasource, error: mock error"),
		},
		{
			name:             "Test prevfid and nextfid",
			firstRow:         true,
			feature:          &Feature{Properties: NewFeatureProperties(false)},
			columns:          []string{PrevFid, NextFid},
			values:           []any{int64(1), int64(2)},
			expectedFeature:  &Feature{Properties: NewFeatureProperties(false)},
			expectedPrevNext: &PrevNextFID{Prev: int64(1), Next: int64(2)},
		},
		{
			name:             "Test different types",
			feature:          &Feature{Properties: NewFeatureProperties(false)},
			columns:          []string{"str_col", "int_col", "float_col", "time_col", "bool_col"},
			values:           []any{"str", int64(42), 3.14, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), true},
			expectedFeature:  &Feature{Properties: NewFeaturePropertiesWithData(false, map[string]any{"str_col": "str", "int_col": int64(42), "float_col": 3.14, "time_col": time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), "bool_col": true})},
			expectedPrevNext: &PrevNextFID{},
		},
		{
			name:             "Test nil value",
			feature:          &Feature{Properties: NewFeatureProperties(false)},
			columns:          []string{"str_col", "nil_col"},
			values:           []any{"str", nil},
			expectedFeature:  &Feature{Properties: NewFeaturePropertiesWithData(false, map[string]any{"str_col": "str", "nil_col": nil})},
			expectedPrevNext: &PrevNextFID{},
		},
		{
			name:          "Test unexpected type",
			feature:       &Feature{Properties: NewFeatureProperties(false)},
			columns:       []string{"str_col", "unexpected_col"},
			values:        []any{"str", []complex128{complex(1, 2)}},
			expectedError: errors.New("unexpected type: unexpected_col: []complex128"),
		},
		{
			name:             "Test conversion of float64 with non floating point value to int64",
			feature:          &Feature{Properties: NewFeatureProperties(false)},
			columns:          []string{"float_col"},
			values:           []any{float64(376422001)},
			expectedFeature:  &Feature{Properties: NewFeaturePropertiesWithData(false, map[string]any{"float_col": int64(376422001)})},
			expectedPrevNext: &PrevNextFID{},
		},
		{
			name:    "Test date is mapped without time component",
			feature: &Feature{Properties: NewFeatureProperties(false)},
			columns: []string{"date_col", "time_col", "date_not_in_schema", "invalid_date_with_timestamp"},
			values: []any{
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),  // date_col
				time.Date(2020, 1, 1, 12, 4, 6, 8, time.UTC), // time_col
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),  // date_col_not_in_schema
				time.Date(2020, 1, 1, 12, 4, 6, 8, time.UTC), // invalid_date_with_timestamp
			},
			schemaFields: []Field{
				{
					Name: "date_col",
					Type: "date",
				},
				{
					Name: "time_col",
					Type: "datetime",
				},
				{
					Name: "date_not_in_schema",
					Type: "datetime", // marked as datetime instead of date
				},
				{
					Name: "invalid_date_with_timestamp",
					Type: "date",
				},
			},
			expectedFeature: &Feature{Properties: NewFeaturePropertiesWithData(false, map[string]any{
				"date_col":                    types.NewDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
				"time_col":                    time.Date(2020, 1, 1, 12, 4, 6, 8, time.UTC),
				"date_not_in_schema":          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				"invalid_date_with_timestamp": time.Date(2020, 1, 1, 12, 4, 6, 8, time.UTC),
			})},
			expectedPrevNext: &PrevNextFID{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := NewSchema([]Field{}, tt.fidColumn, tt.externalFidCol)
			assert.NoError(t, err)
			if tt.schemaFields != nil {
				schema.Fields = tt.schemaFields
			}
			prevNextID, err := mapColumnsToFeature(t.Context(), tt.firstRow, tt.feature, tt.columns, tt.values, tt.fidColumn, tt.externalFidCol, tt.geomColumn, schema, tt.mapGeom, nil, 0)

			if tt.expectedError != nil {
				assert.Nil(t, prevNextID)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPrevNext, prevNextID)
				assert.Equal(t, tt.expectedFeature, tt.feature)
			}
		})
	}
}
