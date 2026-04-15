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

// PadBbox For linestrings specifically, it is possible for the min and max values of a dimension
// to be exactly equal, resulting in a bbox with 0 area. This function pads the max value
// of the relevant dimension by 1 unit to avoid this.
// Returns the original bbox if neither dimension has equal min and max values.
func PadBbox(bbox *geom.Bounds) *geom.Bounds {
	if bbox.Max(0) == bbox.Min(0) {
		// pad bbox.Max(0)
		return geom.NewBounds(geom.XY).Set(bbox.Min(0), bbox.Min(1), bbox.Max(0)+1, bbox.Max(1))
	} else if bbox.Max(1) == bbox.Min(1) {
		// pad bbox.Max(1)
		return geom.NewBounds(geom.XY).Set(bbox.Min(0), bbox.Min(1), bbox.Max(0), bbox.Max(1)+1)
	}
	return bbox
}
