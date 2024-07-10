package domain

import (
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
			firstRow:         false,
			feature:          &Feature{Feature: geojson.Feature{Properties: make(map[string]any)}},
			columns:          []string{"id", "name"},
			values:           []any{1, "test"},
			fidColumn:        "id",
			externalFidCol:   "",
			geomColumn:       "",
			mapGeom:          nil,
			expectedFeature:  &Feature{ID: "1", Feature: geojson.Feature{Properties: map[string]any{"name": "test"}}},
			expectedPrevNext: &PrevNextFID{},
			expectedError:    nil,
		},
		{
			name:             "Test Geometry valid",
			firstRow:         false,
			feature:          &Feature{Feature: geojson.Feature{Properties: make(map[string]any)}},
			columns:          []string{"id", "geom"},
			values:           []any{1, []byte("good")},
			fidColumn:        "id",
			externalFidCol:   "",
			geomColumn:       "geom",
			mapGeom:          mockMapGeom,
			expectedFeature:  &Feature{ID: "1", Feature: geojson.Feature{Properties: map[string]any{}, Geometry: geojson.Geometry{Geometry: geom.Point{1.0, 2.0}}}},
			expectedPrevNext: &PrevNextFID{},
			expectedError:    nil,
		},
		{
			name:             "Test Geometry invalid",
			firstRow:         false,
			feature:          &Feature{Feature: geojson.Feature{Properties: make(map[string]any)}},
			columns:          []string{"id", "geom"},
			values:           []any{1, []byte("mock error")},
			fidColumn:        "id",
			externalFidCol:   "",
			geomColumn:       "geom",
			mapGeom:          mockMapGeom,
			expectedFeature:  nil,
			expectedPrevNext: nil,
			expectedError:    errors.New("failed to map/decode geometry from datasource, error: mock error"),
		},
		{
			name:             "Test prevfid and nextfid",
			firstRow:         true,
			feature:          &Feature{Feature: geojson.Feature{Properties: make(map[string]any)}},
			columns:          []string{"prevfid", "nextfid"},
			values:           []any{int64(1), int64(2)},
			fidColumn:        "",
			externalFidCol:   "",
			geomColumn:       "",
			mapGeom:          nil,
			expectedFeature:  &Feature{Feature: geojson.Feature{Properties: make(map[string]any)}},
			expectedPrevNext: &PrevNextFID{Prev: int64(1), Next: int64(2)},
			expectedError:    nil,
		},
		{
			name:             "Test different types",
			firstRow:         false,
			feature:          &Feature{Feature: geojson.Feature{Properties: make(map[string]any)}},
			columns:          []string{"str_col", "int_col", "float_col", "time_col", "bool_col"},
			values:           []any{"str", int64(42), 3.14, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), true},
			fidColumn:        "",
			externalFidCol:   "",
			geomColumn:       "",
			mapGeom:          nil,
			expectedFeature:  &Feature{Feature: geojson.Feature{Properties: map[string]any{"str_col": "str", "int_col": int64(42), "float_col": 3.14, "time_col": time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), "bool_col": true}}},
			expectedPrevNext: &PrevNextFID{},
			expectedError:    nil,
		},
		{
			name:             "Test nil value",
			firstRow:         false,
			feature:          &Feature{Feature: geojson.Feature{Properties: make(map[string]any)}},
			columns:          []string{"str_col", "nil_col"},
			values:           []any{"str", nil},
			fidColumn:        "",
			externalFidCol:   "",
			geomColumn:       "",
			mapGeom:          nil,
			expectedFeature:  &Feature{Feature: geojson.Feature{Properties: map[string]any{"str_col": "str", "nil_col": nil}}},
			expectedPrevNext: &PrevNextFID{},
			expectedError:    nil,
		},
		{
			name:             "Test unexpected yype",
			firstRow:         false,
			feature:          &Feature{Feature: geojson.Feature{Properties: make(map[string]any)}},
			columns:          []string{"str_col", "unexpected_col"},
			values:           []any{"str", []complex128{complex(1, 2)}},
			fidColumn:        "",
			externalFidCol:   "",
			geomColumn:       "",
			mapGeom:          nil,
			expectedFeature:  nil,
			expectedPrevNext: nil,
			expectedError:    errors.New("unexpected type for sqlite column data: unexpected_col: []complex128"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevNextID, err := mapColumnsToFeature(tt.firstRow, tt.feature, tt.columns, tt.values,
				tt.fidColumn, tt.externalFidCol, tt.geomColumn, tt.mapGeom, nil)

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
