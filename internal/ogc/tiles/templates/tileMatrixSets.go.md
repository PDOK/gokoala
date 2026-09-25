{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
# {{ .Config.Title }} {{ i18n "TileMatrixSets" }}

{{ if .Config.OgcAPI.Tiles.DatasetTiles -}}
{{ i18n "TileMatrixSetsDatasetText" }}
{{ else if .Config.OgcAPI.Tiles.Collections -}}
{{ i18n "TileMatrixSetsCollectionText" }}
{{- end }}

| Tile Matrix Set | {{ i18n "DescriptionLabel" }} |
|---|---|
{{ if .Config.OgcAPI.Tiles.HasProjection "EPSG:28992" -}}
| [NetherlandsRDNewQuad](tileMatrixSets/NetherlandsRDNewQuad) | Amersfoort / RD New scheme for the Netherlands |
{{- end }}
{{ if .Config.OgcAPI.Tiles.HasProjection "EPSG:3035" -}}
| [EuropeanETRS89_LAEAQuad](tileMatrixSets/EuropeanETRS89_LAEAQuad) | Lambert Azimuthal Equal Area ETRS89 for Europe |
{{- end }}
{{ if .Config.OgcAPI.Tiles.HasProjection "EPSG:3857" -}}
| [WebMercatorQuad](tileMatrixSets/WebMercatorQuad) | Google Maps Compatible for the World |
{{- end }}

{{/* @formatter:on */}}