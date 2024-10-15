package config

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/docker/go-units"
)

// +kubebuilder:object:generate=true
type OgcAPIFeatures struct {
	// Basemap to use in embedded viewer on the HTML pages.
	// +kubebuilder:default="OSM"
	// +kubebuilder:validation:Enum=OSM;BRT
	// +optional
	Basemap string `yaml:"basemap,omitempty" json:"basemap,omitempty" default:"OSM" validate:"oneof=OSM BRT"`

	// Collections to be served as features through this API
	Collections GeoSpatialCollections `yaml:"collections" json:"collections" validate:"required,dive"`

	// Limits the amount of features to retrieve with a single call
	// +optional
	Limit Limit `yaml:"limit,omitempty" json:"limit,omitempty"`

	// One or more datasources to get the features from (geopackages, postgis, etc).
	// Optional since you can also define datasources at the collection level
	// +optional
	Datasources *Datasources `yaml:"datasources,omitempty" json:"datasources,omitempty"`

	// Whether GeoJSON/JSON-FG responses will be validated against the OpenAPI spec
	// since it has significant performance impact when dealing with large JSON payloads.
	//
	// +kubebuilder:default=true
	// +optional
	ValidateResponses *bool `yaml:"validateResponses,omitempty" json:"validateResponses,omitempty" default:"true"` // ptr due to https://github.com/creasty/defaults/issues/49
}

func (oaf *OgcAPIFeatures) ProjectionsForCollections() []string {
	return oaf.ProjectionsForCollection("")
}

func (oaf *OgcAPIFeatures) ProjectionsForCollection(collectionID string) []string {
	uniqueSRSs := make(map[string]struct{})
	if oaf.Datasources != nil {
		for _, a := range oaf.Datasources.Additional {
			uniqueSRSs[a.Srs] = struct{}{}
		}
	}
	for _, coll := range oaf.Collections {
		if (coll.ID == collectionID || collectionID == "") && coll.Features != nil && coll.Features.Datasources != nil {
			for _, a := range coll.Features.Datasources.Additional {
				uniqueSRSs[a.Srs] = struct{}{}
			}
			break
		}
	}
	result := util.Keys(uniqueSRSs)
	slices.Sort(result)
	return result
}

// +kubebuilder:object:generate=true
type CollectionEntryFeatures struct {
	// Optional way to explicitly map a collection ID to the underlying table in the datasource.
	// +optional
	TableName *string `yaml:"tableName,omitempty" json:"tableName,omitempty"`

	// Optional collection specific datasources. Mutually exclusive with top-level defined datasources.
	// +optional
	Datasources *Datasources `yaml:"datasources,omitempty" json:"datasources,omitempty"`

	// Filters available for this collection
	// +optional
	Filters FeatureFilters `yaml:"filters,omitempty" json:"filters,omitempty"`

	// Optional way to exclude feature properties and/or determine the ordering of properties in the response.
	// +optional
	FeatureProperties *FeatureProperties `yaml:",inline" json:",inline"`

	// Downloads available for this collection through map sheets. Note that 'map sheets' refer to a map
	// divided in rectangle areas that can be downloaded individually.
	// +optional
	MapSheetDownloads *MapSheetDownloads `yaml:"mapSheetDownloads,omitempty" json:"mapSheetDownloads,omitempty"`

	// Configuration specifically related to HTML/Web representation
	// +optional
	Web *WebConfig `yaml:"web,omitempty" json:"web,omitempty"`
}

// +kubebuilder:object:generate=true
type Datasources struct {
	// Features should always be available in WGS84 (according to spec).
	// This specifies the datasource to be used for features in the WGS84 projection
	DefaultWGS84 Datasource `yaml:"defaultWGS84" json:"defaultWGS84" validate:"required"`

	// One or more additional datasources for features in other projections. GoKoala doesn't do
	// any on-the-fly reprojection so additional datasources need to be reprojected ahead of time.
	// +optional
	Additional []AdditionalDatasource `yaml:"additional" json:"additional" validate:"dive"`
}

// +kubebuilder:object:generate=true
type Datasource struct {
	// GeoPackage to get the features from.
	// +optional
	GeoPackage *GeoPackage `yaml:"geopackage,omitempty" json:"geopackage,omitempty" validate:"required_without_all=PostGIS"`

	// PostGIS database to get the features from (not implemented yet).
	// +optional
	PostGIS *PostGIS `yaml:"postgis,omitempty" json:"postgis,omitempty" validate:"required_without_all=GeoPackage"`

	// Add more datasources here such as Mongo, Elastic, etc
}

