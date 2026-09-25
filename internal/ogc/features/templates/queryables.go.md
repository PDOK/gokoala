{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
# {{ .Config.Title }}. Queryables of {{ .Params.CollectionTitle }}

{{ i18n "QueryablesDescription" }} 

[JSON Schema]({{ .Config.BaseURL }}/collections/{{ .Params.CollectionID }}/queryables?f=json).

| {{ i18n "FieldName" }} | {{ i18n "DataType" }} | {{ i18n "Required" }} | {{ i18n "DescriptionLabel" }} |
|---|---|---|---|
{{block "fields" .}}{{end}}

{{/* @formatter:on */}}