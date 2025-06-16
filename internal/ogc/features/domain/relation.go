package domain

import (
	"log"
	"regexp"
	"sort"
	"strings"
)

const regexRemoveSeparators = "[^a-z0-9]?"

// FeatureRelation a relation/reference from one feature to another in a different
// collection, according to OAF Part 5: https://docs.ogc.org/DRAFTS/23-058r1.html#rc_feature-references.
type FeatureRelation struct {
	Name         string
	CollectionID string
}

func NewFeatureRelation(name, externalFidColumn string, collectionNames []string) *FeatureRelation {
	if !isFeatureRelation(name, externalFidColumn) {
		return nil
	}
	relationName := newFeatureRelationName(name, externalFidColumn)
	return &FeatureRelation{
		Name:         relationName,
		CollectionID: findReferencedCollection(collectionNames, relationName),
	}
}

// newFeatureRelationName derive name of the feature relation.
//
// In the datasource we have fields named 'foobar_external_fid' or 'foobar_sometext_external_fid' containing UUID's to
// features in the 'foobar' collection. The feature relation field name will be 'foobar' or 'foobar_sometext'. This is
// the name of expose in the feature data (GeoJSON) and the schema (JSON-Schema).
func newFeatureRelationName(name string, externalFidColumn string) string {
	regex, _ := regexp.Compile(regexRemoveSeparators + externalFidColumn + regexRemoveSeparators)
	return regex.ReplaceAllString(name, "")
}

// isFeatureRelation "Algorithm" to determine feature reference:
//
// When externalFidColumn (e.g. 'external_fid') is part of the column name (e.g. 'foobar_external_fid' or
// 'foobar_sometext_external_fid') we treat the field as a reference to another feature in the 'foobar' collection.
//
// Meaning data sources should be pre-populated with a 'foobar_external_fid' field containing UUIDs of other features.
// Creating these fields in the data source is beyond the scope of this application.
func isFeatureRelation(columnName string, externalFidColumn string) bool {
	if externalFidColumn == "" || columnName == externalFidColumn {
		return false
	}
	return strings.Contains(columnName, externalFidColumn)
}

func findReferencedCollection(collectionNames []string, name string) string {
	if collectionNames != nil {
		sort.Slice(collectionNames, func(i, j int) bool {
			return len(collectionNames[i]) > len(collectionNames[j])
		})

		// prefer exact matches first
		for _, collName := range collectionNames {
			if name == collName {
				return collName
			}
		}
		// then prefer fuzzy match (to support infix)
		for _, collName := range collectionNames {
			if strings.HasPrefix(name, collName) {
				return collName
			}
		}
	}
	log.Printf("Warning: could not find collection for feature reference '%s'", name)
	return ""
}
