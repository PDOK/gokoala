package util

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/twpayne/go-geom"
)

var ErrDimensionOverflow = errors.New("calculated image dimension exceeds the supported integer range")

// BboxAspectRatio returns the bbox width divided by its height in CRS units.
// Bounds must use horizontal/vertical axis order.
// Best suited to projected CRSs with equal axis units
// Geographic CRS ratios do not account for the variation in ground distance per degree.
func BboxAspectRatio(bounds *geom.Bounds) (float64, error) {
	if bounds == nil || bounds.Layout() != geom.XY || bounds.IsEmpty() {
		return 0, errors.New("bbox must contain non-empty XY bounds")
	}
	for axis := range 2 {
		if !finite(bounds.Min(axis)) || !finite(bounds.Max(axis)) || bounds.Max(axis) <= bounds.Min(axis) {
			return 0, errors.New("bbox coordinates must be finite with positive extents")
		}
	}
	ratio := (bounds.Max(0) - bounds.Min(0)) / (bounds.Max(1) - bounds.Min(1))
	if !finite(ratio) || ratio <= 0 {
		return 0, errors.New("bbox aspect ratio must be finite and positive")
	}
	return ratio, nil
}

// CompleteDimensions calculates exactly one missing (zero) dimension.
func CompleteDimensions(width, height int, ratio float64) (int, int, error) {
	if !finite(ratio) || ratio <= 0 || width < 0 || height < 0 || (width == 0) == (height == 0) {
		return 0, 0, errors.New("a positive ratio and exactly one missing dimension are required")
	}
	value := float64(height) * ratio
	if height == 0 {
		value = float64(width) / ratio
	}
	value = math.Max(1, math.Round(value))
	// float64 rounds MaxInt up on 64-bit platforms; use an exclusive upper bound.
	if !finite(value) || value >= math.Exp2(float64(strconv.IntSize-1)) {
		return 0, 0, ErrDimensionOverflow
	}
	if width == 0 {
		width = int(value)
	} else {
		height = int(value)
	}
	return width, height, nil
}

// BboxAtResolution returns a bbox sized for the given image dimensions and resolution.
// It preserves the center and uses the same CRS-units-per-pixel resolution on both axes.
func BboxAtResolution(bounds *geom.Bounds, width, height int, resolution float64) (*geom.Bounds, error) {
	if _, err := BboxAspectRatio(bounds); err != nil {
		return nil, err
	}
	if width <= 0 || height <= 0 || !finite(resolution) || resolution <= 0 {
		return nil, errors.New("dimensions and resolution must be finite and positive")
	}
	centerX := bounds.Min(0)/2 + bounds.Max(0)/2
	centerY := bounds.Min(1)/2 + bounds.Max(1)/2
	halfWidth := float64(width) * resolution / 2
	halfHeight := float64(height) * resolution / 2
	result := geom.NewBounds(geom.XY).Set(centerX-halfWidth, centerY-halfHeight, centerX+halfWidth, centerY+halfHeight)
	if _, err := BboxAspectRatio(result); err != nil {
		return nil, fmt.Errorf("calculated bbox: %w", err)
	}
	return result, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
