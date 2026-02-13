package util

import (
	"math"

	"github.com/twpayne/go-geom"
)

func SurfaceArea(bbox *geom.Bounds) float64 {
	// Use the same logic as bbox.Area() in https://github.com/go-spatial/geom to calculate surface area.
	// The bounds.Area() in github.com/twpayne/go-geom behaves differently and is not what we're looking for.
	return math.Abs((bbox.Max(1) - bbox.Min(1)) * (bbox.Max(0) - bbox.Min(0)))
}
