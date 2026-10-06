{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
# {{ .Config.Title }}. Queryables of {{ .Params.CollectionTitle }}

{{ i18n "QueryablesDescription" }} [JSON Schema]({{ .Config.BaseURL }}/collections/{{ .Params.CollectionID }}/queryables?f=json).

| {{ i18n "FieldName" }} | {{ i18n "DataType" }} | {{ i18n "Required" }} | {{ i18n "DescriptionLabel" }} |
|---|---|---|---|
{{ $idIsIncluded := false -}}
{{ range $feat := .Params.Fields -}}
{{ if or $feat.IsExternalFid $feat.IsFid -}}
{{ if not $idIsIncluded -}}
| id | {{ if $feat.IsExternalFid }} uuid | {{ else }} integer | {{ end }} {{ i18n "RequiredYes"}} | {{ i18n "FeatureIdDescription"}} |
{{ $idIsIncluded = true -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{/* includes fields.go.md */ -}}
{{block "fields" .}}{{end}}

{{/* @formatter:on */}}