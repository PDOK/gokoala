{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
# {{ .Config.Title }} {{ i18n "Tiles" }}

{{ i18n "TilesTextHTML" }}{{ if .Config.OgcAPI.Styles }} {{ i18n "WithStylesHTML" }}{{ end }}

{{ $baseUrlTiles := .Params.BaseURL }}
{{ $projections := .Params.AllProjections }}

## Tile Sets

{{ range $srs := .Params.SupportedSrs -}}
{{ $projection := get $projections $srs.Srs -}}
### {{ $projection }}

| Property | Value |
|---|---|
| Tile Matrix Set | {{ $projection }} |
| CRS | `{{ $srs.Srs }}` |
| Type | {{ (index $.Params.Types 0) | toString | title }} |
| URL template | `{{ $baseUrlTiles }}/tiles/{{ $projection }}/{z}/{y}/{x}?f=mvt` |
| Metadata | [{{ i18n "View" }} metadata]({{ $baseUrlTiles }}/tiles/{{ $projection }}) |

{{ if $.Config.LastUpdatedBy -}}
**{{ i18n "UpdatedBy" }} {{ $.Config.LastUpdatedBy }} {{ i18n "On" }}:** {{ toDate "2006-01-02T15:04:05Z07:00" $.Config.LastUpdated | date "2006-01-02" }}
{{ else if $.Config.LastUpdated -}}
**{{ i18n "LastUpdated" }}:** {{ toDate "2006-01-02T15:04:05Z07:00" $.Config.LastUpdated | date "2006-01-02" }}
{{- end }}

{{ end -}}
{{/* @formatter:on */}}