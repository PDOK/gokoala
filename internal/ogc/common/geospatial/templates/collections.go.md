{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
# {{ .Config.Title }} {{ i18n "Collections" }}

{{ $cfg := .Config -}}
{{ $baseUrl := $cfg.BaseURL -}}
{{ $collTypes := .Params -}}

{{ range $index, $coll := .Config.AllCollections.Unique -}}
{{ $geomType := $collTypes.GetGeometryType $coll.ID -}}

## {{ if and $coll.Metadata $coll.Metadata.Title }}{{ $coll.Metadata.Title }}{{ else }}{{ $coll.ID }}{{ end }}

{{ if and $coll.Metadata $coll.Metadata.Description -}}
    {{ truncate $coll.Metadata.Description 500 }}
{{- end }}

{{ if and $coll.Metadata $coll.Metadata.Keywords -}}
**{{ i18n "Keywords" }}:** {{ truncateslice ($coll.Metadata.Keywords | join ", ") 400 }}
{{- end }}

{{ if and $cfg.OgcAPI.Features $cfg.OgcAPI.Features.Collections -}}
{{ if $cfg.OgcAPI.Features.Collections.ContainsID $coll.ID -}}
**Schema:** ({{ $baseUrl }}/collections/{{ $coll.ID }}/schema)
{{- end }}
{{- end }}

{{ if and $coll.Metadata $coll.Metadata.LastUpdated -}}
    {{ if and $coll.Metadata $coll.Metadata.LastUpdatedBy -}}
**{{ i18n "UpdatedBy" }} {{ $coll.Metadata.LastUpdatedBy }} {{ i18n "On" }}:** {{ toDate "2006-01-02T15:04:05Z07:00" $coll.Metadata.LastUpdated | date "2006-01-02" }}
    {{ else if $cfg.LastUpdatedBy -}}
**{{ i18n "UpdatedBy" }} {{ $cfg.LastUpdatedBy }} {{ i18n "On" }}:** {{ toDate "2006-01-02T15:04:05Z07:00" $coll.Metadata.LastUpdated | date "2006-01-02" }}
    {{ else -}}
**{{ i18n "LastUpdated" }}:** {{ toDate "2006-01-02T15:04:05Z07:00" $coll.Metadata.LastUpdated | date "2006-01-02" }}
    {{- end }}
{{ else if $cfg.LastUpdated -}}
    {{ if $cfg.LastUpdatedBy -}}
**{{ i18n "UpdatedBy" }} {{ $cfg.LastUpdatedBy }} {{ i18n "On" }}:** {{ toDate "2006-01-02T15:04:05Z07:00" $cfg.LastUpdated | date "2006-01-02" }}
    {{ else -}}
**{{ i18n "LastUpdated" }}:** {{ toDate "2006-01-02T15:04:05Z07:00" $cfg.LastUpdated | date "2006-01-02" }}
    {{- end }}
{{- end }}

{{ if and $coll.Metadata $coll.Metadata.Extent (ne $geomType "none") -}}
**{{ i18n "GeographicExtent" }}{{ if $coll.Metadata.Extent.Srs }} ({{ $coll.Metadata.Extent.Srs }}){{ else }} (CRS84){{ end }}:** {{ $coll.Metadata.Extent.Bbox | join ", " }}
{{- end }}

{{ if and $coll.Metadata $coll.Metadata.Extent $coll.Metadata.Extent.Interval -}}
**{{ i18n "TemporalExtent" }} (ISO-8601):** {{ toDate "2006-01-02T15:04:05Z" ((first $coll.Metadata.Extent.Interval) | replace "\"" "") | date "2006-01-02" }} / {{ if not (contains "null" (last $coll.Metadata.Extent.Interval)) }}{{ toDate "2006-01-02T15:04:05Z" ((last $coll.Metadata.Extent.Interval) | replace "\"" "") | date "2006-01-02" }}{{ else }}..{{ end }}
{{- end }}

{{ if and $coll.Links $coll.Links.Downloads -}}
### Downloads

{{ range $link := $coll.Links.Downloads -}}
- [{{ $link.Name }}{{ if $link.Size }} ({{ $link.Size }}){{ end }}]({{ $link.AssetURL }})
{{ end -}}
{{- end }}

{{ if and $cfg.OgcAPI.FeaturesSearch $cfg.OgcAPI.FeaturesSearch.Collections -}}
    {{ if $cfg.OgcAPI.FeaturesSearch.Collections.ContainsID $coll.ID -}}
        {{ range $idxSearch, $collSearch := $cfg.OgcAPI.FeaturesSearch.Collections -}}
            {{ if eq $collSearch.ID $coll.ID -}}
### {{ i18n "Details" }}
**{{ i18n "Version" }}:** {{ $collSearch.Version }}

**{{ i18n "DisplayedAs" }}:**
{{ range $field := $collSearch.Fields -}}
- {{ $field }}
{{ end -}}

**{{ i18n "DisplayNameExample" }}:** {{ $collSearch.DisplayNameExample }}

            {{ if $collSearch.CollectionFilter -}}
**{{ i18n "CollectionFilter" }}:** {{ $collSearch.CollectionFilter }}
            {{- end }}
        {{- end }}
    {{- end }}
{{- end }}
{{- end }}

---

{{ end -}}

{{/* @formatter:on */}}