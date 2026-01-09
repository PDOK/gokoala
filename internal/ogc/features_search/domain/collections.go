package domain

import "strconv"

const (
	VersionParam     = "version"
	RelevanceParam   = "relevance"
	DefaultRelevance = 0.5
)

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
