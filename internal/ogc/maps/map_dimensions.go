package maps

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/twpayne/go-geom"
	proj "github.com/twpayne/go-proj/v11"
)

var errDimensionOverflow = errors.New("calculated image dimension exceeds the supported integer range")

// bboxAspectRatio returns the bbox width divided by its height in CRS units.
// Bounds must use horizontal/vertical axis order.
// Best suited to projected CRSs with equal axis units
// Geographic CRS ratios do not account for the variation in ground distance per degree.
func bboxAspectRatio(bounds *geom.Bounds) (float64, error) {
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

// completeDimensions calculates exactly one missing (zero) dimension.
func completeDimensions(width, height int, ratio float64) (int, int, error) {
	if !finite(ratio) || ratio <= 0 || width < 0 || height < 0 || (width == 0) == (height == 0) {
		return 0, 0, errors.New("a positive ratio and exactly one missing dimension are required")
	}
	value := float64(height) * ratio
	if height == 0 {
		value = float64(width) / ratio
	}
	dimension, err := roundDimension(value)
	if err != nil {
		return 0, 0, err
	}
	if width == 0 {
		width = dimension
	} else {
		height = dimension
	}
	return width, height, nil
}

// roundDimension rounds a calculated dimension to the nearest pixel, with a minimum of one.
func roundDimension(value float64) (int, error) {
	if math.IsNaN(value) || value < 0 {
		return 0, errors.New("calculated image dimension must be non-negative")
	}
	value = math.Max(1, math.Round(value))
	// float64 rounds MaxInt up on 64-bit platforms; use an exclusive upper bound.
	if !finite(value) || value >= math.Exp2(float64(strconv.IntSize-1)) {
		return 0, errDimensionOverflow
	}
	return int(value), nil
}

// bboxAtResolution returns a bbox sized for the given image dimensions and resolution.
// It preserves the center and uses the same CRS-units-per-pixel resolution on both axes.
func bboxAtResolution(bounds *geom.Bounds, width, height int, resolution float64) (*geom.Bounds, error) {
	if _, err := bboxAspectRatio(bounds); err != nil {
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
	if _, err := bboxAspectRatio(result); err != nil {
		return nil, fmt.Errorf("calculated bbox: %w", err)
	}
	return result, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

// physicalBboxDimensions estimates ground dimensions in metres using local scale at the bbox center.
// The bbox must be in horizontal/vertical axis order in the given CRS.
func physicalBboxDimensions(bbox *geom.Bounds, srs string) (float64, float64, error) {
	if _, err := bboxAspectRatio(bbox); err != nil {
		return 0, 0, err
	}
	// Transform samples to WGS84 longitude/latitude for ellipsoidal ground-distance
	// measurements. The map bbox itself remains in the output CRS.
	crs84Transform, crs84, err := setupCRS84DistanceCalculation(srs)
	if err != nil {
		return 0, 0, err
	}
	center := [2]float64{
		bbox.Min(0)/2 + bbox.Max(0)/2,
		bbox.Min(1)/2 + bbox.Max(1)/2,
	}
	var dimensions [2]float64
	for axis := range 2 {
		factor, err := physicalAxisScale(crs84Transform, crs84, bbox, center, axis)
		if err != nil {
			return 0, 0, err
		}
		dimensions[axis] = factor * (bbox.Max(axis) - bbox.Min(axis))
		if !finite(dimensions[axis]) || dimensions[axis] <= 0 {
			return 0, 0, errors.New("invalid physical bbox dimensions")
		}
	}
	return dimensions[0], dimensions[1], nil
}

// setupCRS84DistanceCalculation creates a transformation to CRS84 and a CRS84 object
// used to measure ground distances on the WGS84 ellipsoid.
func setupCRS84DistanceCalculation(srs string) (*proj.PJ, *proj.PJ, error) {
	crs84Transform, err := util.NewTransformation(srs, "OGC:CRS84")
	if err != nil {
		return nil, nil, err
	}
	crs84, err := proj.New("OGC:CRS84")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CRS84 CRS for distance calculations: %w", err)
	}
	return crs84Transform, crs84, nil
}

// physicalAxisScale estimates ground metres per CRS unit at the center.
func physicalAxisScale(crs84Transform, crs84 *proj.PJ, bbox *geom.Bounds, center [2]float64, axis int) (float64, error) {
	span := bbox.Max(axis) - bbox.Min(axis)
	// Sample a small interval around the center to measure local projection distortion.
	step := math.Min(span/2, math.Max(span*0.00001, math.Max(math.Abs(center[axis]), 1)*0.00000001))
	lower, upper := center[axis]-step, center[axis]+step
	if lower == center[axis] {
		lower = math.Nextafter(center[axis], bbox.Min(axis))
	}
	if upper == center[axis] {
		upper = math.Nextafter(center[axis], bbox.Max(axis))
	}
	coordinates := [2][2]float64{center, center}
	coordinates[0][axis], coordinates[1][axis] = lower, upper
	var points [2]proj.Coord
	for i := range points {
		coordinate := coordinates[i]
		point, err := crs84Transform.Forward(proj.NewCoord(coordinate[0], coordinate[1], 0, 0))
		if err != nil {
			return 0, fmt.Errorf("transform bbox center sample: %w", err)
		}
		if !finite(point.X()) || !finite(point.Y()) || math.Abs(point.Y()) > 90 {
			return 0, errors.New("invalid geographic coordinate at bbox center")
		}
		// PROJ's geodesic distance function expects longitude/latitude in radians.
		points[i] = point.DegToRad()
	}
	distance := crs84.LPDist(points[0], points[1])
	if !finite(distance) || distance <= 0 || upper <= lower {
		return 0, errors.New("cannot determine physical scale at bbox center")
	}
	factor := distance / (upper - lower)
	if !finite(factor) || factor <= 0 {
		return 0, errors.New("invalid physical scale at bbox center")
	}
	return factor, nil
}

// bboxAtScale estimates an output CRS bbox from a center and image dimensions.
func bboxAtScale(center geom.Coord, srs string, width, height int, scale, mmPerPixel float64) (*geom.Bounds, error) {
	if len(center) != 2 || !finite(center[0]) || !finite(center[1]) || width <= 0 || height <= 0 {
		return nil, errors.New("center and image dimensions must be finite and valid")
	}
	resolution, err := physicalResolution(scale, mmPerPixel)
	if err != nil {
		return nil, err
	}
	crs84Transform, crs84, err := setupCRS84DistanceCalculation(srs)
	if err != nil {
		return nil, err
	}
	// A unit-sized sampling bbox determines the local scale at the center.
	sample := geom.NewBounds(geom.XY).Set(center[0]-0.5, center[1]-0.5, center[0]+0.5, center[1]+0.5)
	if _, err := bboxAspectRatio(sample); err != nil {
		return nil, fmt.Errorf("center cannot be sampled: %w", err)
	}
	point := [2]float64{center[0], center[1]}
	var halfExtent [2]float64
	for axis, dimension := range [2]int{width, height} {
		factor, err := physicalAxisScale(crs84Transform, crs84, sample, point, axis)
		if err != nil {
			return nil, err
		}
		// Convert half the image's ground size on each axis to CRS units at the center.
		halfExtent[axis] = (float64(dimension) / 2) * (resolution / factor)
	}
	// Construct the output bbox.
	bbox := geom.NewBounds(geom.XY).Set(center[0]-halfExtent[0], center[1]-halfExtent[1], center[0]+halfExtent[0], center[1]+halfExtent[1])
	if _, err := bboxAspectRatio(bbox); err != nil {
		return nil, fmt.Errorf("invalid calculated bbox: %w", err)
	}
	return bbox, nil
}

// physicalResolution calculates ground metres per pixel from the scale
// denominator and display pixel size in millimetres.
func physicalResolution(scale, mmPerPixel float64) (float64, error) {
	resolution := scale * (mmPerPixel / 1000)
	if resolution == 0 {
		return 0, fmt.Errorf("%w: calculated resolution underflows to zero", errDimensionOverflow)
	}
	if !finite(resolution) || resolution < 0 {
		return 0, errors.New("scale and display resolution produce an unrepresentable resolution")
	}
	return resolution, nil
}

// dimensionsAtScale converts local physical bbox dimensions to rounded pixel dimensions.
func dimensionsAtScale(bbox *geom.Bounds, srs string, scale, mmPerPixel float64) (int, int, error) {
	resolution, err := physicalResolution(scale, mmPerPixel)
	if err != nil {
		return 0, 0, err
	}
	widthMetres, heightMetres, err := physicalBboxDimensions(bbox, srs)
	if err != nil {
		return 0, 0, fmt.Errorf("calculate physical map extent: %w", err)
	}
	width, err := roundDimension(widthMetres / resolution)
	if err != nil {
		return 0, 0, err
	}
	height, err := roundDimension(heightMetres / resolution)
	if err != nil {
		return 0, 0, err
	}
	return width, height, nil
}
