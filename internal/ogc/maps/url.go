package maps

import (
	"errors"
	"fmt"
	"math"
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
	CrsParam     = "crs"
	BboxParam    = "bbox"
	BboxCrsParam = "bbox-crs"
	WidthParam   = "width"
	HeightParam  = "height"
	Transparent  = "transparent"
	MMPerPixel   = "mm-per-pixel"
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

func (m *mapsURL) parse(supportedCrs []domain.ContentCrs) (MapRequest, error) {
	err := m.validateNoUnknownParams()
	if err != nil {
		return MapRequest{}, err
	}

	crs, crsErr := parseCrsParam(m.params, CrsParam)
	if crsErr == nil && crs != "" {
		crsErr = validateSupportedCrs(CrsParam, crs, supportedCrs)
	}
	bbox, bboxSRID, bboxErr := parseBbox(m.params)
	width, widthErr := parseWidth(m.params)
	height, heightErr := parseHeight(m.params)
	transparent, transparentErr := parseTransparent(m.params)
	mmPerPixel, mmPerPixelErr := parseMMPerPixel(m.params)

	if err = errors.Join(err, crsErr, bboxErr, widthErr, heightErr, transparentErr, mmPerPixelErr); err != nil {
		return MapRequest{}, err
	}

	request := MapRequest{
		transparent: transparent,
		requested:   MapViewParameters{crs: crs, bboxSRID: bboxSRID, bbox: bbox, width: width, height: height, mmPerPixel: mmPerPixel},
	}

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

	crs, err := parseCrsParam(params, BboxCrsParam)
	if err != nil {
		return nil, domain.UndefinedSRID, err
	}
	bbox, bboxSRID, err := features.ParseBbox(url.Values{
		BboxParam:    {params.Get(BboxParam)},
		BboxCrsParam: {string(crs)},
	})
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

// parseCrsParam normalizes and parses a given crs param (i.e. 'crs', 'bbox-crs', etc.).
func parseCrsParam(params url.Values, name string) (domain.ContentCrs, error) {
	identifier := strings.TrimSpace(params.Get(name))
	if identifier == "" {
		return "", nil
	}
	crs, _, err := parseCrs(identifier)
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	return crs, nil
}

func parseCrs(identifier string) (domain.ContentCrs, domain.SRID, error) {
	identifier = strings.TrimSpace(identifier)
	
	// Handle Safe CURIE
	if domain.HasSafeCURIESyntax(identifier) {
		srid, err := domain.ParseSafeCURIEToSRID(identifier)
		if err != nil {
			return "", domain.UndefinedSRID, err
		}
		if identifier == "["+domain.OGCPrefix+domain.WGS84CodeOGC+"]" {
			return domain.ContentCrs(domain.WGS84CrsURI), srid, nil
		}
		return srid.ToContentCrs(), srid, nil
	}

	// Handle CRS URIs
	if identifier != domain.WGS84CrsURI && !strings.HasPrefix(identifier, domain.CrsURIPrefix+"EPSG/0/") {
		return "", domain.UndefinedSRID, fmt.Errorf("expected an EPSG CRS URI or CRS84 URI, got %q", identifier)
	}
	srid, err := domain.ParseCrsURIToSRID(identifier)
	if err != nil {
		return "", domain.UndefinedSRID, fmt.Errorf("invalid CRS %q: %w", identifier, err)
	}
	if srid <= 0 {
		return "", domain.UndefinedSRID, fmt.Errorf("CRS code must be positive, got %q", identifier)
	}
	return domain.ContentCrs(identifier), srid, nil
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

	if math.IsNaN(mmPerPixelParam) || math.IsInf(mmPerPixelParam, 0) || mmPerPixelParam <= 0 {
		return 0, fmt.Errorf("value is not a finite positive float: %s", param)
	}

	if _, err = toMapResolution(mmPerPixelParam); err != nil {
		return 0, err
	}

	return mmPerPixelParam, nil
}
