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
		request.mmPerPixel = defaults.MMPerPixel
	}

	return nil
}

func (request *MapRequest) validate() error {
	// Place to do cross-field validation.
	// I.e., if we want to support Scaling, we cannot have both 'bbox' and 'center' params in a single request

	return nil
}
