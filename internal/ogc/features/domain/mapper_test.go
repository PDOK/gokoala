package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/geojson"
	"github.com/stretchr/testify/assert"
)

func mockMapGeom(data []byte) (geom.Geometry, error) {
	if string(data) == "mock error" {
		return nil, errors.New(string(data))
	}
	return geom.Point{1.0, 2.0}, nil
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
			expectedFeature:  &Feature{ID: "1", Properties: NewFeaturePropertiesWithData(false, map[string]any{}), Geometry: geojson.Geometry{Geometry: geom.Point{1.0, 2.0}}},
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
			expectedError: errors.New("unexpected type for sqlite column data: unexpected_col: []complex128"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevNextID, err := mapColumnsToFeature(context.Background(), tt.firstRow, tt.feature, tt.columns, tt.values, tt.fidColumn, tt.externalFidCol, tt.geomColumn, tt.mapGeom, nil)

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
