package config

// +kubebuilder:object:generate=true
//
//nolint:recvcheck
type FeaturesCollection struct {
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

func (cf FeaturesCollection) GetID() string {
	return cf.ID
}

func (cf FeaturesCollection) GetMetadata() *GeoSpatialCollectionMetadata {
	return cf.Metadata
}

func (cf FeaturesCollection) GetLinks() *CollectionLinks {
	return cf.Links
}

func (cf FeaturesCollection) HasTableName(table string) bool {
	return cf.TableName != nil && table == *cf.TableName
}

func (cf FeaturesCollection) Merge(other GeoSpatialCollection) GeoSpatialCollection {
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

	// OAF Part 3: enhanced filtering capabilities expressed using "Common Query Language" (CQL2)
	// https://docs.ogc.org/is/19-079r2/19-079r2.html
	//
	// +optional
	CQL CQL `yaml:"cql,omitempty" json:"cql,omitempty"`
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
	PropertiesExcludeUnknown *bool `yaml:"propertiesExcludeUnknown,omitempty" json:"propertiesExcludeUnknown,omitempty" default:"false"`

	// When true properties are returned according to the ordering specified under 'properties'. When false
	// properties are returned in alphabetical order.
	// +optional
	// +kubebuilder:default=false
	PropertiesInSpecificOrder *bool `yaml:"propertiesInSpecificOrder,omitempty" json:"propertiesInSpecificOrder,omitempty" default:"false"`
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

// CQL Enable/disable CQL2 conformance classes (https://docs.ogc.org/is/21-065r2/21-065r2.html#cql2-enhancements)
// +kubebuilder:object:generate=true
type CQL struct {
	// Global setting to enable/disable CQL. When set to false, all other CQL settings are ignored.
	//
	// +kubebuilder:default=true
	// +optionals
	Enabled *bool `yaml:"enableCql,omitempty" json:"enableCql,omitempty" default:"true"`

	// Allow filtering using boolean operators (AND, OR, NOT) and simple comparison predicates (=, <>, <, >, <=, >=).
	//
	// This setting enables conformance class: https://docs.ogc.org/is/21-065r2/21-065r2.html#cql2-core
	//
	// +kubebuilder:default=true
	// +optional
	EnableBasicOperators *bool `yaml:"enableBasicOperators,omitempty" json:"enableBasicOperators,omitempty" default:"true"`

	// Allow filtering using advanced operators (LIKE, BETWEEN, IN, IS NULL).
	//
	// This setting enables conformance class: https://docs.ogc.org/is/21-065r2/21-065r2.html#advanced-comparison-operators
	//
	// +kubebuilder:default=true
	// +optional
	EnableAdvancedComparisonOperators bool `yaml:"enableAdvancedComparisonOperators,omitempty" json:"enableAdvancedComparisonOperators,omitempty" default:"true"`

	// Allow upper/lowercase insensitive filtering (CASEI).
	//
	// This setting enables conformance class: https://docs.ogc.org/is/21-065r2/21-065r2.html#case-insensitive-comparison
	//
	// +kubebuilder:default=true
	// +optional
	EnableCaseInsensitiveComparison bool `yaml:"enableCaseInsensitiveComparison,omitempty" json:"enableCaseInsensitiveComparison,omitempty" default:"true"`

	// Allow accent- / diacritics-insensitive filtering (ACCENTI).
	//
	// This setting enables conformance class: https://docs.ogc.org/is/21-065r2/21-065r2.html#accent-insensitive-comparison
	//
	// +kubebuilder:default=true
	// +optional
	EnableAccentInsensitiveComparison bool `yaml:"enableAccentInsensitiveComparison,omitempty" json:"enableAccentInsensitiveComparison,omitempty" default:"true"`

	// Allow filtering using spatial intersection (S_INTERSECTS) on two types of geometries: POINT and BBOX.
	//
	// This setting enables conformance class: https://docs.ogc.org/is/21-065r2/21-065r2.html#basic-spatial-functions
	//
	// +kubebuilder:default=true
	// +optional
	EnableBasicSpatialFunctions bool `yaml:"enableBasicSpatialFunctions,omitempty" json:"enableBasicSpatialFunctions,omitempty" default:"true"`

	// Allow filtering using spatial intersection (S_INTERSECTS) on all types of geometries: POINT, BBOX, POLYGON,
	// LINESTRING, MULTIPOINT, MULTILINESTRING, MULTIPOLYGON, GEOMETRYCOLLECTION.
	//
	// This setting enables conformance class: https://docs.ogc.org/is/21-065r2/21-065r2.html#basic-spatial-functions-plus
	//
	// +kubebuilder:default=true
	// +optional
	EnableBasicSpatialFunctionsPlus bool `yaml:"enableBasicSpatialFunctionsPlus,omitempty" json:"enableBasicSpatialFunctionsPlus,omitempty" default:"true"`

	// Allow filtering using all spatial operators (S_INTERSECTS, S_CONTAINS, S_WITHIN, S_OVERLAPS, S_EQUALS, S_DISJOINT) on all
	// types of geometries: POINT, BBOX, POLYGON, LINESTRING, MULTIPOINT, MULTILINESTRING, MULTIPOLYGON, GEOMETRYCOLLECTION.
	//
	// This setting enables conformance class: https://docs.ogc.org/is/21-065r2/21-065r2.html#spatial-functions
	//
	// +kubebuilder:default=true
	// +optional
	EnableSpatialFunctions bool `yaml:"enableSpatialFunctions,omitempty" json:"enableSpatialFunctions,omitempty" default:"true"`

	// Allow filtering using temporal operators (T_AFTER, T_BEFORE, T_DISJOINT, T_EQUALS, T_INTERSECTS, T_CONTAINS,
	// T_DURING, T_FINISHEDBY, T_FINISHES, T_MEETS, T_METBY, T_OVERLAPPEDBY, T_OVERLAPS, T_STARTEDBY, T_STARTS) on
	// instants and intervals.
	//
	// This setting enables conformance class: https://docs.ogc.org/is/21-065r2/21-065r2.html#temporal-functions
	//
	// +kubebuilder:default=true
	// +optional
	EnableTemporalFunctions bool `yaml:"enableTemporalFunctions,omitempty" json:"enableTemporalFunctions,omitempty" default:"true"`

	// Concerning remaining CQL2 conformance classes:
	// - Array functions are not supported, since we don't have arrays in the API/datasource
	// - Property-property is not supported (no need for currently)
	// - Custom functions are not supported (no need for currently)
	// - Arithmetic expressions are not supported (no need for currently)
}

// +kubebuilder:object:generate=true
type TemporalProperties struct {
	// Name of field in datasource to be used in temporal queries as the start date
	StartDate string `yaml:"startDate" json:"startDate" validate:"required"`

	// Name of field in datasource to be used in temporal queries as the end date
	EndDate string `yaml:"endDate" json:"endDate" validate:"required"`
}
