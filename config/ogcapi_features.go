package config

import (
	"fmt"
	"slices"

	"github.com/PDOK/gokoala/internal/engine/util"
)

// +kubebuilder:object:generate=true
type OgcAPIFeatures struct {
	// Basemap to use in embedded viewer on the HTML pages.
	// +kubebuilder:default="OSM"
	// +kubebuilder:validation:Enum=OSM;BRT
	// +optional
	Basemap string `yaml:"basemap,omitempty" json:"basemap,omitempty" default:"OSM" validate:"oneof=OSM BRT"`

	// Collections to be served as features through this API
	Collections FeaturesCollections `yaml:"collections" json:"collections" validate:"required,dive"`

	// Limits the number of features to retrieve with a single call
	// +optional
	Limit Limit `yaml:"limit,omitempty" json:"limit,omitempty"`

	// One or more datasources to get the features from (geopackages, postgres, etc).
	// Optional since you can also define datasources at the collection level
	// +optional
	Datasources *Datasources `yaml:"datasources,omitempty" json:"datasources,omitempty"`

	// Whether GeoJSON/JSON-FG responses will be validated against the OpenAPI spec
	// since it has a significant performance impact when dealing with large JSON payloads.
	//
	// +kubebuilder:default=true
	// +optional
	ValidateResponses *bool `yaml:"validateResponses,omitempty" json:"validateResponses,omitempty" default:"true"` // ptr due to https://github.com/creasty/defaults/issues/49

	// Maximum number of decimals allowed in geometry coordinates. When not specified (default value of 0) no limit is enforced.
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxDecimals int `yaml:"maxDecimals,omitempty" json:"maxDecimals,omitempty" default:"0"`

	// Force timestamps in features to the UTC timezone.
	//
	// +kubebuilder:default=false
	// +optional
	ForceUTC bool `yaml:"forceUtc,omitempty" json:"forceUtc,omitempty"`

	// SupportsNonGeoData, when set to true, enables the API to advertise and handle collections
	// that do not contain geometric data (i.e., non-geo collections). This is useful for APIs
	// that need to serve tabular or attribute-only data alongside traditional geospatial collections.
	// When enabled, the geometryType for such collections will be advertised as "none".
	//
	// +kubebuilder:default=false
	// +optional
	SupportsNonGeoData bool `yaml:"supportsNonGeoData,omitempty" json:"supportsNonGeoData,omitempty"`
}

func (oaf *OgcAPIFeatures) CollectionsSRS() []string {
	return oaf.CollectionSRS("")
}

func (oaf *OgcAPIFeatures) CollectionSRS(collectionID string) []string {
	uniqueSRSs := make(map[string]struct{})
	if oaf.Datasources != nil {
		for _, d := range oaf.Datasources.OnTheFly {
			for _, srs := range d.SupportedSrs {
				uniqueSRSs[srs.Srs] = struct{}{}
			}
		}
		for _, d := range oaf.Datasources.Additional {
			uniqueSRSs[d.Srs] = struct{}{}
		}
	}
	for _, coll := range oaf.Collections {
		if (coll.ID == collectionID || collectionID == "") && coll.Datasources != nil {
			for _, d := range coll.Datasources.OnTheFly {
				for _, srs := range d.SupportedSrs {
					uniqueSRSs[srs.Srs] = struct{}{}
				}
			}
			for _, d := range coll.Datasources.Additional {
				uniqueSRSs[d.Srs] = struct{}{}
			}

			break
		}
	}
	result := util.Keys(uniqueSRSs)
	slices.Sort(result)

	return result
}

// SupportsPart3 true when OGC API supports Part 3 is supported, this depends on whether any of the collections supports CQL.
func (oaf *OgcAPIFeatures) SupportsPart3() bool {
	for _, coll := range oaf.Collections {
		if coll.Filters.CQL.Enabled != nil && *coll.Filters.CQL.Enabled {
			return true
		}
	}
	return false
}

// SupportsBasicCQL2 true when basic CQL2 (boolean operators and simple comparisons) is enabled for at least one collection.
func (oaf *OgcAPIFeatures) SupportsBasicCQL2() bool {
	for _, coll := range oaf.Collections {
		cql := coll.Filters.CQL
		if cql.Enabled != nil && *cql.Enabled && cql.EnableBasicOperators != nil && *cql.EnableBasicOperators {
			return true
		}
	}
	return false
}

// SupportsAdvancedComparisonOperators true when advanced comparison operators are enabled for at least one collection.
func (oaf *OgcAPIFeatures) SupportsAdvancedComparisonOperators() bool {
	for _, coll := range oaf.Collections {
		cql := coll.Filters.CQL
		if cql.Enabled != nil && *cql.Enabled && cql.EnableAdvancedComparisonOperators {
			return true
		}
	}
	return false
}

