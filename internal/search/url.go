package search

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/engine/util"
	d "github.com/PDOK/gokoala/internal/search/domain"
	"github.com/twpayne/go-geom"
)

const (
	queryParam   = "q"
	limitParam   = "limit"
	crsParam     = "crs"
	bboxParam    = "bbox"
	bboxCrsParam = "bbox-crs"

	limitDefault = 10
	limitMax     = 50
)

var (
	deepObjectParamRegex = regexp.MustCompile(`\w+\[\w+]`)

	// matches & (AND), | (OR), ! (NOT), and <-> (FOLLOWED BY).
	searchOperatorsRegex = regexp.MustCompile(`&|\||!|<->`)
)

func parseQueryParams(query url.Values) (collections d.CollectionsWithParams, searchTerms string, outputSRID d.SRID, outputCRS string, bbox *geom.Bounds, bboxSRID d.SRID, limit int, err error) {
	err = validateNoUnknownParams(query)
	if err != nil {
		return
	}
	searchTerms, searchTermErr := parseSearchTerms(query)
	collections, collErr := parseCollections(query)
	outputSRID, outputSRIDErr := parseCrsToPostgisSRID(query, crsParam)
	outputCRS = query.Get(crsParam)
	limit, limitErr := parseLimit(query)
	bbox, bboxSRID, bboxErr := parseBbox(query)

	err = errors.Join(collErr, searchTermErr, limitErr, outputSRIDErr, bboxErr)
	return
}

// Parse collections as "deep object" params, e.g. collectionName[prop1]=value1&collectionName[prop2]=value2&....
func parseCollections(query url.Values) (d.CollectionsWithParams, error) {
	deepObjectParams := make(d.CollectionsWithParams, len(query))
	for key, values := range query {
		if strings.Contains(key, "[") {
			// Extract deepObject parameters
			parts := strings.SplitN(key, "[", 2)
			mainKey := parts[0]
			subKey := strings.TrimSuffix(parts[1], "]")

			if _, exists := deepObjectParams[mainKey]; !exists {
				deepObjectParams[mainKey] = make(map[string]string)
			}
			deepObjectParams[mainKey][subKey] = values[0]
		}
	}
	errMsg := "specify at least one collection and version. For example: 'foo[version]=1' where 'foo' is the collection and '1' the version"
	if len(deepObjectParams) == 0 {
		return nil, fmt.Errorf("no collection(s) specified in request, %s", errMsg)
	}
	for name := range deepObjectParams {
		if version, ok := deepObjectParams[name][d.VersionParam]; !ok || version == "" {
			return nil, fmt.Errorf("no version specified in request for collection %s, %s", name, errMsg)
		}
	}
	return deepObjectParams, nil
}

func parseSearchTerms(query url.Values) (string, error) {
	searchTerms := strings.TrimSpace(strings.ToLower(query.Get(queryParam)))
	if searchTerms == "" {
		return "", fmt.Errorf("no search terms provided, '%s' query parameter is required", queryParam)
	}
	if searchOperatorsRegex.MatchString(searchTerms) {
		return "", errors.New("provided search terms contain one ore more boolean operators " +
			"such as & (AND), | (OR), ! (NOT) which aren't allowed")
	}
	return searchTerms, nil
}

// implements req 7.6 (https://docs.ogc.org/is/17-069r4/17-069r4.html#query_parameters)
func validateNoUnknownParams(query url.Values) error {
	copyParams := clone(query)
	copyParams.Del(engine.FormatParam)
	copyParams.Del(queryParam)
	copyParams.Del(limitParam)
	copyParams.Del(crsParam)
	copyParams.Del(bboxParam)
	copyParams.Del(bboxCrsParam)
	for key := range query {
		if deepObjectParamRegex.MatchString(key) {
			copyParams.Del(key)
		}
	}
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

func parseCrsToPostgisSRID(params url.Values, paramName string) (d.SRID, error) {
	param := params.Get(paramName)
	if param == "" {
		return d.WGS84SRIDPostgis, nil // default to WGS84
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
			return d.WGS84SRIDPostgis, nil // CRS84 is WGS84, we use EPSG:4326 for Postgres TODO: check if correct since axis order differs between CRS84 and EPSG:4326
		}
		val, err := strconv.Atoi(crsCode)
		if err != nil {
			return 0, fmt.Errorf("expected numerical CRS code, received: %s", crsCode)
		}
		srid = d.SRID(val)
	}
	return srid, nil
}

func parseLimit(params url.Values) (int, error) {
	limit := limitDefault
	var err error
	if params.Get(limitParam) != "" {
		limit, err = strconv.Atoi(params.Get(limitParam))
		if err != nil {
			err = errors.New("limit must be numeric")
		}
		// "If the value of the limit parameter is larger than the maximum value, this SHALL NOT result
		//  in an error (instead use the maximum as the parameter value)."
		if limit > limitMax {
			limit = limitMax
		}
	}
	if limit < 0 {
		err = errors.New("limit can't be negative")
	}
	return limit, err
}

func parseBbox(params url.Values) (*geom.Bounds, d.SRID, error) {
	bboxSRID, err := parseCrsToPostgisSRID(params, bboxCrsParam)
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
	if util.SurfaceArea(bbox) <= 0 {
		return nil, bboxSRID, errors.New("bbox has no surface area")
	}
	return bbox, bboxSRID, nil
}
