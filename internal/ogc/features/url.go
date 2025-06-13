package features

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources"
	d "github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/twpayne/go-geom"
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
	profileParam   = "profile"

	propertyFilterMaxLength = 512
	propertyFilterWildcard  = "*"
)

var (
	// don't include these params in checksum
	checksumExcludedParams = []string{
		engine.FormatParam,
		cursorParam,
	}

	// request for /items should only accept these params (+ filters)
	// defined as a set for fast lookup
	featuresKnownParams = map[string]struct{}{
		engine.FormatParam: {},
		limitParam:         {},
		cursorParam:        {},
		crsParam:           {},
		dateTimeParam:      {},
		bboxParam:          {},
		bboxCrsParam:       {},
		filterParam:        {},
		filterCrsParam:     {},
		profileParam:       {},
	}

	// request for /item/{id} should only accept these params
	// defined as a set for fast lookup
	featureKnownParams = map[string]struct{}{
		engine.FormatParam: {},
		crsParam:           {},
		profileParam:       {},
	}
)

// URL to a page in a collection of features
type featureCollectionURL struct {
	baseURL                   url.URL
	params                    url.Values
	limit                     config.Limit
	configuredPropertyFilters map[string]datasources.PropertyFilterWithAllowedValues
	configuredCollectionNames []string
	supportsDatetime          bool
}

