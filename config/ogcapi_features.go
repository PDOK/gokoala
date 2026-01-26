package config

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/PDOK/gokoala/internal/engine/util"
	"gopkg.in/yaml.v3"
)

// +kubebuilder:object:generate=true
type OgcAPIFeatures struct {
	// Basemap to use in embedded viewer on the HTML pages.
	// +kubebuilder:default="OSM"
	// +kubebuilder:validation:Enum=OSM;BRT
	// +optional
	Basemap string `yaml:"basemap,omitempty" json:"basemap,omitempty" default:"OSM" validate:"oneof=OSM BRT"`

	// Collections to be served as features through this API
	Collections CollectionsFeatures `yaml:"collections" json:"collections" validate:"required,dive"`

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

type CollectionsFeatures []CollectionFeatures

// ContainsID check if a given collection - by ID - exists.
func (csf CollectionsFeatures) ContainsID(id string) bool {
	for _, coll := range csf {
		if coll.ID == id {
			return true
		}
	}
	return false
}

// FeaturePropertiesByID returns a map of collection IDs to their corresponding FeatureProperties.
// Skips collections that do not have features defined.
func (csf CollectionsFeatures) FeaturePropertiesByID() map[string]*FeatureProperties {
	result := make(map[string]*FeatureProperties)
	for _, collection := range csf {
		result[collection.ID] = collection.FeatureProperties
	}

	return result
}

// +kubebuilder:object:generate=true
//
//nolint:recvcheck
type CollectionFeatures struct {
	// Unique ID of the collection
	// +kubebuilder:validation:Pattern=`^[a-z0-9"]([a-z0-9_-]*[a-z0-9"]+|)$`
	ID string `yaml:"id" validate:"required,lowercase_id" json:"id"`

	// Metadata describing the collection contents
	// +optional
	Metadata *GeoSpatialCollectionMetadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// Links pertaining to this collection (e.g., downloads, documentation)
	// +optional
	Links *CollectionLinks `yaml:"links,omitempty" json:"links,omitempty"`

	// Optional way to explicitly map a collection ID to the underlying table in the datasource.
	// +optional
	TableName *string `yaml:"tableName,omitempty" json:"tableName,omitempty"`

	// Optional collection-specific datasources. Mutually exclusive with top-level defined datasources.
	// +optional
	Datasources *Datasources `yaml:"datasources,omitempty" json:"datasources,omitempty"`

	// Filters available for this collection
	// +optional
	Filters FeatureFilters `yaml:"filters,omitempty" json:"filters,omitempty"`

	// Relations define relationships between features across collections
	// +optional
	Relations []Relation `yaml:"relations,omitempty" json:"relations,omitempty"`

	// Optional way to exclude feature properties and/or determine the ordering of properties in the response.
	// +optional
	*FeatureProperties `yaml:",inline" json:",inline"`

	// Downloads available for this collection through map sheets. Note that 'map sheets' refer to a map
	// divided in rectangle areas that can be downloaded individually.
	// +optional
	MapSheetDownloads *MapSheetDownloads `yaml:"mapSheetDownloads,omitempty" json:"mapSheetDownloads,omitempty"`

	// Configuration specifically related to HTML/Web representation
	// +optional
	Web *WebConfig `yaml:"web,omitempty" json:"web,omitempty"`
}

// MarshalJSON custom because inlining only works on embedded structs.
// Value instead of pointer receiver because only that way it can be used for both.
func (cf CollectionFeatures) MarshalJSON() ([]byte, error) {
	return json.Marshal(cf)
}

// UnmarshalJSON parses a string to CollectionFeatures.
func (cf CollectionFeatures) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, cf)
}

func (cf CollectionFeatures) GetID() string {
	return cf.ID
}

func (cf CollectionFeatures) GetMetadata() *GeoSpatialCollectionMetadata {
	return cf.Metadata
}

func (cf CollectionFeatures) GetLinks() *CollectionLinks {
	return cf.Links
}

func (cf CollectionFeatures) HasDateTime() bool {
	return cf.Metadata != nil && cf.Metadata.TemporalProperties != nil
}

func (cf CollectionFeatures) HasTableName(table string) bool {
	return cf.TableName != nil && table == *cf.TableName
}

func (cf CollectionFeatures) Merge(other GeoSpatialCollection) GeoSpatialCollection {
	cf.Metadata = mergeMetadata(cf, other)
	cf.Links = mergeLinks(cf, other)
	return cf
}

// +kubebuilder:object:generate=true
type FeatureFilters struct {
	// OAF Part 1: filter on feature properties
	// https://docs.ogc.org/is/17-069r4/17-069r4.html#_parameters_for_filtering_on_feature_properties
	//
	// +optional
	Properties []PropertyFilter `yaml:"properties,omitempty" json:"properties,omitempty" validate:"dive"`

	// OAF Part 3: add config for complex/CQL filters here
	// <placeholder>
}

