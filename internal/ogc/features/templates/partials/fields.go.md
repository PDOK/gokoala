{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
{{define "fields"}}
{{- range $feat := .Params.Fields -}}
{{ $typeFormat := $feat.ToTypeFormat -}}
{{ if $feat.IsFid }}{{ continue }}{{ end -}}
{{ if $feat.IsExternalFid }}{{ continue }}{{ end -}}
{{ if $feat.IsPrimaryGeometry -}}
| geometry | {{ $typeFormat.Type }} | {{ if $feat.IsRequired }}{{ i18n "RequiredYes" }}{{ else }}{{ i18n "RequiredNo" }}{{ end }} | {{ $feat.Description }} |
{{ else -}}
| {{ if $feat.FeatureRelation }}{{ $feat.FeatureRelation.Name }}{{ else }}{{ $feat.Name }}{{ end }} | {{ if $typeFormat.Format }}{{ $typeFormat.Format }}{{ else }}{{ $typeFormat.Type }}{{ end }} | {{ if $feat.IsRequired }}{{ i18n "RequiredYes" }}{{ else }}{{ i18n "RequiredNo" }}{{ end }} | {{ if $feat.IsPrimaryIntervalStart }}{{ i18n "FeatureIntervalStartDescription" }} {{ else if $feat.IsPrimaryIntervalEnd }}{{ i18n "FeatureIntervalEndDescription" }} {{ else if $feat.FeatureRelation }}{{ i18n "FeatureRelationDescription" }} {{ end }}{{ $feat.Description }} |
{{ end -}}
{{ end -}}
{{end}}

{{/* @formatter:on */}}