package maps

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/twpayne/go-geom"
)

const (
	CrsParam        = "crs"
	BboxParam       = "bbox"
	BboxCrsParam    = "bbox-crs"
	WidthParam      = "width"
	HeightParam     = "height"
	Transparent = "transparent"
	MMPerPixel  = "mm-per-pixel"

	maxAllowed = 5000
)

var mapsKnownParams = map[string]struct{}{
	engine.FormatParam: {},
	CrsParam:           {},
	BboxParam:          {},
	BboxCrsParam:       {},
	WidthParam:         {},
	HeightParam:        {},
	Transparent:        {},
	MMPerPixel:         {},
}

type mapsURL struct {
	baseURL url.URL
	params  url.Values
}

func (m *mapsURL) parse(defaultCrs domain.ContentCrs, supportedCrs []domain.ContentCrs) (MapRequest, error) {
	err := m.validateNoUnknownParams()
	if err != nil {
		return MapRequest{}, err
	}

	crs := parseCrsToContentCrs(m.params, defaultCrs)
	crsErr := validateSupportedCrs(CrsParam, crs, supportedCrs)
	bbox, bboxSRID, bboxErr := parseBbox(m.params)
	if bboxErr == nil && bbox != nil && m.params.Get(BboxCrsParam) != "" {
		bboxCrs := domain.ContentCrs(m.params.Get(BboxCrsParam))
		bboxErr = validateSupportedCrs(BboxCrsParam, bboxCrs, supportedCrs)
	}
	width, widthErr := parseWidth(m.params)
	height, heightErr := parseHeight(m.params)
	transparent, transparentErr := parseTransparent(m.params)
	mmPerPixel, mmPerPixelErr := parseMMPerPixel(m.params)

	if height > maxAllowed {
		msg := fmt.Sprintf("height %d exceeds the maximum allowed number (%d)", height, maxAllowed)
		maxHeightErr := errors.New(msg)
		err = errors.Join(err, maxHeightErr)
	}

	if width > maxAllowed {
		msg := fmt.Sprintf("width %d exceeds the maximum allowed number (%d)", width, maxAllowed)
		maxWidthError := errors.New(msg)
		err = errors.Join(err, maxWidthError)
	}

	if err = errors.Join(err, crsErr, bboxErr, widthErr, heightErr, transparentErr, mmPerPixelErr); err != nil {
		return MapRequest{}, err
	}

	request := MapRequest{crs: crs, bboxSRID: bboxSRID, bbox: bbox, width: width, height: height, transparent: transparent, mmPerPixel: mmPerPixel}

	return request, nil
}

func validateSupportedCrs(param string, crs domain.ContentCrs, supportedCrs []domain.ContentCrs) error {
	if slices.Contains(supportedCrs, crs) {
		return nil
	}

	return fmt.Errorf("%s %s is not supported, must be one of: %v", param, crs, supportedCrs)
}

func parseBbox(params url.Values) (*geom.Bounds, domain.SRID, error) {
	if params.Get(BboxParam) == "" {
		return nil, domain.UndefinedSRID, nil
	}

	bbox, bboxSRID, err := features.ParseBbox(params)
	if err != nil {
		return nil, bboxSRID, err
	}

	if bbox.Min(1) > bbox.Max(1) {
		return nil, bboxSRID, fmt.Errorf("bbox miny (%v) must not exceed maxy (%v)", bbox.Min(1), bbox.Max(1))
	}

	if bbox.Min(0) > bbox.Max(0) {
		return nil, bboxSRID, fmt.Errorf("bbox minx (%v) must not exceed maxx (%v)", bbox.Min(0), bbox.Max(0))
	}

	return bbox, bboxSRID, nil
}

func parseWidth(params url.Values) (int, error) {
	widthStr := params.Get(WidthParam)
	return parseAndValidatePositiveInt(widthStr)
}

func parseHeight(params url.Values) (int, error) {
	heightStr := params.Get(HeightParam)
	return parseAndValidatePositiveInt(heightStr)
}

func parseTransparent(params url.Values) (bool, error) {
	if !params.Has(Transparent) {
		return true, nil
	}
	switch strings.ToLower(params.Get(Transparent)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be true or false", Transparent)
	}
}

func (m *mapsURL) validateNoUnknownParams() error {
	for param := range m.params {
		if _, ok := mapsKnownParams[param]; !ok {
			return fmt.Errorf("unknown query parameter(s) found: %s", param)
		}
	}
	return nil
}

func parseAndValidatePositiveInt(intStr string) (int, error) {
	if intStr == "" {
		return 0, nil
	}
	i, err := strconv.Atoi(intStr)
	if err != nil {
		return 0, fmt.Errorf("value is not a valid integer: %s", intStr)
	}
	if i <= 0 {
		return 0, fmt.Errorf("value is not a positive integer: %d", i)
	}
	return i, err
}

func parseCrsToContentCrs(params url.Values, fallback domain.ContentCrs) domain.ContentCrs {
	if param := strings.TrimSpace(params.Get(CrsParam)); param != "" {
		return domain.ContentCrs(param)
	}

	return fallback
}

func parseMMPerPixel(params url.Values) (float64, error) {
	param := strings.TrimSpace(params.Get(MMPerPixel))
	if param == "" {
		return 0, nil
	}

	mmPerPixelParam, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return 0, fmt.Errorf("value is not a valid float: %s", param)
	}

	// mm-per-pixel cannot be negative or 0
	if mmPerPixelParam <= 0 {
		return 0, fmt.Errorf("value is not a positive float: %f", mmPerPixelParam)
	}

	return mmPerPixelParam, nil
}
