package maps

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-geom"
)

func TestMapRequestSetDefaultsBBox(t *testing.T) {
	tests := []struct {
		name string
		bbox *geom.Bounds
		want *geom.Bounds
	}{
		{name: "nil bbox uses default", bbox: nil, want: geom.NewBounds(geom.XY).Set(10000, 300000, 280000, 620000)},
		{name: "explicit bbox preserved", bbox: geom.NewBounds(geom.XY).Set(1, 2, 3, 4), want: geom.NewBounds(geom.XY).Set(1, 2, 3, 4)},
		{name: "explicit zero bbox preserved", bbox: geom.NewBounds(geom.XY).Set(0, 0, 0, 0), want: geom.NewBounds(geom.XY).Set(0, 0, 0, 0)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := MapRequest{bbox: test.bbox}
			require.NoError(t, request.setDefaults(mapsConfig.DefaultMap))
			require.NotNil(t, request.bbox)
			require.Equal(t, test.want, request.bbox)
			if test.bbox != nil {
				require.Same(t, test.bbox, request.bbox)
			}
		})
	}
}
