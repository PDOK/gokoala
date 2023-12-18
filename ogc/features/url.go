package features

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/go-spatial/geom"
)

const (
	cursorParam    = "cursor"
	limitParam     = "limit"
	crsParam       = "crs"
	dateTimeParam  = "datetime"
	bboxParam      = "bbox"
	bboxCrsParam   = "bbox-crs"
	filterParam    = "filter"
	filterCrsParam = "filter-crs"
)

var (
	checksumExcludedParams = []string{engine.FormatParam, cursorParam} // don't include these in checksum
)

// SRID Spatial Reference System Identifier: a unique value to unambiguously identify a spatial coordinate system.
// For example '28992' in https://www.opengis.net/def/crs/EPSG/0/28992
type SRID int

func (s SRID) GetOrDefault() int {
	val := int(s)
	if val <= 0 {
		return wgs84SRID
	}
	return val
}

type ContentCrs string

// ToLink returns link target conforming to RFC 8288
func (c ContentCrs) ToLink() string {
	return fmt.Sprintf("<%s>", c)
}

func (c ContentCrs) IsWGS84() bool {
	return string(c) == wgs84CrsURI
}

type URL interface {
	validateNoUnknownParams() error
}

// URL to a page in a collection of features
type featureCollectionURL struct {
	baseURL url.URL
	params  url.Values
	limit   engine.Limit
}

// parse the given URL to values required to delivery a set of Features
func (fc featureCollectionURL) parse() (encodedCursor domain.EncodedCursor, limit int,
	inputSRID SRID, outputSRID SRID, contentCrs ContentCrs, bbox *geom.Extent, err error) {

	encodedCursor = domain.EncodedCursor(fc.params.Get(cursorParam))
	limit, limitErr := parseLimit(fc.params, fc.limit)
	outputSRID, outputSRIDErr := parseCrsToSRID(fc.params, crsParam)
	contentCrs = parseCrsToContentCrs(fc.params)
	bbox, bboxSRID, bboxErr := parseBbox(fc.params)
	dateTimeErr := parseDateTime(fc.params)
	_, filterSRID, filterErr := parseFilter(fc.params)

	inputSRID, inputSRIDErr := consolidateSRIDs(bboxSRID, filterSRID)

	err = errors.Join(limitErr, outputSRIDErr, bboxErr, dateTimeErr, filterErr, inputSRIDErr)
	return
}

// Calculate checksum over the query parameters that have a "filtering effect" on
// the result set such as limit, bbox, CQL filters, etc. These query params
// aren't allowed to be changed during pagination. The checksum allows for the latter
// to be verified
func (fc featureCollectionURL) checksum() []byte {
	var valuesToHash bytes.Buffer
	sortedQueryParams := make([]string, 0, len(fc.params))
	for k := range fc.params {
		sortedQueryParams = append(sortedQueryParams, k)
	}
	sort.Strings(sortedQueryParams) // sort keys
OUTER:
	for _, k := range sortedQueryParams {
		for _, skip := range checksumExcludedParams {
			if k == skip {
				continue OUTER
			}
		}
		paramValues := fc.params[k]
		if paramValues != nil {
			slices.Sort(paramValues) // sort values belonging to key
		}
		for _, s := range paramValues {
			valuesToHash.WriteString(s)
		}
	}

	bytesToHash := valuesToHash.Bytes()
	if len(bytesToHash) > 0 {
		hasher := fnv.New32a() // fast non-cryptographic hash
		hasher.Write(bytesToHash)
		return hasher.Sum(nil)
	}
	return []byte{}
}

func (fc featureCollectionURL) toSelfURL(collectionID string, format string) string {
	newParams := url.Values{}
	newParams.Set(engine.FormatParam, format)

	result := fc.baseURL.JoinPath("collections", collectionID, "items")
	result.RawQuery = newParams.Encode()
	return result.String()
}

func (fc featureCollectionURL) toPrevNextURL(collectionID string, cursor domain.EncodedCursor, format string) string {
	copyParams := clone(fc.params)
	copyParams.Set(engine.FormatParam, format)
	copyParams.Set(cursorParam, cursor.String())

	result := fc.baseURL.JoinPath("collections", collectionID, "items")
	result.RawQuery = copyParams.Encode()
	return result.String()
}

// implements req 7.6 (https://docs.ogc.org/is/17-069r4/17-069r4.html#query_parameters)
func (fc featureCollectionURL) validateNoUnknownParams() error {
	copyParams := clone(fc.params)
	copyParams.Del(engine.FormatParam)
	copyParams.Del(limitParam)
	copyParams.Del(cursorParam)
	copyParams.Del(crsParam)
	copyParams.Del(dateTimeParam)
	copyParams.Del(bboxParam)
	copyParams.Del(bboxCrsParam)
	copyParams.Del(filterParam)
	copyParams.Del(filterCrsParam)
	if len(copyParams) > 0 {
		return fmt.Errorf("unknown query parameter(s) found: %v", copyParams.Encode())
	}
	return nil
}

// URL to a specific Feature
type featureURL struct {
	baseURL url.URL
	params  url.Values
}

