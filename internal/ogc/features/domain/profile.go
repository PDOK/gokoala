package domain

import (
	"fmt"
	"net/url"
	"regexp"
)

const regexRemoveSeparators = "[^a-z0-9]?"
const featurePath = "%s/collections/%s/items/%s"

type ProfileName string

// Profiles from OAF Part 5 as specified in https://docs.ogc.org/DRAFTS/23-058r1.html#rc_profile-parameter
const (
	RelAsKey  ProfileName = "rel-as-key"  // RelAsKey a feature reference in the response SHALL be represented by: The featureId
	RelAsURI  ProfileName = "rel-as-uri"  // RelAsURI a feature reference in the response SHALL be represented by: an HTTP(S) URI.
	RelAsLink ProfileName = "rel-as-link" // RelAsLink a feature reference in the response SHALL be represented by: an object with the property "href" and, optionally a "title"
)

// Profile from OAF Part 5, used to express relations between features
type Profile struct {
	profileName ProfileName
	baseURL     string
}

func NewProfile(profileName ProfileName, baseURL url.URL) Profile {
	return Profile{
		profileName: profileName,
		baseURL:     baseURL.String(),
	}
}

func (p *Profile) MapRelationUsingProfile(columnName string, columnValue any, externalFidColumn string) (newColumnName string, newColumnValue any) {
	regex, _ := regexp.Compile(regexRemoveSeparators + externalFidColumn + regexRemoveSeparators)
	switch p.profileName {
	case RelAsLink:
		collectionName := regex.ReplaceAllString(columnName, "")
		newColumnName = collectionName + ".href"
		if columnValue != nil {
			newColumnValue = fmt.Sprintf(featurePath, p.baseURL, collectionName, columnValue)
		}
	case RelAsKey:
		newColumnName = regex.ReplaceAllString(columnName, "")
		newColumnValue = columnValue
	case RelAsURI:
		// almost identical to rel-as-link except that there's no ".href" suffix
		newColumnName = regex.ReplaceAllString(columnName, "")
		if columnValue != nil {
			newColumnValue = fmt.Sprintf(featurePath, p.baseURL, newColumnName, columnValue)
		}
	}
	return
}