// +kubebuilder:object:generate=true
type AdditionalDatasource struct {
	// Projection (SRS/CRS) used for the features in this datasource
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	Srs string `yaml:"srs" json:"srs" validate:"required,startswith=EPSG:"`

	// The additional datasource
	Datasource `yaml:",inline" json:",inline"`
}

// +kubebuilder:object:generate=true
type PostGIS struct {
	// placeholder
}

// +kubebuilder:object:generate=true
type GeoPackage struct {
	// Settings to read a GeoPackage from local disk
	// +optional
	Local *GeoPackageLocal `yaml:"local,omitempty" json:"local,omitempty" validate:"required_without_all=Cloud"`

	// Settings to read a GeoPackage as a Cloud-Backed SQLite database
	// +optional
	Cloud *GeoPackageCloud `yaml:"cloud,omitempty" json:"cloud,omitempty" validate:"required_without_all=Local"`
}

// +kubebuilder:object:generate=true
type GeoPackageCommon struct {
	// Feature id column name
	// +kubebuilder:default="fid"
	// +optional
	Fid string `yaml:"fid,omitempty" json:"fid,omitempty" validate:"required" default:"fid"`

	// External feature id column name. When specified this ID column will be exposed to clients instead of the regular FID column.
	// It allows one to offer a more stable ID to clients instead of an auto-generated FID. External FID column should contain UUIDs.
	// +optional
	ExternalFid string `yaml:"externalFid" json:"externalFid"`

	// Optional timeout after which queries are canceled
	// +kubebuilder:default="15s"
	// +optional
	QueryTimeout Duration `yaml:"queryTimeout,omitempty" json:"queryTimeout,omitempty" validate:"required" default:"15s"`

	// ADVANCED SETTING. When the number of features in a bbox stay within the given value use an RTree index, otherwise use a BTree index.
	// +kubebuilder:default=8000
	// +optional
	MaxBBoxSizeToUseWithRTree int `yaml:"maxBBoxSizeToUseWithRTree,omitempty" json:"maxBBoxSizeToUseWithRTree,omitempty" validate:"required" default:"8000"`

	// ADVANCED SETTING. Sets the SQLite "cache_size" pragma which determines how many pages are cached in-memory.
	// See https://sqlite.org/pragma.html#pragma_cache_size for details.
	// Default in SQLite is 2000 pages, which equates to 2000KiB (2048000 bytes). Which is denoted as -2000.
	// +kubebuilder:default=-2000
	// +optional
	InMemoryCacheSize int `yaml:"inMemoryCacheSize,omitempty" json:"inMemoryCacheSize,omitempty" validate:"required" default:"-2000"`
}

// +kubebuilder:object:generate=true
type GeoPackageLocal struct {
	// GeoPackageCommon shared config between local and cloud GeoPackage
	GeoPackageCommon `yaml:",inline" json:",inline"`

	// Location of GeoPackage on disk.
	// You can place the GeoPackage here manually (out-of-band) or you can specify Download
	// and let the application download the GeoPackage for you and store it at this location.
	File string `yaml:"file" json:"file" validate:"required,omitempty,filepath"`

	// Optional initialization task to download a GeoPackage during startup. GeoPackage will be
	// downloaded to local disk and stored at the location specified in File.
	// +optional
	Download *GeoPackageDownload `yaml:"download,omitempty" json:"download,omitempty"`
}

// +kubebuilder:object:generate=true
type GeoPackageDownload struct {
	// Location of GeoPackage on remote HTTP(S) URL. GeoPackage will be downloaded to local disk
	// during startup and stored at the location specified in "file".
	From URL `yaml:"from" json:"from" validate:"required"`

	// ADVANCED SETTING. Determines how many workers (goroutines) in parallel will download the specified GeoPackage.
	// Setting this to 1 will disable concurrent downloads.
	// +kubebuilder:default=4
	// +kubebuilder:validation:Minimum=1
	// +optional
	Parallelism int `yaml:"parallelism,omitempty" json:"parallelism,omitempty" validate:"required,gte=1" default:"4"`

	// ADVANCED SETTING. When true TLS certs are NOT validated, false otherwise. Only use true for your own self-signed certificates!
	// +kubebuilder:default=false
	// +optional
	TLSSkipVerify bool `yaml:"tlsSkipVerify,omitempty" json:"tlsSkipVerify,omitempty" default:"false"`

	// ADVANCED SETTING. HTTP request timeout when downloading (part of) GeoPackage.
	// +kubebuilder:default="2m"
	// +optional
	Timeout Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required" default:"2m"`

	// ADVANCED SETTING. Minimum delay to use when retrying HTTP request to download (part of) GeoPackage.
	// +kubebuilder:default="1s"
	// +optional
	RetryDelay Duration `yaml:"retryDelay,omitempty" json:"retryDelay,omitempty" validate:"required" default:"1s"`

	// ADVANCED SETTING. Maximum overall delay of the exponential backoff while retrying HTTP requests to download (part of) GeoPackage.
	// +kubebuilder:default="30s"
	// +optional
	RetryMaxDelay Duration `yaml:"retryMaxDelay,omitempty" json:"retryMaxDelay,omitempty" validate:"required" default:"30s"`

	// ADVANCED SETTING. Maximum number of retries when retrying HTTP requests to download (part of) GeoPackage.
	// +kubebuilder:default=5
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxRetries int `yaml:"maxRetries,omitempty" json:"maxRetries,omitempty" validate:"required,gte=1" default:"5"`
}

