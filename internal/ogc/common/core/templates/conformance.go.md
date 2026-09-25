{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}

# {{ .Config.Title }} Conformance

{{ i18n "ConformanceAbstract" }}

## Common

| Conformance                                                | Status                |
|------------------------------------------------------------|-----------------------|
| http://www.opengis.net/spec/ogcapi-common-1/1.0/conf/core  | {{ i18n "Standard" }} |
| http://www.opengis.net/spec/ogcapi-common-1/1.0/conf/json  | {{ i18n "Standard" }} |
| http://www.opengis.net/spec/ogcapi-common-1/1.0/conf/html  | {{ i18n "Standard" }} |
| http://www.opengis.net/spec/ogcapi-common-1/1.0/conf/oas30 | {{ i18n "Standard" }} |
{{- if .Config.HasCollections }}
| http://www.opengis.net/spec/ogcapi-common-2/1.0/conf/collections | {{ i18n "Draft" }} |
{{- end }}

{{ if .Config.OgcAPI.Styles -}}

## Styles

| Conformance                                                        | Status             |
|--------------------------------------------------------------------|--------------------|
| http://www.opengis.net/spec/ogcapi-styles-1/1.0/conf/core          | {{ i18n "Draft" }} |
| http://www.opengis.net/spec/ogcapi-styles-1/1.0/conf/mapbox-styles | {{ i18n "Draft" }} |

{{- end }}

{{ if .Config.OgcAPI.Features -}}

## Features

| Conformance                                                    | Status                |
|----------------------------------------------------------------|-----------------------|
| http://www.opengis.net/spec/ogcapi-features-1/1.0/conf/core    | {{ i18n "Standard" }} |
| http://www.opengis.net/spec/ogcapi-features-1/1.0/conf/html    | {{ i18n "Standard" }} |
| http://www.opengis.net/spec/ogcapi-features-1/1.0/conf/geojson | {{ i18n "Standard" }} |
| http://www.opengis.net/spec/ogcapi-features-2/1.0/conf/crs     | {{ i18n "Standard" }} |
{{- if and .Config.OgcAPI.Features .Config.OgcAPI.Features.SupportsPart3 }}
| http://www.opengis.net/spec/ogcapi-features-3/1.0/conf/filter | {{ i18n "Standard" }} |
| http://www.opengis.net/spec/ogcapi-features-3/1.0/conf/features-filter | {{ i18n "Standard" }} |
| http://www.opengis.net/spec/ogcapi-features-3/1.0/conf/queryables | {{ i18n "Standard" }} |
| http://www.opengis.net/spec/ogcapi-features-3/1.0/conf/queryables-query-parameters | {{ i18n "Standard" }} |
{{- end }}
| http://www.opengis.net/spec/ogcapi-features-5/1.0/conf/schemas | {{ i18n "Draft" }} |
| http://www.opengis.net/spec/ogcapi-features-5/1.0/conf/core-roles-features | {{ i18n "Draft" }} |
| http://www.opengis.net/spec/ogcapi-features-5/1.0/conf/returnables-and-receivables | {{ i18n "Draft" }} |
| http://www.opengis.net/spec/ogcapi-features-5/1.0/conf/feature-references | {{ i18n "Draft" }} |
| http://www.opengis.net/spec/ogcapi-features-5/1.0/conf/profile-parameter | {{ i18n "Draft" }} |
| http://www.opengis.net/spec/ogcapi-features-5/1.0/conf/profile-references | {{ i18n "Draft" }} |
| http://www.opengis.net/spec/json-fg-1/0.2 | {{ i18n "Draft" }} |

{{- end }}

{{ if and .Config.OgcAPI.Features .Config.OgcAPI.Features.SupportsPart3 -}}

## CQL (Common Query Language)

