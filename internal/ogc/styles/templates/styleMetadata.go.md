{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
{{ if .Params -}}

{{ $baseUrl := .Config.BaseURL -}}
{{ $style := .Params.Metadata.ID -}}
{{ $projection := .Params.Projection -}}
# {{ .Config.Title }} {{ .Params.Metadata.Title }} Metadata

{{ .Params.Metadata.Description }}

{{ if .Params.Metadata.Keywords -}}
**{{ i18n "Keywords" }}:** {{ .Params.Metadata.Keywords | join ", " }}
{{- end }}
{{ if .Params.Metadata.LastUpdated -}}
**{{ i18n "LastUpdated" }}:** {{ default "-" (toDate "2006-01-02T15:04:05Z07:00" .Params.Metadata.LastUpdated | date "2006-01-02") }}
{{- end }}
{{ if .Params.Metadata.Version -}}
**{{ i18n "Version" }}:** {{ default "-" .Params.Metadata.Version }}
{{- end }}
**{{ i18n "License" }}:** [{{ .Config.License.Name }}]({{ .Config.License.URL }})
{{ if .Config.Support -}}
**{{ i18n "Support" }}:** [{{ .Config.Support.Name }}]({{ .Config.Support.URL }})
{{- end }}

{{ range $sh_index, $styleFormat := .Params.Metadata.Formats -}}
    {{ if eq $styleFormat.Format "mapbox" -}}
**Styling:** [{{ i18n "View" }} Mapbox Style]({{ $baseUrl }}/styles/{{ $style }}__{{ lower $projection }}?f=mapbox)
    {{ else if eq $styleFormat.Format "sld10" -}}
**Styling:** [{{ i18n "View" }} SLD 1.0 Style]({{ $baseUrl }}/styles/{{ $style }}__{{ lower $projection }}?f=sld10)
    {{ end -}}
{{- end }}
    
{{ if and .Params.Metadata.Legend .Config.Resources -}}
## {{ i18n "Legend" }}

![{{ .Params.Metadata.Legend }} {{ i18n "Legend" }}]({{ $baseUrl }}/styles/{{ $style }}__{{ lower $projection }}/legend)
{{- end }}

{{- end }}
{{/* @formatter:on */}}