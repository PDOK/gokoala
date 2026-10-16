package util

import (
	"errors"
	"fmt"
	"math"

	"github.com/twpayne/go-geom"
	proj "github.com/twpayne/go-proj/v11"
)

// TransformPoint transforms a 2D point using longitude/latitude or easting/northing order.
func TransformPoint(point geom.Coord, inputSRS, outputSRS string) (geom.Coord, error) {
	if len(point) != 2 {
		return nil, errors.New("point must contain exactly two coordinates")
	}
	transformation, err := newTransformation(inputSRS, outputSRS)
	if err != nil {
		return nil, err
	}
	result, err := transformation.Forward(proj.NewCoord(point[0], point[1], 0, 0))
	if err != nil {
		return nil, fmt.Errorf("transform point: %w", err)
	}
	return geom.Coord{result.X(), result.Y()}, nil
}

// TransformBoundingBox transforms XY bounds using longitude/latitude or easting/northing order.
func TransformBoundingBox(bbox *geom.Bounds, inputSRS, outputSRS string) (*geom.Bounds, error) {
	if bbox == nil || bbox.Layout() != geom.XY || bbox.IsEmpty() {
		return nil, errors.New("bbox must contain non-empty XY bounds")
	}
	transformation, err := newTransformation(inputSRS, outputSRS)
	if err != nil {
		return nil, err
	}
	lowerLeft, err := transformation.Forward(proj.NewCoord(bbox.Min(0), bbox.Min(1), 0, 0))
	if err != nil {
		return nil, fmt.Errorf("transform lower-left corner: %w", err)
	}
	upperRight, err := transformation.Forward(proj.NewCoord(bbox.Max(0), bbox.Max(1), 0, 0))
	if err != nil {
		return nil, fmt.Errorf("transform upper-right corner: %w", err)
	}
	// Consider using densification:
	// bounds := proj.Bounds{XMin: bbox.Min(0), YMin: bbox.Min(1), XMax: bbox.Max(0), YMax: bbox.Max(1)}
	// result, err := transformation.ForwardBounds(bounds, 21)
	// Sampling edges can capture extrema missed by transforming only two corners.
	return geom.NewBounds(geom.XY).Set(
		math.Min(lowerLeft.X(), upperRight.X()), math.Min(lowerLeft.Y(), upperRight.Y()),
		math.Max(lowerLeft.X(), upperRight.X()), math.Max(lowerLeft.Y(), upperRight.Y()),
	), nil
}

func newTransformation(inputSRS, outputSRS string) (*proj.PJ, error) {
	transformation, err := proj.NewCRSToCRS(inputSRS, outputSRS, nil)
	if err != nil {
		return nil, fmt.Errorf("create transformation from %q to %q: %w", inputSRS, outputSRS, err)
	}
	normalized, err := transformation.NormalizeForVisualization()
	if err != nil {
		return nil, fmt.Errorf("normalize transformation axis order: %w", err)
	}
	return normalized, nil
}
