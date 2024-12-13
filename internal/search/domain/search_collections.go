package domain

import "strconv"

const (
	VersionParam = "version"
)

// CollectionsWithParams collection name with associated CollectionParams
// These are provided though a URL query string as "deep object" params, e.g. paramName[prop1]=value1&paramName[prop2]=value2&....
type CollectionsWithParams map[string]CollectionParams

// CollectionParams parameter key with associated value
type CollectionParams map[string]string

func (cp CollectionsWithParams) NamesAndVersions() (names []string, versions []int) {
	for name := range cp {
		version, ok := cp[name][VersionParam]
		if !ok {
			continue
		}
		versionNr, err := strconv.Atoi(version)
		if err != nil {
			continue
		}
		versions = append(versions, versionNr)
		names = append(names, name)
	}
	return names, versions
}