// +kubebuilder:object:generate=true
type GeoPackageCloud struct {
	// GeoPackageCommon shared config between local and cloud GeoPackage
	GeoPackageCommon `yaml:",inline" json:",inline"`

	// Reference to the cloud storage (either azure or google at the moment).
	// For example 'azure?emulator=127.0.0.1:10000&sas=0' or 'google'
	Connection string `yaml:"connection" json:"connection" validate:"required"`

	// Username of the storage account, like devstoreaccount1 when using Azurite
	User string `yaml:"user" json:"user" validate:"required"`

	// Some kind of credential like a password or key to authenticate with the storage backend, e.g:
	// 'Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==' when using Azurite
	Auth string `yaml:"auth" json:"auth" validate:"required"`

	// Container/bucket on the storage account
	Container string `yaml:"container" json:"container" validate:"required"`

	// Filename of the GeoPackage
	File string `yaml:"file" json:"file" validate:"required"`

	// Local cache of fetched blocks from cloud storage
	// +optional
	Cache GeoPackageCloudCache `yaml:"cache,omitempty" json:"cache,omitempty"`

	// ADVANCED SETTING. Only for debug purposes! When true all HTTP requests executed by sqlite to cloud object storage are logged to stdout
	// +kubebuilder:default=false
	// +optional
	LogHTTPRequests bool `yaml:"logHttpRequests,omitempty" json:"logHttpRequests,omitempty" default:"false"`
}

func (gc *GeoPackageCloud) CacheDir() (string, error) {
	fileNameWithoutExt := strings.TrimSuffix(gc.File, filepath.Ext(gc.File))
	if gc.Cache.Path != nil {
		randomSuffix := strconv.Itoa(rand.Intn(99999)) //nolint:gosec // random isn't used for security purposes
		return filepath.Join(*gc.Cache.Path, fileNameWithoutExt+"-"+randomSuffix), nil
	}
	cacheDir, err := os.MkdirTemp("", fileNameWithoutExt)
	if err != nil {
		return "", fmt.Errorf("failed to create tempdir to cache %s, error %w", fileNameWithoutExt, err)
	}
	return cacheDir, nil
}

// +kubebuilder:object:generate=true
type GeoPackageCloudCache struct {
	// Optional path to directory for caching cloud-backed GeoPackage blocks, when omitted a temp dir will be used.
	// +optional
	Path *string `yaml:"path,omitempty" json:"path,omitempty" validate:"omitempty,dirpath|filepath"`

	// Max size of the local cache. Accepts human-readable size such as 100Mb, 4Gb, 1Tb, etc. When omitted 1Gb is used.
	// +kubebuilder:default="1Gb"
	// +optional
	MaxSize string `yaml:"maxSize,omitempty" json:"maxSize,omitempty" default:"1Gb"`

	// When true a warm-up query is executed on startup which aims to fill the local cache. Does increase startup time.
	// +kubebuilder:default=false
	// +optional
	WarmUp bool `yaml:"warmUp,omitempty" json:"warmUp,omitempty" default:"false"`
}

func (cache *GeoPackageCloudCache) MaxSizeAsBytes() (int64, error) {
	return units.FromHumanSize(cache.MaxSize)
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
	// B) allows one sort the properties in the given order, when propertiesInSpecificOrder=true
	//
	// When not set all available properties are returned in API responses, in alphabetical order.
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

	// Derive list of allowed values for this property filter from the corresponding column in the datastore.
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

func validateFeatureCollections(collections GeoSpatialCollections) error {
	var errMessages []string
	for _, collection := range collections {
		if collection.Metadata != nil && collection.Metadata.TemporalProperties != nil &&
			(collection.Metadata.Extent == nil || collection.Metadata.Extent.Interval == nil) {
			errMessages = append(errMessages, fmt.Sprintf("validation failed for collection '%s'; "+
				"field 'Extent.Interval' is required with field 'TemporalProperties'\n", collection.ID))
		}
		if collection.Features != nil && collection.Features.Filters.Properties != nil {
			for _, pf := range collection.Features.Filters.Properties {
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
