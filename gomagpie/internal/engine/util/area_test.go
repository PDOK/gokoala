package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/twpayne/go-geom"
)

func TestSurfaceArea(t *testing.T) {
	tests := []struct {
		name string
		bbox *geom.Bounds
		want float64
	}{
		{
			name: "Test correct bbox",
			bbox: geom.NewBounds(geom.XY).Set(0.0, 0.0, 5.0, 5.0),
			want: 25.0,
		},
		{
			name: "Test bbox with zero area",
			bbox: geom.NewBounds(geom.XY).Set(0.0, 5.0, 0.0, 5.0),
			want: 0.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SurfaceArea(tt.bbox)
			assert.Equal(t, tt.want, got) //nolint:testifylint
		})
	}
}
