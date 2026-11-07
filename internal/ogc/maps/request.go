package maps

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/features"
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
	crs        domain.ContentCrs
	bboxSRID   domain.SRID
	bbox       *geom.Bounds
	width      int
	height     int
	mmPerPixel float64
}

func (request *MapRequest) setDefaults(defaults config.DefaultMap) error {
	resolved := request.requested
	if resolved.crs == "" {
		crs, err := getDefaultCrs(defaults)
		if err != nil {
			return err
		}
		resolved.crs = crs
	}
	if resolved.width == 0 {
		resolved.width = defaults.Width
	}

	if resolved.height == 0 {
		resolved.height = defaults.Height
	}

	if resolved.bbox == nil {
		if len(defaults.Bbox) != 4 {
			return errors.New("default bbox must contain four coordinates")
		}
		var bbox [4]float64
		for index, coordinate := range defaults.Bbox {
			value, err := strconv.ParseFloat(coordinate, 64)
			if err != nil {
				return fmt.Errorf("invalid default bbox coordinate %q: %w", coordinate, err)
			}
			bbox[index] = value
		}

		bboxSRID, err := domain.EpsgToSrid(defaults.CRS)
		if err != nil {
			return fmt.Errorf("invalid default bbox CRS %q: %w", defaults.CRS, err)
		}

		resolved.bbox = geom.NewBounds(geom.XY).Set(bbox[:]...)
		resolved.bboxSRID = bboxSRID
	}

	if resolved.mmPerPixel == 0 {
		mmPerPixel, err := strconv.ParseFloat(defaults.MMPerPixel, 64)
		if err != nil {
			return fmt.Errorf("invalid default mm-per-pixel %q: %w", defaults.MMPerPixel, err)
		}
		resolved.mmPerPixel = mmPerPixel
	}

	bboxSRID, err := mapCRSToSRID(resolved.crs)
	if err != nil {
		return err
	}
	bbox := [4]float64{resolved.bbox.Min(0), resolved.bbox.Min(1), resolved.bbox.Max(0), resolved.bbox.Max(1)}
	if inputSRID := resolved.bboxSRID.GetOrDefault(); inputSRID != int(bboxSRID) {
		bbox, err = transformBoundingBox(bbox, inputSRID, int(bboxSRID), bboxSRID == domain.WGS84SRID)
		if err != nil {
			return err
		}
	}
	resolved.bbox = geom.NewBounds(geom.XY).Set(bbox[:]...)
	resolved.bboxSRID = bboxSRID
	request.resolved = resolved
	return nil
}

func mapCRSToSRID(crs domain.ContentCrs) (domain.SRID, error) {
	crsURI := strings.TrimSpace(string(crs))
	if crsURI == domain.WGS84CrsURI {
		return domain.WGS84SRID, nil
	}
	if !strings.HasPrefix(crsURI, domain.CrsURIPrefix+"EPSG/0/") {
		return domain.UndefinedSRID, fmt.Errorf("expected an EPSG CRS URI or CRS84 URI, got %q", crs)
	}
	srid, err := features.ParseCrsToSRID(url.Values{CrsParam: {crsURI}}, CrsParam)
	if err != nil {
		return domain.UndefinedSRID, fmt.Errorf("invalid CRS %q: %w", crs, err)
	}
	if srid <= 0 {
		return domain.UndefinedSRID, fmt.Errorf("CRS code must be positive, got %q", crs)
	}
	return srid, nil
}

func getAxisOrder(srs string) (domain.AxisOrder, error) {
	projJSON, err := proj.ExecProjInfo(srs)
	if err != nil {
		return domain.AxisOrderXY, fmt.Errorf("get CRS information for %q: %w", srs, err)
	}
	axisOrder, err := proj.GetAxisOrder(projJSON)
	if err != nil {
		return domain.AxisOrderXY, fmt.Errorf("get axis order for %q: %w", srs, err)
	}
	return axisOrder, nil
}

func transformBoundingBox(bbox [4]float64, inputSRID, outputSRID int, outputCRS84 bool) ([4]float64, error) {
	inputSRS := fmt.Sprintf("EPSG:%d", inputSRID)
	if inputSRID == domain.WGS84SRID {
		inputSRS = "OGC:CRS84"
	}
	outputSRS := fmt.Sprintf("EPSG:%d", outputSRID)
	if outputCRS84 {
		outputSRS = "OGC:CRS84"
	}
	inputAxisOrder, err := getAxisOrder(inputSRS)
	if err != nil {
		return [4]float64{}, err
	}
	outputAxisOrder, err := getAxisOrder(outputSRS)
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

// validate checks resolved values after defaults have been applied.
func (request *MapRequest) validate(limits *config.OgcAPIMaps) error {
	// Place to do cross-field validation.
	// I.e., if we want to support Scaling, we cannot have both 'bbox' and 'center' params in a single request

	resolved := request.resolved
	if _, err := toMapResolution(resolved.mmPerPixel); err != nil {
		return err
	}

	if resolved.width <= 0 || resolved.height <= 0 {
		return errors.New("map dimensions must be positive")
	}
	if resolved.bbox == nil {
		return errors.New("resolved bbox is required")
	}
	bboxSRID, err := mapCRSToSRID(resolved.crs)
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
