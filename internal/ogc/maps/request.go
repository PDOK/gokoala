package maps

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/PDOK/gokoala/internal/ogc/features/proj"
	"github.com/twpayne/go-geom"
)

type MapType string

const (
	DatasetMap    MapType = "Dataset Map"
	CollectionMap MapType = "Collection Map"
)

var ErrMapTooLarge = errors.New("map dimensions exceed configured limits")

type MapRequest struct {
	mapType     MapType
	collection  string
	style       string
	transparent bool
	format      string

	requested MapViewParameters // Parsed input parameters from the request
	resolved  MapViewParameters // Resolved parameters including defaults and transformations
}

type MapViewParameters struct {
	crs              domain.ContentCrs
	bboxSRID         domain.SRID
	bbox             *geom.Bounds
	center           geom.Coord
	centerSRID       domain.SRID
	width            int
	height           int
	mmPerPixel       float64
	scaleDenominator float64
}

// validateParamCombinations rejects unsupported param combinations
// before defaults and calculated values are applied.
func (request *MapRequest) validateParamCombinations() error {
	requested := request.requested
	if requested.scaleDenominator != 0 && (requested.width != 0 || requested.height != 0) {
		return errors.New("scale-denominator cannot be combined with width or height")
	}
	if requested.center != nil {
		if requested.bbox != nil {
			return errors.New("center cannot be combined with bbox")
		}
		if requested.scaleDenominator == 0 {
			return errors.New("center requires scale-denominator")
		}
		if requested.width != 0 || requested.height != 0 {
			return errors.New("center with scale-denominator cannot be combined with width or height")
		}
	}
	return nil
}

// resolveValues applies defaults and resolves the bbox and dimensions in the output CRS.
func (request *MapRequest) resolveValues(defaults config.DefaultMap) error {
	resolved := request.requested
	if resolved.crs == "" {
		crs, err := getDefaultCrs(defaults)
		if err != nil {
			return err
		}
		resolved.crs = crs
	}
	if resolved.mmPerPixel == 0 {
		mmPerPixel, err := strconv.ParseFloat(defaults.MMPerPixel, 64)
		if err != nil {
			return fmt.Errorf("invalid default mm-per-pixel %q: %w", defaults.MMPerPixel, err)
		}
		resolved.mmPerPixel = mmPerPixel
	}

	if resolved.center != nil {
		if err := resolveForCenter(&resolved, defaults); err != nil {
			return err
		}
	} else {
		if err := resolveForBbox(&resolved, request.requested, defaults); err != nil {
			return err
		}
	}

	request.resolved = resolved
	return nil
}

// resolveForBbox resolves dimensions and the requested or default bbox in the output CRS.
func resolveForBbox(resolved *MapViewParameters, requested MapViewParameters, defaults config.DefaultMap) error {
	if resolved.bbox == nil {
		bbox, err := parseDefaultBbox(defaults)
		if err != nil {
			return err
		}

		bboxSRID, err := domain.EpsgToSrid(defaults.CRS)
		if err != nil {
			return fmt.Errorf("invalid default bbox CRS %q: %w", defaults.CRS, err)
		}

		resolved.bbox = bbox
		resolved.bboxSRID = bboxSRID
	}

	bbox, swapped, bboxSRID, err := resolved.resolveNormalizedOutputBbox()
	if err != nil {
		return err
	}
	bbox, err = resolveMapDimensions(resolved, requested, bbox, defaults)
	if err != nil {
		return err
	}
	if swapped {
		bbox = swapBboxAxes(bbox)
	}
	resolved.bbox = bbox
	resolved.bboxSRID = bboxSRID
	return nil
}

// resolveForCenter calculates an output-CRS bbox around the center using the scale denominator and configured default dimensions.
func resolveForCenter(resolved *MapViewParameters, defaults config.DefaultMap) error {
	if _, err := toMapResolution(resolved.mmPerPixel); err != nil {
		return err
	}
	_, outputSRID, err := parseCrs(string(resolved.crs))
	if err != nil {
		return err
	}
	outputSRS, err := domain.SridToSrs(outputSRID)
	if err != nil {
		return err
	}
	center, err := transformCenter(resolved.center, resolved.centerSRID, outputSRS)
	if err != nil {
		return err
	}
	resolved.width, resolved.height = defaults.Width, defaults.Height
	bbox, err := bboxAtScale(center, outputSRS, resolved.width, resolved.height, resolved.scaleDenominator, resolved.mmPerPixel)
	if err != nil {
		return fmt.Errorf("calculate bbox from center: %w", err)
	}
	outputOrder, err := proj.GetAxisOrderForCRS(outputSRS)
	if err != nil {
		return err
	}
	if outputOrder == domain.AxisOrderYX {
		bbox = swapBboxAxes(bbox)
	}
	resolved.bbox, resolved.bboxSRID = bbox, outputSRID
	return nil
}

