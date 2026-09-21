{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
# {{ .Config.Title }} (OGC API)

{{ .Config.Abstract }}

{{ if .Config.Keywords -}}
**{{ i18n "Keywords" }}:** {{ .Config.Keywords | join ", " }}
{{- end }}
**{{ i18n "License" }}:** [{{ .Config.License.Name }}]({{ .Config.License.URL }})
{{ if .Config.Support -}}
**{{ i18n "Support" }}:** [{{ .Config.Support.Name }}]({{ .Config.Support.URL }})
{{- end }}
{{ if .Config.MetadataLinks -}}
  {{ range $metadataLink := .Config.MetadataLinks -}}
**{{ i18n "MetadataFor" }} {{ $metadataLink.Category }}:** [{{ i18n "ViewAt" }} {{ $metadataLink.Name }}]({{ $metadataLink.URL }})
  {{ end -}}
{{- end }}

{{- if .Config.DatasetDetails -}}
  {{ range $detailField := .Config.DatasetDetails -}}
**{{ $detailField.Name }}:** {{ $detailField.Value }}
  {{ end -}}
{{- end }}

- [OpenAPI {{ i18n "Specification" }}](api)
- [{{ i18n "Conformance" }}](conformance) indicates which [OGC APIs](https://ogcapi.ogc.org/) and building blocks are implemented by this API.
{{- if .Config.OgcAPI.FeaturesSearch -}}
- [{{ i18n "Search" }}](search): {{ i18n "SearchText" }}
{{- end }}
{{ if .Config.HasCollections -}}
- [{{ i18n "Collections" }}](collections): features/tiles/etc available in this API.
{{- end }}
{{ if and .Config.OgcAPI.Tiles .Config.OgcAPI.Tiles.DatasetTiles -}}
- [{{ i18n "Tiles" }}](tiles): OGC API Tiles supported by this API.
{{- end }}
{{ if .Config.OgcAPI.Styles -}}
- [{{ i18n "Styles" }}](styles): OGC API Styles supported by this API.
{{- end }}
{{ if .Config.OgcAPI.Tiles -}}
- [{{ i18n "TileMatrixSets" }}](tileMatrixSets): OGC Two Dimensional Tile Matrix Sets supported by this APi.
{{- end }}

{{/* @formatter:on */}}