package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-geom"
)

func TestEncodeBBox(t *testing.T) {
	type args struct {
		bbox geom.T
	}
	type want struct {
		result *[]float64
		err    error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Nil bbox",
			args: args{
				bbox: nil,
			},
			want: want{
				result: nil,
				err:    nil,
			},
		},
		{
			name: "XY layout",
			args: args{
				bbox: geom.NewPointFlat(geom.XY, []float64{10, 20}),
			},
			want: want{
				result: &[]float64{10, 20, 10, 20},
				err:    nil,
			},
		},
		{
			name: "XYM layout",
			args: args{
				bbox: geom.NewPointFlat(geom.XYM, []float64{15, 25, 0}),
			},
			want: want{
				result: &[]float64{15, 25, 15, 25},
				err:    nil,
			},
		},
		{
			name: "XYZ layout",
			args: args{
				bbox: geom.NewPointFlat(geom.XYZ, []float64{5, 10, 20}),
			},
			want: want{
				result: nil,
				err:    fmt.Errorf("unsupported type: %d", rune(geom.XYZ)),
			},
		},
		{
			name: "XYZM layout",
			args: args{
				bbox: geom.NewPointFlat(geom.XYZM, []float64{1, 2, 3, 4}),
			},
			want: want{
				result: nil,
				err:    fmt.Errorf("unsupported type: %d", rune(geom.XYZM)),
			},
		},
		{
			name: "NoLayout layout",
			args: args{
				bbox: geom.NewPointFlat(geom.NoLayout, nil),
			},
			want: want{
				result: nil,
				err:    fmt.Errorf("unsupported type: %d", rune(geom.NoLayout)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeBBox(tt.args.bbox)
			if tt.want.err != nil {
				require.EqualError(t, err, tt.want.err.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want.result, got)
		})
	}
}