// SupportsCaseInsensitiveComparison true when case-insensitive comparison is enabled for at least one collection.
func (oaf *OgcAPIFeatures) SupportsCaseInsensitiveComparison() bool {
	for _, coll := range oaf.Collections {
		cql := coll.Filters.CQL
		if cql.Enabled != nil && *cql.Enabled && cql.EnableCaseInsensitiveComparison {
			return true
		}
	}
	return false
}

// SupportsAccentInsensitiveComparison true when accent/diacritics-insensitive comparison is enabled for at least one collection.
func (oaf *OgcAPIFeatures) SupportsAccentInsensitiveComparison() bool {
	for _, coll := range oaf.Collections {
		cql := coll.Filters.CQL
		if cql.Enabled != nil && *cql.Enabled && cql.EnableAccentInsensitiveComparison {
			return true
		}
	}
	return false
}

// SupportsBasicSpatialFunctions true when basic spatial functions are enabled for at least one collection.
func (oaf *OgcAPIFeatures) SupportsBasicSpatialFunctions() bool {
	if oaf.SupportsSpatialFunctions() || oaf.SupportsBasicSpatialFunctionsPlus() {
		return true
	}
	for _, coll := range oaf.Collections {
		cql := coll.Filters.CQL
		if cql.Enabled != nil && *cql.Enabled && cql.EnableBasicSpatialFunctions {
			return true
		}
	}
	return false
}

// SupportsBasicSpatialFunctionsPlus true when basic spatial functions plus (S_INTERSECTS on all geometry types) are enabled for at least one collection.
func (oaf *OgcAPIFeatures) SupportsBasicSpatialFunctionsPlus() bool {
	if oaf.SupportsSpatialFunctions() {
		return true
	}
	for _, coll := range oaf.Collections {
		cql := coll.Filters.CQL
		if cql.Enabled != nil && *cql.Enabled && cql.EnableBasicSpatialFunctionsPlus {
			return true
		}
	}
	return false
}

// SupportsSpatialFunctions true when ALL spatial operators are enabled for at least one collection.
func (oaf *OgcAPIFeatures) SupportsSpatialFunctions() bool {
	for _, coll := range oaf.Collections {
		cql := coll.Filters.CQL
		if cql.Enabled != nil && *cql.Enabled && cql.EnableSpatialFunctions {
			return true
		}
	}
	return false
}

// SupportsTemporalFunctions true when temporal operators are enabled for at least one collection.
func (oaf *OgcAPIFeatures) SupportsTemporalFunctions() bool {
	for _, coll := range oaf.Collections {
		cql := coll.Filters.CQL
		if cql.Enabled != nil && *cql.Enabled && cql.EnableTemporalFunctions {
			return true
		}
	}
	return false
}

type FeaturesCollections []FeaturesCollection

// ContainsID check if a given collection - by ID - exists.
func (csf FeaturesCollections) ContainsID(id string) bool {
	for _, coll := range csf {
		if coll.ID == id {
			return true
		}
	}
	return false
}

// FeaturePropertiesByID returns a map of collection IDs to their corresponding FeatureProperties.
// Skips collections that do not have features defined.
func (csf FeaturesCollections) FeaturePropertiesByID() map[string]*FeatureProperties {
	result := make(map[string]*FeatureProperties)
	for _, collection := range csf {
		result[collection.ID] = collection.FeatureProperties
	}

	return result
}

func validateFeatureCollections(collections []FeaturesCollection) error {
	var errMessages []string
	for _, collection := range collections {
		if collection.Metadata != nil && collection.Metadata.TemporalProperties != nil &&
			(collection.Metadata.Extent == nil || collection.Metadata.Extent.Interval == nil) {
			errMessages = append(errMessages, fmt.Sprintf("validation failed for collection '%s'; "+
				"field 'Extent.Interval' is required with field 'TemporalProperties'\n", collection.ID))
		}
		if collection.Filters.Properties != nil {
			for _, pf := range collection.Filters.Properties {
				if pf.AllowedValues != nil && *pf.DeriveAllowedValuesFromDatasource {
					errMessages = append(errMessages, fmt.Sprintf("validation failed for property filter '%s'; "+
						"field 'AllowedValues' and field 'DeriveAllowedValuesFromDatasource' are mutually exclusive\n", pf.Name))
				}
			}
		}
	}
	if len(errMessages) > 0 {
		return fmt.Errorf("invalid config provided:\n%v", errMessages)
	}

	return nil
}
