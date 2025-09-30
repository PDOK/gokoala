package domain

import (
	"fmt"
	"net/url"
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

func (p *Profile) MapRelationUsingProfile(columnName string, columnValue any, externalFidColumn string) (newColumnName, relationName string, newColumnValue any) {
	switch p.profileName {
	case RelAsLink:
		relationName = newFeatureRelationName(columnName, externalFidColumn)
		featureRelation := p.schema.findFeatureRelation(relationName)
		newColumnName = relationName + ".href"
		if columnValue != nil && featureRelation != nil {
			newColumnValue = fmt.Sprintf(featurePath, p.baseURL, featureRelation.CollectionID, columnValue)
		}
	case RelAsKey:
		relationName = newFeatureRelationName(columnName, externalFidColumn)
		newColumnName = relationName
		newColumnValue = columnValue
	case RelAsURI:
		// almost identical to rel-as-link except that there's no ".href" suffix (and potentially a title in the future)
		relationName = newFeatureRelationName(columnName, externalFidColumn)
		featureRelation := p.schema.findFeatureRelation(relationName)
		newColumnName = relationName
		if columnValue != nil && featureRelation != nil {
			newColumnValue = fmt.Sprintf(featurePath, p.baseURL, featureRelation.CollectionID, columnValue)
		}
	}

	return
}