// transformCenter normalizes center coordinates to x/y order when needed, then transforms them to the output CRS.
func transformCenter(center geom.Coord, inputSRID domain.SRID, outputSRS string) (geom.Coord, error) {
	inputSRS, err := domain.SridToSrs(inputSRID)
	if err != nil {
		return nil, fmt.Errorf("center CRS: %w", err)
	}
	inputOrder, err := proj.GetAxisOrderForCRS(inputSRS)
	if err != nil {
		return nil, fmt.Errorf("center CRS: %w", err)
	}
	point := append(geom.Coord(nil), center...)
	if inputOrder == domain.AxisOrderYX {
		point[0], point[1] = point[1], point[0]
	}
	point, err = util.TransformPoint(point, inputSRS, outputSRS)
	if err != nil {
		return nil, fmt.Errorf("transform center: %w", err)
	}
	if !finite(point[0]) || !finite(point[1]) {
		return nil, errors.New("transformed center coordinates must be finite")
	}
	return point, nil
}

// resolveMapDimensions completes image dimensions from the normalized output bbox.
// With both dimensions but no explicit bbox, it resizes the bbox preserving default center and vertical resolution.
func resolveMapDimensions(resolved *MapViewParameters, requested MapViewParameters,
	bbox *geom.Bounds, defaults config.DefaultMap) (*geom.Bounds, error) {
	if requested.scaleDenominator != 0 {
		err := resolveScaleDimensions(resolved, bbox)
		return bbox, err
	}
	hasBbox := requested.bbox != nil
	hasWidth, hasHeight := requested.width != 0, requested.height != 0
	ratio, err := bboxAspectRatio(bbox)
	if err != nil {
		return nil, err
	}
	switch {
	case !hasBbox && !hasWidth && !hasHeight:
		resolved.width, resolved.height = defaults.Width, defaults.Height
	case !hasBbox && hasWidth && hasHeight:
		resolutionY := (bbox.Max(1) - bbox.Min(1)) / float64(defaults.Height)
		bbox, err = bboxAtResolution(bbox, resolved.width, resolved.height, resolutionY)
	case !hasWidth || !hasHeight:
		if !hasWidth && !hasHeight {
			resolved.height = defaults.Height
		}
		resolved.width, resolved.height, err = completeDimensions(resolved.width, resolved.height, ratio)
	}
	if err != nil {
		if errors.Is(err, errDimensionOverflow) {
			return nil, fmt.Errorf("%w: %w", ErrMapTooLarge, err)
		}
		return nil, err
	}
	return bbox, nil
}

func parseDefaultBbox(defaults config.DefaultMap) (*geom.Bounds, error) {
	if len(defaults.Bbox) != 4 {
		return nil, errors.New("default bbox must contain four coordinates")
	}
	var bbox [4]float64
	for index, coordinate := range defaults.Bbox {
		value, err := strconv.ParseFloat(coordinate, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid default bbox coordinate %q: %w", coordinate, err)
		}
		bbox[index] = value
	}
	bounds := geom.NewBounds(geom.XY).Set(bbox[:]...)
	if _, err := bboxAspectRatio(bounds); err != nil {
		return nil, fmt.Errorf("invalid default bbox: %w", err)
	}
	return bounds, nil
}

// resolveNormalizedOutputBbox returns a new bbox in the output CRS with horizontal/vertical axes.
// It also returns whether native output axes must be restored and the output SRID.
func (resolved MapViewParameters) resolveNormalizedOutputBbox() (*geom.Bounds, bool, domain.SRID, error) {
	if resolved.bbox == nil {
		return nil, false, domain.UndefinedSRID, errors.New("bbox is required")
	}
	_, outputSRID, err := parseCrs(string(resolved.crs))
	if err != nil {
		return nil, false, domain.UndefinedSRID, err
	}
	inputSRID := domain.SRID(resolved.bboxSRID.GetOrDefault())
	coordinates := [4]float64{resolved.bbox.Min(0), resolved.bbox.Min(1), resolved.bbox.Max(0), resolved.bbox.Max(1)}
	if inputSRID != outputSRID {
		coordinates, err = transformBoundingBox(coordinates, inputSRID, outputSRID)
		if err != nil {
			return nil, false, domain.UndefinedSRID, err
		}
	}
	bbox := geom.NewBounds(geom.XY).Set(coordinates[:]...)
	srs, err := domain.SridToSrs(outputSRID)
	if err != nil {
		return nil, false, domain.UndefinedSRID, err
	}
	order, err := proj.GetAxisOrderForCRS(srs)
	if err != nil {
		return nil, false, domain.UndefinedSRID, err
	}
	swapped := order == domain.AxisOrderYX
	if swapped {
		bbox = swapBboxAxes(bbox)
	}
	return bbox, swapped, outputSRID, nil
}

