package domain

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

const featurePath = "%s/collections/%s/items/%s"

type ProfileName string

// Profiles from OAF Part 5 as specified in https://docs.ogc.org/DRAFTS/23-058r1.html#rc_profile-parameter
const (
	RelAsKey  ProfileName = "rel-as-key"  // RelAsKey a feature reference in the response SHALL be represented by: The featureId
	RelAsURI  ProfileName = "rel-as-uri"  // RelAsURI a feature reference in the response SHALL be represented by: an HTTP(S) URI.
	RelAsLink ProfileName = "rel-as-link" // RelAsLink a feature reference in the response SHALL be represented by: an object with the property "href" and, optionally a "title"
)

var SupportedProfiles = []ProfileName{
	RelAsKey, RelAsURI, RelAsLink,
}

// Profile from OAF Part 5, used to express relations between features.
type Profile struct {
	profileName ProfileName
	baseURL     string
	schema      Schema
}

func NewProfile(profileName ProfileName, baseURL url.URL, schema Schema) Profile {
	return Profile{
		profileName: profileName,
		baseURL:     baseURL.String(),
		schema:      schema,
	}
}

func (p *Profile) MapRelationUsingProfile(columnName string, columnValue any, externalFidColumn string) (string, string, any) {
	relationName := newFeatureRelationName(columnName, externalFidColumn)
	var newColumnName string
	var newColumnValue any

	switch p.profileName {
	case RelAsLink:
		newColumnName = relationName + ".href"
		newColumnValue = p.mapRelationValue(columnValue, true, relationName)
	case RelAsKey:
		newColumnName = relationName
		newColumnValue = p.mapRelationValue(columnValue, false, relationName)
	case RelAsURI:
		// almost identical to rel-as-link except that there's no ".href" suffix (and potentially a title in the future)
		newColumnName = relationName
		newColumnValue = p.mapRelationValue(columnValue, true, relationName)
	}

	return newColumnName, relationName, newColumnValue
}

func (p *Profile) mapRelationValue(columnValue any, formatAsURL bool, relationName string) any {
	if columnValue == nil {
		return nil
	}
	featureRelation := p.schema.findFeatureRelation(relationName)
	if featureRelation == nil {
		log.Printf("Warning: relation %s not found in schema", relationName)
		return nil
	}

	values := strings.Split(fmt.Sprintf("%v", columnValue), ",")
	result := make([]string, 0, len(values))
	for _, v := range values {
		if formatAsURL {
			v = fmt.Sprintf(featurePath, p.baseURL, featureRelation.CollectionID, v)
		}
		result = append(result, v)
	}

	if len(result) == 1 && !featureRelation.IsArray {
		return result[0]
	}
	return result
}