// +kubebuilder:object:generate=true
type FeatureProperties struct {
	// Properties/fields of features in this collection. This setting controls two things:
	//
	// A) allows one to exclude certain properties, when propertiesExcludeUnknown=true
	// B) allows one to sort the properties in the given order, when propertiesInSpecificOrder=true
	//
	// When not set, all available properties are returned in API responses, in alphabetical order.
	// +optional
	Properties []string `yaml:"properties,omitempty" json:"properties,omitempty"`

	// When true properties not listed under 'properties' are excluded from API responses. When false
	// unlisted properties are also included in API responses.
	// +optional
	// +kubebuilder:default=false
	PropertiesExcludeUnknown bool `yaml:"propertiesExcludeUnknown,omitempty" json:"propertiesExcludeUnknown,omitempty" default:"false"`

	// When true properties are returned according to the ordering specified under 'properties'. When false
	// properties are returned in alphabetical order.
	// +optional
	// +kubebuilder:default=false
	PropertiesInSpecificOrder bool `yaml:"propertiesInSpecificOrder,omitempty" json:"propertiesInSpecificOrder,omitempty" default:"false"`
}

// +kubebuilder:object:generate=true
type MapSheetDownloads struct {
	// Properties that provide the download details per map sheet. Note that 'map sheets' refer to a map
	// divided in rectangle areas that can be downloaded individually.
	Properties MapSheetDownloadProperties `yaml:"properties" json:"properties" validate:"required"`
}

// +kubebuilder:object:generate=true
type MapSheetDownloadProperties struct {
	// Property/column containing file download URL
	AssetURL string `yaml:"assetUrl" json:"assetUrl" validate:"required"`

	// Property/column containing file size
	Size string `yaml:"size" json:"size" validate:"required"`

	// The actual media type (not a property/column) of the download, like application/zip.
	MediaType MediaType `yaml:"mediaType" json:"mediaType" validate:"required"`

	// Property/column containing the map sheet identifier
	MapSheetID string `yaml:"mapSheetId" json:"mapSheetId" validate:"required"`
}

// +kubebuilder:object:generate=true
type WebConfig struct {
	// Viewer config for displaying multiple features on a map
	// +optional
	FeaturesViewer *FeaturesViewer `yaml:"featuresViewer,omitempty" json:"featuresViewer,omitempty"`

	// Viewer config for displaying a single feature on a map
	// +optional
	FeatureViewer *FeaturesViewer `yaml:"featureViewer,omitempty" json:"featureViewer,omitempty"`

	// Whether URLs (to external resources) in the HTML representation of features should be rendered as hyperlinks.
	// +optional
	URLAsHyperlink bool `yaml:"urlAsHyperlink,omitempty" json:"urlAsHyperlink,omitempty"`
}

// +kubebuilder:object:generate=true
type FeaturesViewer struct {
	// Maximum initial zoom level of the viewer when rendering features, specified by scale denominator.
	// Defaults to 1000 (= scale 1:1000).
	// +optional
	MinScale int `yaml:"minScale,omitempty" json:"minScale,omitempty" validate:"gt=0" default:"1000"`

	// Minimal initial zoom level of the viewer when rendering features, specified by scale denominator
	// (not set by default).
	// +optional
	MaxScale *int `yaml:"maxScale,omitempty" json:"maxScale,omitempty" validate:"omitempty,gt=0,gtefield=MinScale"`
}

// +kubebuilder:object:generate=true
type Limit struct {
	// Number of features to return by default.
	// +kubebuilder:default=10
	// +kubebuilder:validation:Minimum=2
	// +optional
	Default int `yaml:"default,omitempty" json:"default,omitempty" validate:"gt=1" default:"10"`

	// Max number of features to return. Should be larger than 100 since the HTML interface always offers a 100 limit option.
	// +kubebuilder:default=1000
	// +kubebuilder:validation:Minimum=100
	// +optional
	Max int `yaml:"max,omitempty" json:"max,omitempty" validate:"gte=100" default:"1000"`
}

// +kubebuilder:object:generate=true
type PropertyFilter struct {
	// Needs to match with a column name in the feature table (in the configured datasource)
	Name string `yaml:"name" json:"name" validate:"required"`

	// Explains this property filter
	// +kubebuilder:default="Filter features by this property"
	// +optional
	Description string `yaml:"description,omitempty" json:"description,omitempty" default:"Filter features by this property"`

	// When true the property/column in the feature table needs to be indexed. Initialization will fail
	// when no index is present, when false the index check is skipped. For large tables an index is recommended!
	//
	// +kubebuilder:default=true
	// +optional
	IndexRequired *bool `yaml:"indexRequired,omitempty" json:"indexRequired,omitempty" default:"true"` // ptr due to https://github.com/creasty/defaults/issues/49

	// Static list of allowed values to be used as input for this property filter. Will be enforced by OpenAPI spec.
	// +optional
	AllowedValues []string `yaml:"allowedValues,omitempty" json:"allowedValues,omitempty"`

	// Derive a list of allowed values for this property filter from the corresponding column in the datastore.
	// Use with caution since it can increase startup time when used on large tables. Make sure an index in present.
	//
	// +kubebuilder:default=false
	// +optional
	DeriveAllowedValuesFromDatasource *bool `yaml:"deriveAllowedValuesFromDatasource,omitempty" json:"deriveAllowedValuesFromDatasource,omitempty" default:"false"`
}

// +kubebuilder:object:generate=true
type TemporalProperties struct {
	// Name of field in datasource to be used in temporal queries as the start date
	StartDate string `yaml:"startDate" json:"startDate" validate:"required"`

	// Name of field in datasource to be used in temporal queries as the end date
	EndDate string `yaml:"endDate" json:"endDate" validate:"required"`
}

func validateFeatureCollections(collections []CollectionFeatures) error {
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