// parse the given URL to values required to delivery a specific Feature
func (f featureURL) parse() (srid SRID, contentCrs ContentCrs, err error) {
	srid, err = parseCrsToSRID(f.params, crsParam)
	contentCrs = parseCrsToContentCrs(f.params)
	return
}

func (f featureURL) toSelfURL(collectionID string, featureID int64, format string) string {
	newParams := url.Values{}
	newParams.Set(engine.FormatParam, format)

	result := f.baseURL.JoinPath("collections", collectionID, "items", strconv.FormatInt(featureID, 10))
	result.RawQuery = newParams.Encode()
	return result.String()
}

func (f featureURL) toCollectionURL(collectionID string, format string) string {
	newParams := url.Values{}
	newParams.Set(engine.FormatParam, format)

	result := f.baseURL.JoinPath("collections", collectionID)
	result.RawQuery = newParams.Encode()
	return result.String()
}

// implements req 7.6 (https://docs.ogc.org/is/17-069r4/17-069r4.html#query_parameters)
func (f featureURL) validateNoUnknownParams() error {
	copyParams := clone(f.params)
	copyParams.Del(engine.FormatParam)
	copyParams.Del(crsParam)
	if len(copyParams) > 0 {
		return fmt.Errorf("unknown query parameter(s) found: %v", copyParams.Encode())
	}
	return nil
}

func clone(params url.Values) url.Values {
	copyParams := url.Values{}
	for k, v := range params {
		copyParams[k] = v
	}
	return copyParams
}

func consolidateSRIDs(bboxSRID SRID, filterSRID SRID) (inputSRID SRID, err error) {
	if bboxSRID != undefinedSRID && filterSRID != undefinedSRID && bboxSRID != filterSRID {
		return 0, errors.New("bbox-crs and filter-crs need to be equal. " +
			"Can't use more than one CRS as input, but input and output CRS may differ")
	}
	if bboxSRID != undefinedSRID || filterSRID != undefinedSRID {
		inputSRID = bboxSRID // or filterCrs, both the same
	}
	return inputSRID, err
}

func parseLimit(params url.Values, limitCfg engine.Limit) (int, error) {
	limit := limitCfg.Default
	var err error
	if params.Get(limitParam) != "" {
		limit, err = strconv.Atoi(params.Get(limitParam))
		if err != nil {
			err = fmt.Errorf("limit must be numeric")
		}
		// OpenAPI validation already guards against exceeding max limit, this is just a defense in-depth measure.
		if limit > limitCfg.Max {
			limit = limitCfg.Max
		}
	}
	if limit < 0 {
		err = fmt.Errorf("limit can't be negative")
	}
	return limit, err
}

func parseBbox(params url.Values) (*geom.Extent, SRID, error) {
	bboxSRID, err := parseCrsToSRID(params, bboxCrsParam)
	if err != nil {
		return nil, undefinedSRID, err
	}

	if params.Get(bboxParam) == "" {
		return nil, undefinedSRID, nil
	}
	bboxValues := strings.Split(params.Get(bboxParam), ",")
	if len(bboxValues) != 4 {
		return nil, bboxSRID, fmt.Errorf("bbox should contain exactly 4 values " +
			"separated by commas: minx,miny,maxx,maxy")
	}

	var extent geom.Extent
	for i, v := range bboxValues {
		extent[i], err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, bboxSRID, fmt.Errorf("failed to parse value %s in bbox, error: %w", v, err)
		}
	}

	return &extent, bboxSRID, nil
}

func parseCrsToContentCrs(params url.Values) ContentCrs {
	param := params.Get(crsParam)
	if param == "" {
		return wgs84CrsURI
	}
	return ContentCrs(param)
}

func parseCrsToSRID(params url.Values, paramName string) (SRID, error) {
	param := params.Get(paramName)
	if param == "" {
		return undefinedSRID, nil
	}
	param = strings.TrimSpace(param)
	if !strings.HasPrefix(param, crsURIPrefix) {
		return undefinedSRID, fmt.Errorf("%s param should start with %s, got: %s", paramName, crsURIPrefix, param)
	}
	var srid SRID
	lastIndex := strings.LastIndex(param, "/")
	if lastIndex != -1 {
		crsCode := param[lastIndex+1:]
		if crsCode == wgs84CodeOGC {
			return wgs84SRID, nil // CRS84 is WGS84, just like EPSG:4326 (only axis order differs but SRID is the same)
		}
		val, err := strconv.Atoi(crsCode)
		if err != nil {
			return 0, fmt.Errorf("expected numerical CRS code, received: %s", crsCode)
		}
		srid = SRID(val)
	}
	return srid, nil
}

func parseDateTime(params url.Values) error {
	if params.Get(dateTimeParam) != "" {
		return fmt.Errorf("datetime param is currently not supported")
	}
	return nil
}

func parseFilter(params url.Values) (filter string, filterSRID SRID, err error) {
	filter = params.Get(filterParam)
	filterSRID, _ = parseCrsToSRID(params, filterCrsParam)

	if filter != "" {
		return filter, filterSRID, fmt.Errorf("CQL filter param is currently not supported")
	}
	return filter, filterSRID, nil
}
