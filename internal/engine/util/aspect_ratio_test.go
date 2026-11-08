package util

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-geom"
)

func TestBboxAspectRatio(t *testing.T) {
	for _, tt := range []struct {
		name   string
		bounds *geom.Bounds
		ratio  float64
	}{
		{name: "landscape", bounds: geom.NewBounds(geom.XY).Set(10, 20, 210, 120), ratio: 2},
		{name: "portrait", bounds: geom.NewBounds(geom.XY).Set(-100, -200, 0, 0), ratio: 0.5},
		{name: "square", bounds: geom.NewBounds(geom.XY).Set(0, 0, 1, 1), ratio: 1},
		{name: "nil"},
		{name: "wrong layout", bounds: geom.NewBounds(geom.XYZ).Set(0, 0, 0, 1, 1, 1)},
		{name: "empty", bounds: geom.NewBounds(geom.XY)},
		{name: "zero width", bounds: geom.NewBounds(geom.XY).Set(1, 1, 1, 2)},
		{name: "zero height", bounds: geom.NewBounds(geom.XY).Set(1, 1, 2, 1)},
		{name: "inverted", bounds: geom.NewBounds(geom.XY).Set(2, 2, 1, 1)},
		{name: "nan", bounds: geom.NewBounds(geom.XY).Set(math.NaN(), 0, 2, 1)},
		{name: "infinity", bounds: geom.NewBounds(geom.XY).Set(0, 0, math.Inf(1), 1)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ratio, err := BboxAspectRatio(tt.bounds)
			if tt.ratio == 0 {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.InDelta(t, tt.ratio, ratio, 1e-12)
		})
	}
}

func TestCompleteDimensions(t *testing.T) {
	for _, tt := range []struct {
		name                  string
		width, height         int
		ratio                 float64
		wantWidth, wantHeight int
		overflow              bool
	}{
		{name: "height from landscape", width: 800, ratio: 2, wantWidth: 800, wantHeight: 400},
		{name: "width from landscape", height: 300, ratio: 2, wantWidth: 600, wantHeight: 300},
		{name: "height from portrait", width: 300, ratio: 0.5, wantWidth: 300, wantHeight: 600},
		{name: "nearest integer", width: 100, ratio: 3, wantWidth: 100, wantHeight: 33},
		{name: "half rounds up", width: 101, ratio: 2, wantWidth: 101, wantHeight: 51},
		{name: "minimum one pixel", width: 1, ratio: 100, wantWidth: 1, wantHeight: 1},
		{name: "zero ratio", width: 100},
		{name: "negative ratio", width: 100, ratio: -1},
		{name: "nan ratio", width: 100, ratio: math.NaN()},
		{name: "infinite ratio", width: 100, ratio: math.Inf(1)},
		{name: "both missing", ratio: 1},
		{name: "both present", width: 1, height: 1, ratio: 1},
		{name: "negative dimension", width: -1, ratio: 1},
		{name: "width overflow", height: math.MaxInt, ratio: 2, overflow: true},
		{name: "height overflow", width: math.MaxInt, ratio: 0.5, overflow: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			width, height, err := CompleteDimensions(tt.width, tt.height, tt.ratio)
			if tt.wantWidth == 0 {
				require.Error(t, err)
				if tt.overflow {
					require.ErrorIs(t, err, ErrDimensionOverflow)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantWidth, width)
			require.Equal(t, tt.wantHeight, height)
		})
	}
}

func TestBboxAtResolution(t *testing.T) {
	original := geom.NewBounds(geom.XY).Set(10, 20, 210, 120)
	for _, tt := range []struct {
		width, height int
		want          []float64
	}{
		{width: 200, height: 100, want: []float64{10, 20, 210, 120}},
		{width: 100, height: 100, want: []float64{60, 20, 160, 120}},
		{width: 400, height: 200, want: []float64{-90, -30, 310, 170}},
	} {
		bounds, err := BboxAtResolution(original, tt.width, tt.height, 1)
		require.NoError(t, err)
		require.Equal(t, tt.want, []float64{bounds.Min(0), bounds.Min(1), bounds.Max(0), bounds.Max(1)})
		require.NotSame(t, original, bounds)
	}
	require.Equal(t, geom.NewBounds(geom.XY).Set(10, 20, 210, 120), original)
	for _, resolution := range []float64{0, -1, math.Inf(1), math.NaN(), math.MaxFloat64} {
		_, err := BboxAtResolution(original, 200, 100, resolution)
		require.Error(t, err)
	}
	_, err := BboxAtResolution(nil, 1, 1, 1)
	require.Error(t, err)
}
