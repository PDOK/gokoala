package features_search

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features"
	fd "github.com/PDOK/gokoala/internal/ogc/features/domain"
	d "github.com/PDOK/gokoala/internal/ogc/features_search/domain"
	"github.com/twpayne/go-geom"
)

const (
	queryParam = "q"

	limitDefault = 10
	limitMax     = 50
)

var (
	deepObjectParamRegex = regexp.MustCompile(`\w+\[\w+\]`)

	// matches & (AND), | (OR), ! (NOT), and <-> (FOLLOWED BY).
	searchOperatorsRegex = regexp.MustCompile(`&|\||!|<->`)
	// matches ' (apostrophe), ( (left parenthesis), and ) (right parenthesis).
	searchDiscardCharactersRegex = regexp.MustCompile(`'|\(|\)`)

	searchKnownParams = map[string]struct{}{
		queryParam:            {},
		engine.FormatParam:    {},
		features.LimitParam:   {},
		features.CrsParam:     {},
		features.BboxParam:    {},
		features.BboxCrsParam: {},
	}
)

func parseQueryParams(query url.Values) (collections d.CollectionsWithParams, searchTerms string,
	outputSRID fd.SRID, contentCrs fd.ContentCrs, bbox *geom.Bounds, bboxSRID fd.SRID, limit int, err error) {

	err = validateNoUnknownParams(query)
	if err != nil {
		return
	}
	searchTerms, searchTermErr := parseSearchTerms(query)
	collections, collErr := parseCollections(query)
	outputSRID, outputSRIDErr := features.ParseCrsToSRID(query, features.CrsParam)
	contentCrs = features.ParseCrsToContentCrs(query)
	limit, limitErr := features.ParseLimit(query, config.Limit{
		Default: limitDefault,
		Max:     limitMax,
	})
	bbox, bboxSRID, bboxErr := features.ParseBbox(query)

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
	searchTerms := searchDiscardCharactersRegex.ReplaceAllLiteralString(strings.TrimSpace(strings.ToLower(query.Get(queryParam))), "")
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
	for param := range query {
		if deepObjectParamRegex.MatchString(param) {
			continue
		}
		if _, ok := searchKnownParams[param]; !ok {
			return fmt.Errorf("unknown query parameter(s) found: %s", param)
		}
	}
	return nil
}
