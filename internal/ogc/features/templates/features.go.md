{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
{{ $cfg := .Config }}
{{ $baseUrl := $cfg.BaseURL }}
{{ $mapSheetProperties := .Params.MapSheetProperties }}
{{ $webConfig := .Params.WebConfig }}

# {{ .Config.Title }} {{ if and .Params.Metadata .Params.Metadata.Title }}{{ .Params.Metadata.Title }}{{ else }} Collection: {{ .Params.CollectionID }}{{ end }}

{{ if and .Params.Metadata .Params.Metadata.Description -}}
    {{ truncate .Params.Metadata.Description 400 }}
{{- end }}

{{template "feature_forms" .}}

## Features

{{ if .Params.Cursor.HasPrev -}}
[Paginate to previous features]({{ .Params.PrevLink }})
{{- end }}
{{ if .Params.Cursor.HasNext -}}
[Paginate to next features]({{ .Params.NextLink }})
{{- end }}

{{ range $feat := .Params.Features -}}
### [Feature: {{ $feat.ID }}]({{ $baseUrl }}/collections/{{ $.Params.CollectionID }}/items/{{ $feat.ID }})

{{ $skipKeys := list "" -}}
{{ if $mapSheetProperties -}}
{{ $skipKeys = append $skipKeys $mapSheetProperties.Size -}}
{{- end }}

| Property | Value |
|---|---|
{{- range $key := $feat.Keys }}
    {{- if not (has $key $skipKeys) }}
        {{- $value := $feat.Properties.Value $key }}
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

{{ if and $mapSheetProperties $feat.Properties -}}
[Download]({{ $feat.Properties.Value $mapSheetProperties.AssetURL }}) ({{ i18n "Size" }}: {{ humansize ($feat.Properties.Value $mapSheetProperties.Size) }})
{{- end }}

{{ end -}}

{{/* @formatter:on */}}