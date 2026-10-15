package maps

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapRequestSetDefaultsBBox(t *testing.T) {
	tests := []struct {
		name string
		bbox *[4]float64
		want [4]float64
	}{
		{name: "nil bbox uses default", bbox: nil, want: [4]float64{10000, 300000, 280000, 620000}},
		{name: "explicit bbox preserved", bbox: &[4]float64{1, 2, 3, 4}, want: [4]float64{1, 2, 3, 4}},
		{name: "explicit zero bbox preserved", bbox: &[4]float64{}, want: [4]float64{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := MapRequest{bbox: test.bbox}
			require.NoError(t, request.setDefaults(mapsConfig.DefaultMap))
			require.NotNil(t, request.bbox)
			require.Equal(t, test.want, *request.bbox)
			if test.bbox != nil {
				require.Same(t, test.bbox, request.bbox)
			}
		})
	}
}
