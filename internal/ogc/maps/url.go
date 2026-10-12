package maps

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/PDOK/gokoala/internal/ogc/features"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/twpayne/go-geom"
)

const (
	CrsParam     = "crs"
	BboxParam    = "bbox"
	BboxCrsParam = "bbox-crs"
	Width        = "width"
	Height       = "height"

	maxAllowed = 5000
)

var mapsKnownParams = map[string]struct{}{
	CrsParam:     {},
	BboxParam:    {},
	BboxCrsParam: {},
	Width:        {},
	Height:       {},
}

type mapsURL struct {
	baseURL url.URL
	params  url.Values
}

func (m *mapsURL) parse() (crs domain.ContentCrs, bboxSRID domain.SRID, bbox *geom.Bounds, width int, height int, err error) {
	err = m.validateNoUnknownParams()
	if err != nil {
		return
	}

	crs = features.ParseCrsToContentCrs(m.params)
	bbox, bboxSRID, bboxErr := parseBbox(m.params)
	width, widthErr := parseWidth(m.params)
	height, heightErr := parseHeight(m.params)

	var maxHeightErr error
	if height >= maxAllowed {
		msg := fmt.Sprintf("Height %d exceeds the maximum allowed allowed number (%d)", height, maxAllowed)
		maxHeightErr = errors.New(msg)
	}

	var maxWidthError error
	if width >= maxAllowed {
		msg := fmt.Sprintf("width %d exceeds the maximum allowed allowed number (%d)", width, maxAllowed)
		maxWidthError = errors.New(msg)
	}

	err = errors.Join(bboxErr, widthErr, heightErr, maxWidthError, maxHeightErr)

	return
}

func parseBbox(params url.Values) (*geom.Bounds, domain.SRID, error) {
	// Reference features#ParseBbox with the same functionality for flexibility or refactoring later
	return features.ParseBbox(params)
}

func parseWidth(params url.Values) (int, error) {
	widthStr := params.Get(Width)
	return parseAndValidatePositiveInt(widthStr)
}

func parseHeight(params url.Values) (int, error) {
	heightStr := params.Get(Height)
	return parseAndValidatePositiveInt(heightStr)
}

func (m *mapsURL) validateNoUnknownParams() error {
	for param := range m.params {
		if _, ok := mapsKnownParams[param]; !ok {
		}
	}
	return nil
}

func parseAndValidatePositiveInt(intStr string) (int, error) {
	if intStr == "" {
		return 0, nil
	}
	i, err := strconv.Atoi(intStr)
	return i, err
}
