package domain

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
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
	profileName     ProfileName
	baseURL         string
	collectionNames []string
}

func NewProfile(profileName ProfileName, baseURL url.URL, collectionNames []string) Profile {
	return Profile{
		profileName:     profileName,
		baseURL:         baseURL.String(),
		collectionNames: collectionNames,
	}
}

func (p *Profile) MapRelationUsingProfile(columnName string, columnValue any, externalFidColumn string) (newColumnName string, newColumnValue any) {
	regex, _ := regexp.Compile(regexRemoveSeparators + externalFidColumn + regexRemoveSeparators)
	switch p.profileName {
	case RelAsLink:
		newColumnName = regex.ReplaceAllString(columnName, "")
		collectionName := p.findCollectionWithPrefix(newColumnName)
		newColumnName += ".href"
		if columnValue != nil && collectionName != "" {
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

func (p *Profile) findCollectionWithPrefix(prefix string) string {
	for _, collName := range p.collectionNames {
		if strings.HasPrefix(prefix, collName) {
			return collName
		}
	}
	return ""
}
