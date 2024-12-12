package search

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PDOK/gomagpie/internal/engine"
	ds "github.com/PDOK/gomagpie/internal/search/datasources"
	"github.com/PDOK/gomagpie/internal/search/domain"
)

const (
	queryParam = "q"
	limitParam = "limit"
	crsParam   = "crs"

	limitDefault = 10
	limitMax     = 50
)

var (
	deepObjectParamRegex = regexp.MustCompile(`\w+\[\w+]`)
)

func parseQueryParams(query url.Values) (collections ds.CollectionsWithParams, searchTerm string, outputSRID domain.SRID, limit int, err error) {
	err = validateNoUnknownParams(query)
	if err != nil {
		return
	}
	searchTerm, searchTermErr := parseSearchTerm(query)
	collections = parseDeepObjectParams(query)
	if len(collections) == 0 {
		return nil, "", 0, 0, errors.New(
			"no collection(s) specified in request, specify at least one collection and version. " +
				"For example: foo[version]=1&bar[version]=2 where 'foo' and 'bar' are collection names")
	}
	outputSRID, outputSRIDErr := parseCrsToSRID(query, crsParam)
	limit, limitErr := parseLimit(query)
	err = errors.Join(searchTermErr, limitErr, outputSRIDErr)
	return
}

// Parse "deep object" params, e.g. paramName[prop1]=value1&paramName[prop2]=value2&....
func parseDeepObjectParams(query url.Values) ds.CollectionsWithParams {
	deepObjectParams := make(ds.CollectionsWithParams, len(query))
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
	return deepObjectParams
}

func parseSearchTerm(query url.Values) (searchTerm string, err error) {
	searchTerm = query.Get(queryParam)
	if searchTerm == "" {
		err = fmt.Errorf("no search term provided, '%s' query parameter is required", queryParam)
	}
	return
}

// implements req 7.6 (https://docs.ogc.org/is/17-069r4/17-069r4.html#query_parameters)
func validateNoUnknownParams(query url.Values) error {
	copyParams := clone(query)
	copyParams.Del(engine.FormatParam)
	copyParams.Del(queryParam)
	copyParams.Del(limitParam)
	copyParams.Del(crsParam)
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

func parseCrsToSRID(params url.Values, paramName string) (domain.SRID, error) {
	param := params.Get(paramName)
	if param == "" {
		return domain.UndefinedSRID, nil
	}
	param = strings.TrimSpace(param)
	if !strings.HasPrefix(param, domain.CrsURIPrefix) {
		return domain.UndefinedSRID, fmt.Errorf("%s param should start with %s, got: %s", paramName, domain.CrsURIPrefix, param)
	}
	var srid domain.SRID
	lastIndex := strings.LastIndex(param, "/")
	if lastIndex != -1 {
		crsCode := param[lastIndex+1:]
		if crsCode == domain.WGS84CodeOGC {
			return domain.WGS84SRIDPostgis, nil // CRS84 is WGS84, just like EPSG:4326 (only axis order differs but SRID is the same)
		}
		val, err := strconv.Atoi(crsCode)
		if err != nil {
			return 0, fmt.Errorf("expected numerical CRS code, received: %s", crsCode)
		}
		srid = domain.SRID(val)
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
