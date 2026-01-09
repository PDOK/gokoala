package util

import (
	"fmt"

	"github.com/twpayne/go-geom"
)

// EncodeBBox encodes b as a GeoJson Bounding Box.
// adapted from https://github.com/twpayne/go-geom/blob/b22fd061f1531a51582333b5bd45710a455c4978/encoding/geojson/geojson.go#L525
func EncodeBBox(bbox geom.T) (*[]float64, error) {
	if bbox == nil {
		return nil, nil
	}
	b := bbox.Bounds()
	switch l := b.Layout(); l {
	case geom.XY, geom.XYM:
		return &[]float64{b.Min(0), b.Min(1), b.Max(0), b.Max(1)}, nil
	case geom.XYZ, geom.XYZM, geom.NoLayout:
		return nil, fmt.Errorf("unsupported type: %d", rune(l))
	default:
		return nil, fmt.Errorf("unsupported type: %d", rune(l))
	}
}