func swapBboxAxes(bbox *geom.Bounds) *geom.Bounds {
	return geom.NewBounds(geom.XY).Set(bbox.Min(1), bbox.Min(0), bbox.Max(1), bbox.Max(0))
}

func transformBoundingBox(bbox [4]float64, inputSRID, outputSRID domain.SRID) ([4]float64, error) {
	inputSRS, err := domain.SridToSrs(inputSRID)
	if err != nil {
		return [4]float64{}, err
	}
	outputSRS, err := domain.SridToSrs(outputSRID)
	if err != nil {
		return [4]float64{}, err
	}
	inputAxisOrder, err := proj.GetAxisOrderForCRS(inputSRS)
	if err != nil {
		return [4]float64{}, err
	}
	outputAxisOrder, err := proj.GetAxisOrderForCRS(outputSRS)
	if err != nil {
		return [4]float64{}, err
	}
	if inputAxisOrder == domain.AxisOrderYX {
		bbox = [4]float64{bbox[1], bbox[0], bbox[3], bbox[2]}
	}
	transformed, err := util.TransformBoundingBox(
		geom.NewBounds(geom.XY).Set(bbox[:]...), inputSRS, outputSRS,
	)
	if err != nil {
		return [4]float64{}, fmt.Errorf("transform map bbox: %w", err)
	}
	bbox = [4]float64{transformed.Min(0), transformed.Min(1), transformed.Max(0), transformed.Max(1)}
	if outputAxisOrder == domain.AxisOrderYX {
		bbox = [4]float64{bbox[1], bbox[0], bbox[3], bbox[2]}
	}
	return bbox, nil
}

// validateResolved checks resolved values after defaults have been applied.
func (request *MapRequest) validateResolved(limits *config.OgcAPIMaps) error {
	resolved := request.resolved
	if _, err := toMapResolution(resolved.mmPerPixel); err != nil {
		return err
	}

	if resolved.width <= 0 || resolved.height <= 0 {
		return errors.New("map dimensions must be positive")
	}
	if _, err := bboxAspectRatio(resolved.bbox); err != nil {
		return fmt.Errorf("invalid resolved bbox: %w", err)
	}
	_, bboxSRID, err := parseCrs(string(resolved.crs))
	if err != nil {
		return err
	}
	if resolved.bboxSRID != bboxSRID {
		return errors.New("resolved bbox CRS must match the map CRS")
	}
	var errs []error
	if resolved.width > limits.MaxWidth {
		errs = append(errs, fmt.Errorf("width %d exceeds maxWidth (%d)", resolved.width, limits.MaxWidth))
	}
	if resolved.height > limits.MaxHeight {
		errs = append(errs, fmt.Errorf("height %d exceeds maxHeight (%d)", resolved.height, limits.MaxHeight))
	}
	// Divide to avoid overflowing when multiplying request dimensions.
	if int64(resolved.width) > limits.MaxPixels()/int64(resolved.height) {
		errs = append(errs, fmt.Errorf("width times height exceeds maxPixels (%d)", limits.MaxPixels()))
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("%w: %w", ErrMapTooLarge, err)
	}
	return nil
}

// resolveScaleDimensions calculates image dimensions for the resolved bbox and requested scale.
func resolveScaleDimensions(resolved *MapViewParameters, bbox *geom.Bounds) error {
	if _, err := toMapResolution(resolved.mmPerPixel); err != nil {
		return err
	}
	_, srid, err := parseCrs(string(resolved.crs))
	if err != nil {
		return err
	}
	srs, err := domain.SridToSrs(srid)
	if err != nil {
		return err
	}
	resolved.width, resolved.height, err = dimensionsAtScale(bbox, srs, resolved.scaleDenominator, resolved.mmPerPixel)
	if errors.Is(err, errDimensionOverflow) {
		return fmt.Errorf("%w: %w", ErrMapTooLarge, err)
	}
	return err
}
