{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}

{{ if .Params -}}
{{ $baseUrl := .Config.BaseURL -}}
# {{ .Config.Title }} {{ .Params.Title }}

{{ .Params.Description }}

**Style URL:** [{{ $baseUrl }}/styles/{{ .Params.ID }}?f=mapbox]({{ $baseUrl }}/styles/{{ .Params.ID }}?f=mapbox)
{{- end }}

{{/* @formatter:on */}}