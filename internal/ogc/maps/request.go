package maps

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
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
	crs         domain.ContentCrs
	bboxSRID    domain.SRID
	bbox        *geom.Bounds
	width       int
	height      int
	transparent bool
	format      string
	mmPerPixel  float64
}

func (request *MapRequest) setDefaults(defaults config.DefaultMap) error {
	if request.width == 0 {
		request.width = defaults.Width
	}

	if request.height == 0 {
		request.height = defaults.Height
	}

	if request.bbox == nil {
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

		request.bbox = geom.NewBounds(geom.XY).Set(bbox[:]...)
		request.bboxSRID = bboxSRID
	}

	if request.mmPerPixel == 0 {
		mmPerPixel, err := strconv.ParseFloat(defaults.MMPerPixel, 64)
		if err != nil {
			return fmt.Errorf("invalid default mm-per-pixel %q: %w", defaults.MMPerPixel, err)
		}
		request.mmPerPixel = mmPerPixel
	}

	return nil
}

// validate checks the request after defaults have been applied.
func (request *MapRequest) validate(limits *config.OgcAPIMaps) error {
	// Place to do cross-field validation.
	// I.e., if we want to support Scaling, we cannot have both 'bbox' and 'center' params in a single request

	if _, err := toMapResolution(request.mmPerPixel); err != nil {
		return err
	}

	if request.width <= 0 || request.height <= 0 {
		return errors.New("map dimensions must be positive")
	}
	var errs []error
	if request.width > limits.MaxWidth {
		errs = append(errs, fmt.Errorf("width %d exceeds maxWidth (%d)", request.width, limits.MaxWidth))
	}
	if request.height > limits.MaxHeight {
		errs = append(errs, fmt.Errorf("height %d exceeds maxHeight (%d)", request.height, limits.MaxHeight))
	}
	// Divide to avoid overflowing when multiplying request dimensions.
	if int64(request.width) > limits.MaxPixels()/int64(request.height) {
		errs = append(errs, fmt.Errorf("width times height exceeds maxPixels (%d)", limits.MaxPixels()))
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("%w: %w", ErrMapTooLarge, err)
	}
	return nil
}
