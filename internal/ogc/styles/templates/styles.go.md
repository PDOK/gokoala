{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
{{ if .Config.OgcAPI.Styles -}}
# {{ .Config.Title }} {{ i18n "Styles" }}

{{ i18n "StylesText" }}

{{ $baseUrl := .Config.BaseURL -}}
{{ $projections := .Params.AllProjections -}}
{{ $defaultStyle := .Config.OgcAPI.Styles.Default -}}

## Available Styles

{{ range $style := .Config.OgcAPI.Styles.SupportedStyles -}}
{{ range $srs := $.Params.SupportedProjections -}}
{{ $projection := get $projections (index $srs).Srs -}}
### {{ $style.Title }} ({{ $projection }})

- **Style:** [{{ $baseUrl }}/styles/{{ $style.ID }}__{{ lower $projection }}]({{ $baseUrl }}/styles/{{ $style.ID }}__{{ lower $projection }})
- **Format:** Mapbox style
- **Metadata:** [{{ i18n "StyleMetadata" }}]({{ $baseUrl }}/styles/{{ $style.ID }}__{{ lower $projection }}/metadata)

{{ end -}}
{{ end -}}
{{- end }}

{{/* @formatter:on */}}