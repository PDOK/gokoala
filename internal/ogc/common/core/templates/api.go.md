{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
# {{ .Config.Title }} OpenAPI {{ i18n "Specification" }}

{{ .Config.Abstract }}

**{{ i18n "License" }}:** [{{ .Config.License.Name }}]({{ .Config.License.URL }})
{{ if .Config.Support -}}
**{{ i18n "Support" }}:** [{{ .Config.Support.Name }}]({{ .Config.Support.URL }})
{{- end }}

The full OpenAPI specification is available as [JSON]({{ .Config.BaseURL }}/api?f=json).

{{/* @formatter:on */}}