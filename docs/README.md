# Configuration Reference
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|version|[string](#string)|true| |Version of the API. When releasing a new version which contains backwards-incompatible changes, a new major version must be released.|
|title|[string](#string)|true| |Human friendly title of the API. Don&#39;t include "OGC API" in the title, this is added automatically.|
|serviceIdentifier|[string](#string)|true| |Shorted title / abbreviation describing the API.|
|abstract|[string](#string)|true| |Human friendly description of the API and dataset.|
|license|[License](#License)|true| |Licensing term that apply to this API and dataset.|
|baseUrl|[URL](#URL)|true| |The base URL - that&#39;s the part until the OGC API landing page - under which this API is served.|
|datasetCatalogUrl|[URL](#URL)|false| |Optional reference to a catalog/portal/registry that lists all datasets, not just this one<br>&#43;optional.|
|availableLanguages.&#91;&#93; |[Language](#Language)|false| |The languages/translations to offer, valid options are Dutch &#40;nl&#41; and English &#40;en&#41;. Dutch is the default.<br>&#43;optional.|
|ogcApi|[OgcAPI](#OgcAPI)|true| |Define which OGC API building blocks this API supports.|
|collectionOrder.&#91;&#93; |[string](#string)|false| |Order in which collections &#40;containing features, tiles, 3d tiles, etc.&#41; should be returned.<br>When not specified collections are returned in alphabetic order.<br>&#43;optional.|
|thumbnail|[string](#string)|false| |Reference to a PNG image to use a thumbnail on the landing page.<br>The full path is constructed by appending Resources &#43; Thumbnail.<br>&#43;optional.|
|keywords.&#91;&#93; |[string](#string)|false| |Keywords to make this API beter discoverable<br>&#43;optional.|
|lastUpdated|[string](#string)|false| |Moment in time when the dataset was last updated<br>&#43;optional<br>&#43;kubebuilder:validation:Type&#61;string<br>&#43;kubebuilder:validation:Format&#61;"date-time".|
|lastUpdatedBy|[string](#string)|false| |Who updated the dataset<br>&#43;optional.|
|support|[Support](#Support)|false| |Available support channels<br>&#43;optional.|
|datasetDetails.&#91;&#93; |[DatasetDetail](#DatasetDetail)|false| |Key/value pairs to add extra information to the landing page<br>&#43;optional.|
|resources|[Resources](#Resources)|false| |Location where resources &#40;e.g. thumbnails&#41; specific to the given dataset are hosted<br>&#43;optional.|

## License
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|name|[string](#string)|true| |Name of the license, e.g. MIT, CC0, etc.|
|url|[URL](#URL)|true| |URL to license text on the web.|

## URL
URL Custom net.URL compatible with YAML and JSON &#40;un&#41;marshalling and kubebuilder.
In addition, it also removes trailing slash if present, so we can easily
append a longer path without having to worry about double slashes.

Allow only http/https URLs or environment variables like $&#123;FOOBAR&#125;
&#43;kubebuilder:validation:Pattern&#61;`^&#40;https&#63;://.&#43;&#41;&#124;&#40;&#92;$&#92;&#123;.&#43;&#92;&#125;.&#42;&#41;`
&#43;kubebuilder:validation:Type&#61;string.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|

## Language
Language represents a BCP 47 language tag.
&#43;kubebuilder:validation:Type&#61;string.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|

## OgcAPI
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|3dgeovolumes|[OgcAPI3dGeoVolumes](#OgcAPI3dGeoVolumes)|false| |Enable when this API should offer OGC API 3D GeoVolumes. This includes OGC 3D Tiles.<br>&#43;optional.|
|tiles|[OgcAPITiles](#OgcAPITiles)|false| |Enable when this API should offer OGC API Tiles. This also requires OGC API Styles.<br>&#43;optional.|
|styles|[OgcAPIStyles](#OgcAPIStyles)|false| |Enable when this API should offer OGC API Styles.<br>&#43;optional.|
|features|[OgcAPIFeatures](#OgcAPIFeatures)|false| |Enable when this API should offer OGC API Features.<br>&#43;optional.|
|processes|[OgcAPIProcesses](#OgcAPIProcesses)|false| |Enable when this API should offer OGC API Processes.<br>&#43;optional.|

## Support
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|name|[string](#string)|true| |Name of the support organization.|
|url|[URL](#URL)|true| |URL to external support webpage<br>&#43;kubebuilder:validation:Type&#61;string.|
|email|[string](#string)|false| |Email for support questions<br>&#43;optional.|

## DatasetDetail
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|name|[string](#string)|true| |Arbitrary name to add extra information to the landing page.|
|value|[string](#string)|true| |Arbitrary value associated with the given name.|

## Resources
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|url|[URL](#URL)|false| |Location where resources &#40;e.g. thumbnails&#41; specific to the given dataset are hosted. This is optional if Directory is set<br>&#43;optional.|
|directory|[string](#string)|false| |// Location where resources &#40;e.g. thumbnails&#41; specific to the given dataset are hosted. This is optional if URL is set<br>&#43;optional.|

## OgcAPI3dGeoVolumes
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|tileServer|[URL](#URL)|true| |Reference to the server &#40;or object storage&#41; hosting the 3D Tiles.|
|collections|[GeoSpatialCollections](#GeoSpatialCollections)|true| |Collections to be served as 3D GeoVolumes.|
|validateResponses|[bool](#bool)|false|true|Whether JSON responses will be validated against the OpenAPI spec<br>since it has significant performance impact when dealing with large JSON payloads.<br><br>&#43;kubebuilder:default&#61;true<br>&#43;optional.ptr due to https://github.com/creasty/defaults/issues/49.|

## OgcAPITiles
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|tileServer|[URL](#URL)|true| |Reference to the server &#40;or object storage&#41; hosting the tiles.<br>Note: Only marked as optional in CRD to support top-level OR collection-level tiles<br>&#43;optional.|
|types.&#91;&#93; |[TilesType](#TilesType)|true| |Could be &#39;vector&#39; and/or &#39;raster&#39; to indicate the types of tiles offered<br>Note: Only marked as optional in CRD to support top-level OR collection-level tiles<br>&#43;optional.|
|supportedSrs.&#91;&#93; |[SupportedSrs](#SupportedSrs)|true| |Specifies in what projections &#40;SRS/CRS&#41; the tiles are offered<br>Note: Only marked as optional in CRD to support top-level OR collection-level tiles<br>&#43;optional.|
|uriTemplateTiles|[string](#string)|false| |Optional template to the vector tiles on the tileserver. Defaults to &#123;tms&#125;/&#123;z&#125;/&#123;x&#125;/&#123;y&#125;.pbf.<br>&#43;optional.|
|collections|[GeoSpatialCollections](#GeoSpatialCollections)|false| |Tiles per collection. When no collections are specified tiles should be hosted at the root of the API &#40;/tiles endpoint&#41;.<br>&#43;optional.|

## OgcAPIStyles
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|default|[string](#string)|true| |ID of the style to use a default.|
|stylesDir|[string](#string)|true| |Location on disk where the styles are hosted.|
|supportedStyles.&#91;&#93; |[Style](#Style)|true| |Styles exposed though this API.|

## OgcAPIFeatures
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|basemap|[string](#string)|false|OSM|Basemap to use in embedded viewer on the HTML pages.<br>&#43;kubebuilder:default&#61;"OSM"<br>&#43;kubebuilder:validation:Enum&#61;OSM;BRT<br>&#43;optional.|
|collections|[GeoSpatialCollections](#GeoSpatialCollections)|true| |Collections to be served as features through this API.|
|limit|[Limit](#Limit)|false| |Limits the amount of features to retrieve with a single call<br>&#43;optional.|
|datasources|[Datasources](#Datasources)|false| |One or more datasources to get the features from &#40;geopackages, postgis, etc&#41;.<br>Optional since you can also define datasources at the collection level<br>&#43;optional.|
|validateResponses|[bool](#bool)|false|true|Whether GeoJSON/JSON-FG responses will be validated against the OpenAPI spec<br>since it has significant performance impact when dealing with large JSON payloads.<br><br>&#43;kubebuilder:default&#61;true<br>&#43;optional.ptr due to https://github.com/creasty/defaults/issues/49.|

## OgcAPIProcesses
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|supportsDismiss|[bool](#bool)|true| |Enable to advertise dismiss operations on the conformance page.|
|supportsCallback|[bool](#bool)|true| |Enable to advertise callback operations on the conformance page.|
|processesServer|[URL](#URL)|true| |Reference to an external service implementing the process API. GoKoala acts only as a proxy for OGC API Processes.|

## TilesType
**Type:** string
+kubebuilder:validation:Enum=raster;vector.

| Enum Value      | Describe          |
|----------|--------------|
|"raster"||
|"vector"||

## SupportedSrs
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|srs|[string](#string)|true| |Projection &#40;SRS/CRS&#41; used<br>&#43;kubebuilder:validation:Pattern&#61;`^EPSG:&#92;d&#43;$`.|
|zoomLevelRange|[ZoomLevelRange](#ZoomLevelRange)|true| |Available zoom levels.|

## Style
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|id|[string](#string)|true| |Unique ID of this style.|
|title|[string](#string)|true| |Human-friendly name of this style.|
|description|[string](#string)|false| |Explains what is visualized by this style<br>&#43;optional.|
|keywords.&#91;&#93; |[string](#string)|false| |Keywords to make this style better discoverable<br>&#43;optional.|
|lastUpdated|[string](#string)|false| |Moment in time when the style was last updated<br>&#43;optional<br>&#43;kubebuilder:validation:Type&#61;string<br>&#43;kubebuilder:validation:Format&#61;"date-time".|
|version|[string](#string)|false| |Optional version of this style<br>&#43;optional.|
|thumbnail|[string](#string)|false| |Reference to a PNG image to use a thumbnail on the style metadata page.<br>The full path is constructed by appending Resources &#43; Thumbnail.<br>&#43;optional.|
|formats.&#91;&#93; |[StyleFormat](#StyleFormat)|true| |This style is offered in the following formats.|

## Limit
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|default|[int](#int)|false|10|Number of features to return by default.<br>&#43;kubebuilder:default&#61;10<br>&#43;kubebuilder:validation:Minimum&#61;2<br>&#43;optional.|
|max|[int](#int)|false|1000|Max number of features to return. Should be larger than 100 since the HTML interface always offers a 100 limit option.<br>&#43;kubebuilder:default&#61;1000<br>&#43;kubebuilder:validation:Minimum&#61;100<br>&#43;optional.|

## Datasources
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|defaultWGS84|[Datasource](#Datasource)|true| |Features should always be available in WGS84 &#40;according to spec&#41;.<br>This specifies the datasource to be used for features in the WGS84 projection.|
|additional.&#91;&#93; |[AdditionalDatasource](#AdditionalDatasource)|true| |One or more additional datasources for features in other projections. GoKoala doesn&#39;t do<br>any on-the-fly reprojection so additional datasources need to be reprojected ahead of time.<br>&#43;optional.|

## ZoomLevelRange
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|start|[int](#int)|true| |Start zoom level<br>&#43;kubebuilder:validation:Minimum&#61;0.|
|end|[int](#int)|true| |End zoom level.|

## StyleFormat
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|format|[string](#string)|false|mapbox|Name of the format<br>&#43;kubebuilder:default&#61;"mapbox"<br>&#43;optional.|

## Datasource
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|geopackage|[GeoPackage](#GeoPackage)|false| |GeoPackage to get the features from.<br>&#43;optional.|
|postgis|[PostGIS](#PostGIS)|false| |PostGIS database to get the features from &#40;not implemented yet&#41;.<br>&#43;optional.|

## AdditionalDatasource
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|srs|[string](#string)|true| |Projection &#40;SRS/CRS&#41; used for the features in this datasource<br>&#43;kubebuilder:validation:Pattern&#61;`^EPSG:&#92;d&#43;$`.|
|geopackage|[GeoPackage](#GeoPackage)|false| |GeoPackage to get the features from.<br>&#43;optional.|
|postgis|[PostGIS](#PostGIS)|false| |PostGIS database to get the features from &#40;not implemented yet&#41;.<br>&#43;optional.|

## GeoPackage
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|local|[GeoPackageLocal](#GeoPackageLocal)|false| |Settings to read a GeoPackage from local disk<br>&#43;optional.|
|cloud|[GeoPackageCloud](#GeoPackageCloud)|false| |Settings to read a GeoPackage as a Cloud-Backed SQLite database<br>&#43;optional.|

## PostGIS
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|

## GeoPackageLocal
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|fid|[string](#string)|false|fid|Feature id column name<br>&#43;kubebuilder:default&#61;"fid"<br>&#43;optional.|
|externalFid|[string](#string)|true| |External feature id column name. When specified this ID column will be exposed to clients instead of the regular FID column.<br>It allows one to offer a more stable ID to clients instead of an auto-generated FID. External FID column should contain UUIDs.<br>&#43;optional.|
|queryTimeout|[Duration](#Duration)|false|15s|Optional timeout after which queries are canceled<br>&#43;kubebuilder:default&#61;"15s"<br>&#43;optional.|
|maxBBoxSizeToUseWithRTree|[int](#int)|false|8000|ADVANCED SETTING. When the number of features in a bbox stay within the given value use an RTree index, otherwise use a BTree index.<br>&#43;kubebuilder:default&#61;8000<br>&#43;optional.|
|inMemoryCacheSize|[int](#int)|false|-2000|ADVANCED SETTING. Sets the SQLite "cache_size" pragma which determines how many pages are cached in-memory.<br>See https://sqlite.org/pragma.html#pragma_cache_size for details.<br>Default in SQLite is 2000 pages, which equates to 2000KiB &#40;2048000 bytes&#41;. Which is denoted as -2000.<br>&#43;kubebuilder:default&#61;-2000<br>&#43;optional.|
|file|[string](#string)|true| |Location of GeoPackage on disk.<br>You can place the GeoPackage here manually &#40;out-of-band&#41; or you can specify Download<br>and let the application download the GeoPackage for you and store it at this location.|
|download|[GeoPackageDownload](#GeoPackageDownload)|false| |Optional initialization task to download a GeoPackage during startup. GeoPackage will be<br>downloaded to local disk and stored at the location specified in File.<br>&#43;optional.|

## GeoPackageCloud
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|fid|[string](#string)|false|fid|Feature id column name<br>&#43;kubebuilder:default&#61;"fid"<br>&#43;optional.|
|externalFid|[string](#string)|true| |External feature id column name. When specified this ID column will be exposed to clients instead of the regular FID column.<br>It allows one to offer a more stable ID to clients instead of an auto-generated FID. External FID column should contain UUIDs.<br>&#43;optional.|
|queryTimeout|[Duration](#Duration)|false|15s|Optional timeout after which queries are canceled<br>&#43;kubebuilder:default&#61;"15s"<br>&#43;optional.|
|maxBBoxSizeToUseWithRTree|[int](#int)|false|8000|ADVANCED SETTING. When the number of features in a bbox stay within the given value use an RTree index, otherwise use a BTree index.<br>&#43;kubebuilder:default&#61;8000<br>&#43;optional.|
|inMemoryCacheSize|[int](#int)|false|-2000|ADVANCED SETTING. Sets the SQLite "cache_size" pragma which determines how many pages are cached in-memory.<br>See https://sqlite.org/pragma.html#pragma_cache_size for details.<br>Default in SQLite is 2000 pages, which equates to 2000KiB &#40;2048000 bytes&#41;. Which is denoted as -2000.<br>&#43;kubebuilder:default&#61;-2000<br>&#43;optional.|
|connection|[string](#string)|true| |Reference to the cloud storage &#40;either azure or google at the moment&#41;.<br>For example &#39;azure&#63;emulator&#61;127.0.0.1:10000&#38;sas&#61;0&#39; or &#39;google&#39;.|
|user|[string](#string)|true| |Username of the storage account, like devstoreaccount1 when using Azurite.|
|auth|[string](#string)|true| |Some kind of credential like a password or key to authenticate with the storage backend, e.g:<br>&#39;Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw&#61;&#61;&#39; when using Azurite.|
|container|[string](#string)|true| |Container/bucket on the storage account.|
|file|[string](#string)|true| |Filename of the GeoPackage.|
|cache|[GeoPackageCloudCache](#GeoPackageCloudCache)|false| |Local cache of fetched blocks from cloud storage<br>&#43;optional.|
|logHttpRequests|[bool](#bool)|false|false|ADVANCED SETTING. Only for debug purposes&#33; When true all HTTP requests executed by sqlite to cloud object storage are logged to stdout<br>&#43;kubebuilder:default&#61;false<br>&#43;optional.|

## Duration
Duration Custom time.Duration compatible with YAML and JSON &#40;un&#41;marshalling and kubebuilder.
&#40;Already supported in yaml/v3 but not encoding/json.&#41;

&#43;kubebuilder:validation:Type&#61;string
&#43;kubebuilder:validation:Format&#61;duration.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|

## GeoPackageDownload
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|from|[URL](#URL)|true| |Location of GeoPackage on remote HTTP&#40;S&#41; URL. GeoPackage will be downloaded to local disk<br>during startup and stored at the location specified in "file".|
|parallelism|[int](#int)|false|4|ADVANCED SETTING. Determines how many workers &#40;goroutines&#41; in parallel will download the specified GeoPackage.<br>Setting this to 1 will disable concurrent downloads.<br>&#43;kubebuilder:default&#61;4<br>&#43;kubebuilder:validation:Minimum&#61;1<br>&#43;optional.|
|tlsSkipVerify|[bool](#bool)|false|false|ADVANCED SETTING. When true TLS certs are NOT validated, false otherwise. Only use true for your own self-signed certificates&#33;<br>&#43;kubebuilder:default&#61;false<br>&#43;optional.|
|timeout|[Duration](#Duration)|false|2m|ADVANCED SETTING. HTTP request timeout when downloading &#40;part of&#41; GeoPackage.<br>&#43;kubebuilder:default&#61;"2m"<br>&#43;optional.|
|retryDelay|[Duration](#Duration)|false|1s|ADVANCED SETTING. Minimum delay to use when retrying HTTP request to download &#40;part of&#41; GeoPackage.<br>&#43;kubebuilder:default&#61;"1s"<br>&#43;optional.|
|retryMaxDelay|[Duration](#Duration)|false|30s|ADVANCED SETTING. Maximum overall delay of the exponential backoff while retrying HTTP requests to download &#40;part of&#41; GeoPackage.<br>&#43;kubebuilder:default&#61;"30s"<br>&#43;optional.|
|maxRetries|[int](#int)|false|5|ADVANCED SETTING. Maximum number of retries when retrying HTTP requests to download &#40;part of&#41; GeoPackage.<br>&#43;kubebuilder:default&#61;5<br>&#43;kubebuilder:validation:Minimum&#61;1<br>&#43;optional.|

## GeoPackageCloudCache
&#43;kubebuilder:object:generate&#61;true.

| Key      | Type      | Require | Default           | Describe          |
|----------|----------|-----|------------------|--------------|
|path|[string](#string)|false| |Optional path to directory for caching cloud-backed GeoPackage blocks, when omitted a temp dir will be used.<br>&#43;optional.|
|maxSize|[string](#string)|false|1Gb|Max size of the local cache. Accepts human-readable size such as 100Mb, 4Gb, 1Tb, etc. When omitted 1Gb is used.<br>&#43;kubebuilder:default&#61;"1Gb"<br>&#43;optional.|
|warmUp|[bool](#bool)|false|false|When true a warm-up query is executed on startup which aims to fill the local cache. Does increase startup time.<br>&#43;kubebuilder:default&#61;false<br>&#43;optional.|

---
**github.com/PDOK/gokoala/config.Config**
GENERATED BY THE COMMAND [type2md](https://github.com/eleztian/type2md)