| Conformance                                          | Status                |
|------------------------------------------------------|-----------------------|
| http://www.opengis.net/spec/cql2/1.0/conf/cql2-text  | {{ i18n "Standard" }} |
| http://www.opengis.net/spec/cql2/1.0/conf/basic-cql2 | {{ i18n "Standard" }} |
{{- if .Config.OgcAPI.Features.SupportsAdvancedComparisonOperators }}
| http://www.opengis.net/spec/cql2/1.0/conf/advanced-comparison-operators | {{ i18n "Standard" }} |
{{- end }}
{{- if .Config.OgcAPI.Features.SupportsCaseInsensitiveComparison }}
| http://www.opengis.net/spec/cql2/1.0/conf/case-insensitive-comparison | {{ i18n "Standard" }} |
{{- end }}
{{- if .Config.OgcAPI.Features.SupportsAccentInsensitiveComparison }}
| http://www.opengis.net/spec/cql2/1.0/conf/accent-insensitive-comparison | {{ i18n "Standard" }} |
{{- end }}
{{- if .Config.OgcAPI.Features.SupportsBasicSpatialFunctions }}
| http://www.opengis.net/spec/cql2/1.0/conf/basic-spatial-functions | {{ i18n "Standard" }} |
{{- end }}
{{- if .Config.OgcAPI.Features.SupportsBasicSpatialFunctionsPlus }}
| http://www.opengis.net/spec/cql2/1.0/conf/basic-spatial-functions-plus | {{ i18n "Standard" }} |
{{- end }}
{{- if .Config.OgcAPI.Features.SupportsSpatialFunctions }}
| http://www.opengis.net/spec/cql2/1.0/conf/spatial-functions | {{ i18n "Standard" }} |
{{- end }}
{{- if .Config.OgcAPI.Features.SupportsTemporalFunctions }}
| http://www.opengis.net/spec/cql2/1.0/conf/temporal-functions | {{ i18n "Standard" }} |
{{- end }}

{{- end }}

{{ if .Config.OgcAPI.GeoVolumes -}}

## 3D GeoVolumes

| Conformance                                                   | Status             |
|---------------------------------------------------------------|--------------------|
| http://www.opengis.net/spec/ogcapi-geovolumes-1/1.0/conf/core | {{ i18n "Draft" }} |

{{- end }}

{{ if .Config.OgcAPI.Tiles -}}

## Tiles

| Conformance                                                       | Status                |
|-------------------------------------------------------------------|-----------------------|
| http://www.opengis.net/spec/ogcapi-tiles-1/1.0/conf/core          | {{ i18n "Standard" }} |
| http://www.opengis.net/spec/ogcapi-tiles-1/1.0/conf/tileset       | {{ i18n "Standard" }} |
| http://www.opengis.net/spec/ogcapi-tiles-1/1.0/conf/tilesets-list | {{ i18n "Standard" }} |
{{- if .Config.OgcAPI.Tiles.DatasetTiles }}
| http://www.opengis.net/spec/ogcapi-tiles-1/1.0/conf/dataset-tilesets | {{ i18n "Standard" }} |
{{- end }}
{{- if .Config.OgcAPI.Tiles.Collections }}
| http://www.opengis.net/spec/ogcapi-tiles-1/1.0/conf/geodata-tilesets | {{ i18n "Standard" }} |
{{- end }}
{{- if .Config.OgcAPI.Tiles.HasType "raster" }}
| http://www.opengis.net/spec/ogcapi-tiles-1/1.0/conf/png | {{ i18n "Standard" }} |
{{- end }}
{{- if .Config.OgcAPI.Tiles.HasType "vector" }}
| http://www.opengis.net/spec/ogcapi-tiles-1/1.0/conf/mvt | {{ i18n "Standard" }} |
{{- end }}

{{- end }}

{{ if .Config.OgcAPI.Processes -}}

## Processes

| Conformance                                                                     | Status             |
|---------------------------------------------------------------------------------|--------------------|
| http://www.opengis.net/spec/ogcapi-processes-1/1.0/conf/job-list                | {{ i18n "Draft" }} |
| http://www.opengis.net/spec/ogcapi-processes-1/1.0/conf/ogc-process-description | {{ i18n "Draft" }} |
{{- if .Config.OgcAPI.Processes.SupportsDismiss }}
| http://www.opengis.net/spec/ogcapi-processes-1/1.0/conf/dismiss | {{ i18n "Draft" }} |
{{- end }}
{{- if .Config.OgcAPI.Processes.SupportsCallback }}
| http://www.opengis.net/spec/ogcapi-processes-1/1.0/conf/callback | {{ i18n "Draft" }} |
{{- end }}

{{- end }}

## Non-OGC

| Conformance                                          | Status                |
|------------------------------------------------------|-----------------------|
| https://gitdocumentatie.logius.nl/publicatie/api/adr | {{ i18n "Standard" }} |

{{/* @formatter:on */}}