// parse the given URL to values required to delivery a set of Features
func (fc featureCollectionURL) parse() (encodedCursor d.EncodedCursor, limit int, inputSRID d.SRID, outputSRID d.SRID,
	contentCrs d.ContentCrs, bbox *geom.Bounds, referenceDate time.Time, propertyFilters map[string]string,
	profile d.Profile, err error) {

	err = fc.validateNoUnknownParams()
	if err != nil {
		return
	}
	encodedCursor = d.EncodedCursor(fc.params.Get(cursorParam))
	limit, limitErr := parseLimit(fc.params, fc.limit)
	outputSRID, outputSRIDErr := parseCrsToSRID(fc.params, crsParam)
	contentCrs = parseCrsToContentCrs(fc.params)
	propertyFilters, pfErr := parsePropertyFilters(fc.configuredPropertyFilters, fc.params)
	bbox, bboxSRID, bboxErr := parseBbox(fc.params)
	profile = parseProfile(fc.params, fc.baseURL, fc.configuredCollectionNames)
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
		_, _ = hasher.Write(bytesToHash)
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

func (fc featureCollectionURL) toPrevNextURL(collectionID string, cursor d.EncodedCursor, format string) string {
	copyParams := clone(fc.params)
	copyParams.Set(engine.FormatParam, format)
	copyParams.Set(cursorParam, cursor.String())

	result := fc.baseURL.JoinPath("collections", collectionID, "items")
	result.RawQuery = copyParams.Encode()
	return result.String()
}

// implements req 7.6 (https://docs.ogc.org/is/17-069r4/17-069r4.html#query_parameters)
func (fc featureCollectionURL) validateNoUnknownParams() error {
	for param := range fc.params {
		if _, ok := featuresKnownParams[param]; !ok {
			if _, configured := fc.configuredPropertyFilters[param]; !configured {
				return fmt.Errorf("unknown query parameter(s) found: %s", param)
			}
		}
	}
	return nil
}

// URL to a specific Feature
type featureURL struct {
	baseURL                   url.URL
	params                    url.Values
	configuredCollectionNames []string
}

// parse the given URL to values required to delivery a specific Feature
func (f featureURL) parse() (srid d.SRID, contentCrs d.ContentCrs, profile d.Profile, err error) {
	err = f.validateNoUnknownParams()
	if err != nil {
		return
	}

	srid, err = parseCrsToSRID(f.params, crsParam)
	contentCrs = parseCrsToContentCrs(f.params)
	profile = parseProfile(f.params, f.baseURL, f.configuredCollectionNames)
	return
}

func (f featureURL) toSelfURL(collectionID string, featureID string, format string) string {
	newParams := url.Values{}
	newParams.Set(engine.FormatParam, format)

	result := f.baseURL.JoinPath("collections", collectionID, "items", featureID)
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
	for param := range f.params {
		if _, ok := featureKnownParams[param]; !ok {
			return fmt.Errorf("unknown query parameter(s) found: %s", param)
		}
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

func consolidateSRIDs(bboxSRID d.SRID, filterSRID d.SRID) (inputSRID d.SRID, err error) {
	if bboxSRID != d.UndefinedSRID && filterSRID != d.UndefinedSRID && bboxSRID != filterSRID {
		return 0, errors.New("bbox-crs and filter-crs need to be equal. " +
			"Can't use more than one CRS as input, but input and output CRS may differ")
	}
	if bboxSRID != d.UndefinedSRID || filterSRID != d.UndefinedSRID {
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

func parseBbox(params url.Values) (*geom.Bounds, d.SRID, error) {
	bboxSRID, err := parseCrsToSRID(params, bboxCrsParam)
	if err != nil {
		return nil, d.UndefinedSRID, err
	}

	if params.Get(bboxParam) == "" {
		return nil, d.UndefinedSRID, nil
	}
	bboxValues := strings.Split(params.Get(bboxParam), ",")
	if len(bboxValues) != 4 {
		return nil, bboxSRID, errors.New("bbox should contain exactly 4 values " +
			"separated by commas: minx,miny,maxx,maxy")
	}

	bboxFloats := make([]float64, len(bboxValues))
	for i, v := range bboxValues {
		bboxFloats[i], err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, bboxSRID, fmt.Errorf("failed to parse value %s in bbox, error: %w", v, err)
		}
	}

	bbox := geom.NewBounds(geom.XY).Set(bboxFloats...)
	if surfaceArea(bbox) <= 0 {
		return nil, bboxSRID, errors.New("bbox has no surface area")
	}
	return bbox, bboxSRID, nil
}

func surfaceArea(bbox *geom.Bounds) float64 {
	// Use the same logic as bbox.Area() in https://github.com/go-spatial/geom to calculate surface area.
	// The bounds.Area() in github.com/twpayne/go-geom behaves differently and is not what we're looking for.
	return math.Abs((bbox.Max(1) - bbox.Min(1)) * (bbox.Max(0) - bbox.Min(0)))
}

func parseCrsToContentCrs(params url.Values) d.ContentCrs {
	param := params.Get(crsParam)
	if param == "" {
		return d.WGS84CrsURI
	}
	return d.ContentCrs(param)
}

func parseCrsToSRID(params url.Values, paramName string) (d.SRID, error) {
	param := params.Get(paramName)
	if param == "" {
		return d.UndefinedSRID, nil
	}
	param = strings.TrimSpace(param)
	if !strings.HasPrefix(param, d.CrsURIPrefix) {
		return d.UndefinedSRID, fmt.Errorf("%s param should start with %s, got: %s", paramName, d.CrsURIPrefix, param)
	}
	var srid d.SRID
	lastIndex := strings.LastIndex(param, "/")
	if lastIndex != -1 {
		crsCode := param[lastIndex+1:]
		if crsCode == d.WGS84CodeOGC {
			return d.WGS84SRID, nil // CRS84 is WGS84, just like EPSG:4326 (only axis order differs but SRID is the same)
		}
		val, err := strconv.Atoi(crsCode)
		if err != nil {
			return 0, fmt.Errorf("expected numerical CRS code, received: %s", crsCode)
		}
		srid = d.SRID(val)
	}
	return srid, nil
}

// Support simple filtering on properties: https://docs.ogc.org/is/17-069r4/17-069r4.html#_parameters_for_filtering_on_feature_properties
func parsePropertyFilters(configuredPropertyFilters map[string]datasources.PropertyFilterWithAllowedValues, params url.Values) (map[string]string, error) {
	propertyFilters := make(map[string]string)
	for name := range configuredPropertyFilters {
		pf := params.Get(name)
		if pf != "" {
			if len(pf) > propertyFilterMaxLength {
				return nil, fmt.Errorf("property filter %s is too large, "+
					"value is limited to %d characters", name, propertyFilterMaxLength)
			}
			if strings.Contains(pf, propertyFilterWildcard) {
				// if/when we choose to support wildcards in the future, make sure wildcards are
				// only allowed at the END (suffix) of the filter
				return nil, fmt.Errorf("property filter %s contains a wildcard (%s), "+
					"wildcard filtering is not allowed", name, propertyFilterWildcard)
			}
			propertyFilters[name] = pf
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

func parseFilter(params url.Values) (filter string, filterSRID d.SRID, err error) {
	filter = params.Get(filterParam)
	filterSRID, _ = parseCrsToSRID(params, filterCrsParam)

	if filter != "" {
		return filter, filterSRID, errors.New("CQL filter param is currently not supported")
	}
	return filter, filterSRID, nil
}

func parseProfile(params url.Values, baseURL url.URL, configuredCollectionNames []string) d.Profile {
	profile := d.RelAsLink
	if params.Has(profileParam) {
		profile = d.ProfileName(params.Get(profileParam))
	}
	return d.NewProfile(profile, baseURL, configuredCollectionNames)
}
