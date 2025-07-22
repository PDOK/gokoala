package config

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/go-units"
)

// +kubebuilder:object:generate=true
type Datasources struct {
	// Features should always be available in WGS84 (according to spec). This specifies the
	// datasource to be used for features in the WGS84 coordinate reference system.
	//
	// No on-the-fly transformation/reprojection is performed, so the features in this datasource need to be
	// either native WGS84 or reprojected/transformed to WGS84 ahead of time. For example, using ogr2ogr.
	// +optional
	DefaultWGS84 *Datasource `yaml:"defaultWGS84" json:"defaultWGS84"` //nolint:tagliatelle // grandfathered

	// One or more additional datasources for features in other (non-WGS84) coordinate reference systems.
	//
	// No on-the-fly transformation/reprojection is performed, so the features in these additional datasources
	// need to be transformed/reprojected ahead of time. For example, using ogr2ogr.
	// +optional
	Additional []AdditionalDatasource `yaml:"additional" json:"additional" validate:"dive"`

	// Datasource containing features which will be transformed/reprojected on-the-fly to the specified
	// coordinate reference systems. No need to transform/reproject ahead of time.
	//
	// Note: On-the-fly transformation/reprojection may impact performance when using (very) large geometries.
	// +optional
	OnTheFly []OnTheFlyDatasource `yaml:"transformOnTheFly" json:"transformOnTheFly" validate:"dive"`
}

// +kubebuilder:object:generate=true
type Datasource struct {
	// GeoPackage to get the features from.
	// +optional
	GeoPackage *GeoPackage `yaml:"geopackage,omitempty" json:"geopackage,omitempty" validate:"required_without_all=Postgres"`

	// Postgres database to get the features from.
	// +optional
	Postgres *Postgres `yaml:"postgres,omitempty" json:"postgres,omitempty" validate:"required_without_all=GeoPackage"`

	// Add more data sources here such as Mongo, Elastic, etc.
}

// +kubebuilder:object:generate=true
type AdditionalDatasource struct {
	// SRS/CRS used for the features in this datasource
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	Srs string `yaml:"srs" json:"srs" validate:"required,startswith=EPSG:"`

	// The additional datasource
	Datasource `yaml:",inline" json:",inline"`
}

// +kubebuilder:object:generate=true
type OnTheFlyDatasource struct {
	// List of supported SRS/CRS
	SupportedSrs []OnTheFlySupportedSrs `yaml:"supportedSrs,omitempty" json:"supportedSrs,omitempty" validate:"dive,omitempty"`

	// The datasource capable of on-the-fly reprojection/transformation
	Datasource `yaml:",inline" json:",inline"`
}

// +kubebuilder:object:generate=true
type OnTheFlySupportedSrs struct {
	// Supported coordinated reference systems (CRS/SRS) for on-the-fly reprojection/transformation.
	// Note: no need to add 'OGC:CRS84', since that one is required and included by default.
	// +kubebuilder:validation:Pattern=`^EPSG:\d+$`
	Srs string `yaml:"srs" json:"srs" validate:"required,startswith=EPSG:"`
}

// +kubebuilder:object:generate=true
type DatasourceCommon struct {
	// Feature id column name
	// +kubebuilder:default="fid"
	// +optional
	Fid string `yaml:"fid,omitempty" json:"fid,omitempty" validate:"required" default:"fid"`

	// External feature id column name. When specified, this ID column will be exposed to clients instead of the regular FID column.
	// It allows one to offer a more stable ID to clients instead of an auto-generated FID. External FID column should contain UUIDs.
	// +optional
	ExternalFid string `yaml:"externalFid" json:"externalFid"`

	// Optional timeout after which queries are canceled
	// +kubebuilder:default="15s"
	// +optional
	QueryTimeout Duration `yaml:"queryTimeout,omitempty" json:"queryTimeout,omitempty" validate:"required" default:"15s"`
}

// +kubebuilder:object:generate=true
type Postgres struct {
	DatasourceCommon `yaml:",inline" json:",inline"`

	// Hostname of the PostgreSQL server.
	// +kubebuilder:default="localhost"
	Host string `yaml:"host" json:"host" validate:"required,hostname_rfc1123" default:"localhost"`

	// Port number of the PostgreSQL server.
	// +kubebuilder:default="5432"
	Port uint `yaml:"port" json:"port" validate:"required,port" default:"5432"`

	// Name of the PostgreSQL database containing the data.
	// +kubebuilder:default="postgres"
	DatabaseName string `yaml:"databaseName" json:"databaseName" validate:"required" default:"postgres"`

	// Name of the PostgreSQL schema containing the data.
	// +kubebuilder:default="public"
	Schema string `yaml:"schema" json:"schema" validate:"required" default:"public"`

	// The SSL mode to use, e.g. 'disable', 'allow', 'prefer', 'require', 'verify-ca' or 'verify-full'.
	// +kubebuilder:validation:Enum=disable;allow;prefer;require;verify-ca;verify-full
	// +kubebuilder:default="disable"
	SSLMode string `yaml:"sslMode" json:"sslMode" validate:"required" default:"disable"`

	// Username when connecting to the PostgreSQL server.
	// +kubebuilder:default="postgres"
	User string `yaml:"user" json:"user" validate:"required" default:"postgres"`

	// Password when connecting to the PostgreSQL server.
	// +kubebuilder:default="postgres"
	Pass string `yaml:"pass" json:"pass" validate:"required" default:"postgres"`
}

func (p *Postgres) ConnectionString() string {
	port := strconv.FormatUint(uint64(p.Port), 10)
	defaultSearchPath := "public, postgis, topology" // otherwise postgis extension isn't found

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s&search_path=%s,%s&application_name=%s",
		p.User, p.Pass, net.JoinHostPort(p.Host, port), p.DatabaseName, p.SSLMode,
		p.Schema, defaultSearchPath, AppName)
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
	DatasourceCommon `yaml:",inline" json:",inline"`

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
	// For example, 'azure?emulator=127.0.0.1:10000&sas=0' or 'google'.
	Connection string `yaml:"connection" json:"connection" validate:"required"`

	// Username of the storage account, like devstoreaccount1 when using Azurite.
	User string `yaml:"user" json:"user" validate:"required"`

	// Some kind of credential like a password or key to authenticate with the storage backend, e.g:
	// 'Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==' when using Azurite.
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
