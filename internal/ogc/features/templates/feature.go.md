{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
{{ $cfg := .Config -}}
{{ $baseUrl := $cfg.BaseURL -}}
{{ $mapSheetProperties := .Params.MapSheetProperties -}}
{{ $webConfig := .Params.WebConfig -}}

# {{ .Config.Title }} {{ if and .Params.Metadata .Params.Metadata.Title }}{{ .Params.Metadata.Title }}{{ else }}. Collection: {{ .Params.CollectionID }}{{ end }}. Feature: {{ .Params.FeatureID }}

**CRS:** `http://www.opengis.net/def/crs/OGC/1.3/CRS84`

Available CRS options:
{{ range $index, $srs := .Config.OgcAPI.Features.CollectionSRS $.Params.CollectionID }}
- http://www.opengis.net/def/crs/EPSG/0/{{ trimPrefix "EPSG:" $srs }}
{{- end }}

## Properties

{{ $skipKeys := list "" -}}
{{ if $mapSheetProperties -}}
{{ $skipKeys = append $skipKeys $mapSheetProperties.Size -}}
{{- end }}

| Property | Value |
|---|---|
{{- range $key := .Params.Keys }}
    {{- if not (has $key $skipKeys) }}
        {{- $value := $.Params.Properties.Value $key }}
        {{- if isdate $value }}
| {{ $key }} | {{ dateInZone "2006-01-02T15:04:05Z" $value "UTC" }} |
        {{- else if hasSuffix ".href" $key }}
        {{- if isstringslice $value }}
| {{ $key }} | {{ range $relationLink := $value }}[{{ $relationLink }}]({{ $relationLink }}) {{ end }} |
        {{- else }}
| {{ $key }} | [{{ $value }}]({{ $value }}) |
        {{- end }}
        {{- else if and $webConfig $webConfig.URLAsHyperlink (islink $value) }}
| {{ $key }} | [{{ $value }}]({{ $value }}) |
        {{- else }}
| {{ $key }} | {{ default "" $value }} |
        {{- end }}
    {{- end }}
{{- end }}

{{ if and $mapSheetProperties .Params.Properties -}}
[Download]({{ .Params.Properties.Value $mapSheetProperties.AssetURL }}) ({{ i18n "Size" }}: {{ humansize (.Params.Properties.Value $mapSheetProperties.Size) }})
{{- end }}

{{/* @formatter:on */}}