{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
# {{ .Config.Title }}. Schema of {{ .Params.CollectionTitle }}

{{ i18n "SchemaDescription" }} 

[JSON Schema]({{ .Config.BaseURL }}/collections/{{ .Params.CollectionID }}/schema?f=json).

{{ if .Params.CQLEnabled -}}
[Queryables]({{ .Config.BaseURL }}/collections/{{ .Params.CollectionID }}/queryables).
{{- end }}

| {{ i18n "FieldName" }} | {{ i18n "DataType" }} | {{ i18n "Required" }} | {{ i18n "DescriptionLabel" }} |
|---|---|---|---|
| id | {{ if .Params.HasExternalFid }}uuid{{ else }}integer{{ end }} | {{ i18n "RequiredYes" }} | {{ i18n "FeatureIdDescription" }}{{ if .Params.HasExternalFid }} {{ i18n "FeatureIdStableOverTime" }}{{ end }} |
{{block "fields" .}}{{end}}

{{/* @formatter:on */}}