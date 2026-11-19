{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
{{ if or .Params.HasTiles .Params.HasMaps -}}
# {{ .Config.Title }} {{ i18n "Styles" }}

{{ i18n "StylesText" }}

{{ $baseUrl := .Config.BaseURL -}}

## Available Styles

{{ if .Params.HasTiles -}}
{{ $projections := .Params.AllProjections -}}
{{ range $style := .Config.OgcAPI.Styles.SupportedStyles -}}
{{ range $srs := $.Params.SupportedProjections -}}
{{ $projection := get $projections (index $srs).Srs -}}
### {{ $style.Title }} ({{ $projection }})

- **Style:** [{{ $baseUrl }}/styles/{{ $style.ID }}__{{ lower $projection }}]({{ $baseUrl }}/styles/{{ $style.ID }}__{{ lower $projection }})
- **Format:** Mapbox style
- **Metadata:** [{{ i18n "StyleMetadata" }}]({{ $baseUrl }}/styles/{{ $style.ID }}__{{ lower $projection }}/metadata)

{{ end -}}
{{ end -}}
{{ end -}}
{{ if .Params.HasMaps -}}
{{ range $coll := .Config.OgcAPI.Maps.Collections -}}
{{ range $style := $coll.GetStyles -}}
### {{ $style.Title }}

- **Style:** {{ $.Config.MapsStyleID $coll.ID $style.ID }}
- **Map:** [PNG]({{ $baseUrl }}/collections/{{ $coll.ID }}/styles/{{ $style.ID }}/map?f=png), [JPEG]({{ $baseUrl }}/collections/{{ $coll.ID }}/styles/{{ $style.ID }}/map?f=jpeg), [HTML]({{ $baseUrl }}/collections/{{ $coll.ID }}/styles/{{ $style.ID }}/map?f=html)

{{ end -}}
{{ end -}}
{{ end -}}
{{- end }}

{{/* @formatter:on */}}