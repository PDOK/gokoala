{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}

# {{ .Config.Title }} OpenAPI {{ i18n "Specification" }}

{{ .Config.Abstract }}

**{{ i18n "License" }}:** [{{ .Config.License.Name }}]({{ .Config.License.URL }})
{{ if .Config.Support -}}
**{{ i18n "Support" }}:** [{{ .Config.Support.Name }}]({{ .Config.Support.URL }})
{{- end }}

The full OpenAPI specification is available as [JSON](api?f=json).

{{/* @formatter:on */}}