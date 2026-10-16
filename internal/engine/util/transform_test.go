package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-geom"
)

func TestTransformPoint(t *testing.T) {
	tests := []struct {
		name      string
		point     geom.Coord
		inputSRS  string
		outputSRS string
		want      geom.Coord
		tolerance float64
		wantError bool
	}{
		{
			name:  "WGS84 to Web Mercator uses longitude latitude",
			point: geom.Coord{5, 52}, inputSRS: "EPSG:4326", outputSRS: "EPSG:3857",
			want: geom.Coord{556597.453966, 6800125.454397}, tolerance: 0.001,
		},
		{
			name:  "Utrecht to RD",
			point: geom.Coord{5.05, 52.05}, inputSRS: "OGC:CRS84", outputSRS: "EPSG:28992",
			want: geom.Coord{131869.474, 451352.254}, tolerance: 2,
		},
		{
			name:  "identity",
			point: geom.Coord{5.05, 52.05}, inputSRS: "EPSG:4326", outputSRS: "EPSG:4326",
			want: geom.Coord{5.05, 52.05}, tolerance: 0.000000001,
		},
		{name: "invalid input CRS", point: geom.Coord{5, 52}, inputSRS: "invalid", outputSRS: "EPSG:28992", wantError: true},
		{name: "invalid output CRS", point: geom.Coord{5, 52}, inputSRS: "EPSG:4326", outputSRS: "invalid", wantError: true},
		{name: "missing point", inputSRS: "EPSG:4326", outputSRS: "EPSG:28992", wantError: true},
		{name: "unsupported dimension", point: geom.Coord{5, 52, 0}, inputSRS: "EPSG:4326", outputSRS: "EPSG:28992", wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := append(geom.Coord(nil), test.point...)
			got, err := TransformPoint(test.point, test.inputSRS, test.outputSRS)
			if test.wantError {
				require.Error(t, err)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.InDeltaSlice(t, test.want, got, test.tolerance)
			assert.Equal(t, original, test.point)
		})
	}
}

func TestTransformBoundingBox(t *testing.T) {
	tests := []struct {
		name      string
		bbox      *geom.Bounds
		inputSRS  string
		outputSRS string
		want      []float64
		tolerance float64
		wantError bool
	}{
		{
			name:     "Utrecht to RD",
			bbox:     geom.NewBounds(geom.XY).Set(5.05, 52.05, 5.18, 52.14),
			inputSRS: "OGC:CRS84", outputSRS: "EPSG:28992",
			want: []float64{131869.474, 451352.254, 140815.394, 461332.147}, tolerance: 0.01,
		},
		{
			name:     "identity",
			bbox:     geom.NewBounds(geom.XY).Set(5.05, 52.05, 5.18, 52.14),
			inputSRS: "EPSG:4326", outputSRS: "EPSG:4326",
			want: []float64{5.05, 52.05, 5.18, 52.14}, tolerance: 0.000000001,
		},
		{
			name: "invalid input CRS", bbox: geom.NewBounds(geom.XY).Set(5, 52, 6, 53),
			inputSRS: "invalid", outputSRS: "EPSG:28992", wantError: true,
		},
		{
			name: "invalid output CRS", bbox: geom.NewBounds(geom.XY).Set(5, 52, 6, 53),
			inputSRS: "EPSG:4326", outputSRS: "invalid", wantError: true,
		},
		{name: "nil bbox", inputSRS: "EPSG:4326", outputSRS: "EPSG:28992", wantError: true},
		{name: "empty bbox", bbox: geom.NewBounds(geom.XY), inputSRS: "EPSG:4326", outputSRS: "EPSG:28992", wantError: true},
		{
			name: "unsupported layout", bbox: geom.NewBounds(geom.XYZ).Set(5, 52, 0, 6, 53, 1),
			inputSRS: "EPSG:4326", outputSRS: "EPSG:28992", wantError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var expectedBBox []float64
			if test.bbox != nil && test.bbox.Layout() == geom.XY {
				expectedBBox = []float64{test.bbox.Min(0), test.bbox.Min(1), test.bbox.Max(0), test.bbox.Max(1)}
			}
			got, err := TransformBoundingBox(test.bbox, test.inputSRS, test.outputSRS)
			if test.wantError {
				require.Error(t, err)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, geom.XY, got.Layout())
			assert.InDeltaSlice(t, test.want, []float64{got.Min(0), got.Min(1), got.Max(0), got.Max(1)}, test.tolerance)
			assert.NotSame(t, test.bbox, got)
			assert.Equal(t, expectedBBox, []float64{test.bbox.Min(0), test.bbox.Min(1), test.bbox.Max(0), test.bbox.Max(1)})
		})
	}
}
