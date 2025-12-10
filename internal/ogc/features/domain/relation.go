package domain

import (
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/PDOK/gokoala/config"
)

const regexRemoveSeparators = "[^a-z0-9]?"

// FeatureRelation a relation/reference from one feature to other(s) in a different
// collection, according to OAF Part 5: https://docs.ogc.org/DRAFTS/23-058r1.html#rc_feature-references.
type FeatureRelation struct {
	// Name of the relation, e.g. 'foobar' (in 'foobar_external_fid') as displayed in the JSON/HTML output.
	Name string

	// CollectionID of the collection to which the relation points.
	CollectionID string

	// The true relation can point to multiple features, false when it points to a single feature.
	IsArray bool
}

func NewFeatureRelation(table string, name, externalFidColumn string, collections config.GeoSpatialCollections) *FeatureRelation {
	// option 1: deal with relations configured in the config file
	for _, collection := range collections {
		if !collection.HasTableName(table) || collection.Features.Relations == nil {
			continue
		}
		for _, relation := range collection.Features.Relations {
			if relation.Columns.Source == name {
				return &FeatureRelation{
					Name:         relation.Name(),
					CollectionID: relation.RelatedCollection,
					IsArray:      relation.Junction.Name != "",
				}
			}
		}
	}

	// option 2: deal with relations configured in the datasource
	collectionNames := make([]string, 0, len(collections))
	for _, collection := range collections {
		collectionNames = append(collectionNames, collection.ID)
	}
	if !IsFeatureRelation(name, externalFidColumn) {
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
// features in the 'foobar' collection. The field containing this relation will be named 'foobar' or 'foobar_sometext'.
// This name will appear in the feature data (GeoJSON) and the schema (JSON-Schema) to represent the feature relation.
func newFeatureRelationName(name string, externalFidColumn string) string {
	regex, _ := regexp.Compile(regexRemoveSeparators + externalFidColumn + regexRemoveSeparators)
	return regex.ReplaceAllString(name, "")
}

// IsFeatureRelation "Algorithm" to determine feature reference:
//
// When externalFidColumn (e.g. 'external_fid') is part of the column name (e.g. 'foobar_external_fid' or
// 'foobar_sometext_external_fid') we treat the field as a reference to another feature in the 'foobar' collection.
//
// Meaning data sources should be pre-populated with a 'foobar_external_fid' field containing UUIDs of other features.
// Creating these fields in the data source is beyond the scope of this application.
func IsFeatureRelation(columnName string, externalFidColumn string) bool {
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
