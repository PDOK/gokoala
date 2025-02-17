package domain

import (
	"sort"
	"strconv"
	"strings"
)

const (
	VersionParam     = "version"
	RelevanceParam   = "relevance"
	DefaultRelevance = 0.5
)

// GeoJSON properties in search response
const (
	PropCollectionID      = "collectionId"
	PropCollectionVersion = "collectionVersion"
	PropGeomType          = "collectionGeometryType"
	PropDisplayName       = "displayName"
	PropHighlight         = "highlight"
	PropScore             = "score"
	PropHref              = "href"
)

type SearchQuery struct {
	terms []string
}

func NewSearchQuery(terms []string) SearchQuery {
	sort.Strings(terms)
	return SearchQuery{terms: terms}
}

func (q *SearchQuery) ToWildcardQuery() string {
	return q.toString(true)
}

func (q *SearchQuery) ToExactMatchQuery() string {
	return q.toString(false)
}

func (q *SearchQuery) toString(wildcard bool) string {
	sb := &strings.Builder{}
	for i, term := range q.terms {
		sb.WriteByte('(')
		parts := strings.Fields(term)
		for j, part := range parts {
			sb.WriteString(part)
			if wildcard {
				sb.WriteString(":*")
			}
			if j != len(parts)-1 {
				sb.WriteString(" & ")
			}
		}
		sb.WriteByte(')')
		if i != len(q.terms)-1 {
			sb.WriteString(" | ")
		}
	}
	return sb.String()
}

// CollectionsWithParams collection name with associated CollectionParams
// These are provided though a URL query string as "deep object" params, e.g. paramName[prop1]=value1&paramName[prop2]=value2&....
type CollectionsWithParams map[string]CollectionParams

// CollectionParams parameter key with associated value
type CollectionParams map[string]string

func (cp CollectionsWithParams) NamesAndVersionsAndRelevance() (names []string, versions []int, relevance []float64) {
	for name := range cp {
		version, ok := cp[name][VersionParam]
		if !ok {
			continue
		}
		versionNr, err := strconv.Atoi(version)
		if err != nil {
			continue
		}

		relevanceRaw, ok := cp[name][RelevanceParam]
		if ok {
			relevanceFloat, err := strconv.ParseFloat(relevanceRaw, 64)
			if err == nil && relevanceFloat >= 0 && relevanceFloat <= 1 {
				relevance = append(relevance, relevanceFloat)
			} else {
				relevance = append(relevance, DefaultRelevance)
			}
		} else {
			relevance = append(relevance, DefaultRelevance)
		}

		versions = append(versions, versionNr)
		names = append(names, name)
	}
	return names, versions, relevance
}
