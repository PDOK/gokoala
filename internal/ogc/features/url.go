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
	"time"

	"github.com/PDOK/gokoala/config"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
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

	propertyFilterMaxLength = 512
	propertyFilterWildcard  = "*"
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

// ContentCrs the coordinate reference system (represented as a URI) of the content/output to return.
type ContentCrs string

// ToLink returns link target conforming to RFC 8288
func (c ContentCrs) ToLink() string {
	return fmt.Sprintf("<%s>", c)
}

func (c ContentCrs) IsWGS84() bool {
	return string(c) == wgs84CrsURI
}

// URL to a page in a collection of features
type featureCollectionURL struct {
	baseURL                   url.URL
	params                    url.Values
	limit                     config.Limit
	configuredPropertyFilters []config.PropertyFilter
	supportsDatetime          bool
}

// parse the given URL to values required to delivery a set of Features
func (fc featureCollectionURL) parse() (encodedCursor domain.EncodedCursor, limit int, inputSRID SRID, outputSRID SRID,
	contentCrs ContentCrs, bbox *geom.Extent, referenceDate time.Time, propertyFilters map[string]string, err error) {

	err = fc.validateNoUnknownParams()
	if err != nil {
		return
	}
	encodedCursor = domain.EncodedCursor(fc.params.Get(cursorParam))
	limit, limitErr := parseLimit(fc.params, fc.limit)
	outputSRID, outputSRIDErr := parseCrsToSRID(fc.params, crsParam)
	contentCrs = parseCrsToContentCrs(fc.params)
	propertyFilters, pfErr := parsePropertyFilters(fc.configuredPropertyFilters, fc.params)
	bbox, bboxSRID, bboxErr := parseBbox(fc.params)
	referenceDate, dateTimeErr := parseDateTime(fc.params, fc.supportsDatetime)
	_, filterSRID, filterErr := parseFilter(fc.params)
	inputSRID, inputSRIDErr := consolidateSRIDs(bboxSRID, filterSRID)

	err = errors.Join(limitErr, outputSRIDErr, bboxErr, pfErr, dateTimeErr, filterErr, inputSRIDErr)
	return
}

// Calculate checksum over the query parameters that have a "filtering effect" on
// the result set such as limit, bbox, property filters, CQL filters, etc. These query params
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
	copyParams := clone(fc.params)
	copyParams.Set(engine.FormatParam, format)

	result := fc.baseURL.JoinPath("collections", collectionID, "items")
	result.RawQuery = copyParams.Encode()
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
	for _, pf := range fc.configuredPropertyFilters {
		copyParams.Del(pf.Name)
	}
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
	err = f.validateNoUnknownParams()
	if err != nil {
		return
	}

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

func parseLimit(params url.Values, limitCfg config.Limit) (int, error) {
	limit := limitCfg.Default
	var err error
	if params.Get(limitParam) != "" {
		limit, err = strconv.Atoi(params.Get(limitParam))
		if err != nil {
			err = errors.New("limit must be numeric")
		}
		// "If the value of the limit parameter is larger than the maximum value, this SHALL NOT result
		//  in an error (instead use the maximum as the parameter value)."
		if limit > limitCfg.Max {
			limit = limitCfg.Max
		}
	}
	if limit < 0 {
		err = errors.New("limit can't be negative")
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
		return nil, bboxSRID, errors.New("bbox should contain exactly 4 values " +
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

// Support simple filtering on properties: https://docs.ogc.org/is/17-069r4/17-069r4.html#_parameters_for_filtering_on_feature_properties
func parsePropertyFilters(configuredPropertyFilters []config.PropertyFilter, params url.Values) (map[string]string, error) {
	propertyFilters := make(map[string]string)
	for _, cpf := range configuredPropertyFilters {
		pf := params.Get(cpf.Name)
		if pf != "" {
			if len(pf) > propertyFilterMaxLength {
				return nil, fmt.Errorf("property filter %s is too large, "+
					"value is limited to %d characters", cpf.Name, propertyFilterMaxLength)
			}
			if strings.Contains(pf, propertyFilterWildcard) {
				// if/when we choose to support wildcards in the future, make sure wildcards are
				// only allowed at the END (suffix) of the filter
				return nil, fmt.Errorf("property filter %s contains a wildcard (%s), "+
					"wildcard filtering is not allowed", cpf.Name, propertyFilterWildcard)
			}
			propertyFilters[cpf.Name] = pf
		}
	}
	return propertyFilters, nil
}

// Support filtering on datetime: https://docs.ogc.org/is/17-069r4/17-069r4.html#_parameter_datetime
func parseDateTime(params url.Values, datetimeSupported bool) (time.Time, error) {
	datetime := params.Get(dateTimeParam)
	if datetime != "" {
		if !datetimeSupported {
			return time.Time{}, errors.New("datetime param is currently not supported for this collection")
		}
		if strings.Contains(datetime, "/") {
			return time.Time{}, fmt.Errorf("datetime param '%s' represents an interval, intervals are currently not supported", datetime)
		}
		return time.Parse(time.RFC3339, datetime)
	}
	return time.Time{}, nil
}

func parseFilter(params url.Values) (filter string, filterSRID SRID, err error) {
	filter = params.Get(filterParam)
	filterSRID, _ = parseCrsToSRID(params, filterCrsParam)

	if filter != "" {
		return filter, filterSRID, errors.New("CQL filter param is currently not supported")
	}
	return filter, filterSRID, nil
}
