{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
# NetherlandsRDNewQuad

{{ i18n "NetherlandsRDNewQuadAbstract" }}

[TileJSON]({{ .Params.BaseURL }}/tiles/NetherlandsRDNewQuad?f=tilejson).
**URL template:** `{{ .Params.BaseURL }}/tiles/NetherlandsRDNewQuad/{z}/{y}/{x}?f=mvt`

## Tile matrix set limits

{{ i18n "AvailableZoomLevels" }}:

| {{ i18n "ZoomLevel" }} {z} | {{ i18n "MinimumValue" }} {y} | {{ i18n "MaximumValue" }} {y} | {{ i18n "MinimumValue" }} {x} | {{ i18n "MaximumValue" }} {x} |
|---|---|---|---|---|
{{- range $type := .Params.SupportedSrs }}
{{- if eq $type.Srs "EPSG:28992" }}
{{- if eq $type.ZoomLevelRange.Start 0 }}
| 0 | 0 | 0 | 0 | 0 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 1) (ge $type.ZoomLevelRange.End 1) }}
| 1 | 0 | 1 | 0 | 1 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 2) (ge $type.ZoomLevelRange.End 2) }}
| 2 | 0 | 3 | 0 | 3 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 3) (ge $type.ZoomLevelRange.End 3) }}
| 3 | 0 | 7 | 0 | 7 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 4) (ge $type.ZoomLevelRange.End 4) }}
| 4 | 0 | 15 | 0 | 15 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 5) (ge $type.ZoomLevelRange.End 5) }}
| 5 | 0 | 31 | 0 | 31 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 6) (ge $type.ZoomLevelRange.End 6) }}
| 6 | 0 | 63 | 0 | 63 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 7) (ge $type.ZoomLevelRange.End 7) }}
| 7 | 0 | 127 | 0 | 127 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 8) (ge $type.ZoomLevelRange.End 8) }}
| 8 | 0 | 255 | 0 | 255 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 9) (ge $type.ZoomLevelRange.End 9) }}
| 9 | 0 | 511 | 0 | 511 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 10) (ge $type.ZoomLevelRange.End 10) }}
| 10 | 0 | 1023 | 0 | 1023 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 11) (ge $type.ZoomLevelRange.End 11) }}
| 11 | 0 | 2047 | 0 | 2047 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 12) (ge $type.ZoomLevelRange.End 12) }}
| 12 | 0 | 4095 | 0 | 4095 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 13) (ge $type.ZoomLevelRange.End 13) }}
| 13 | 0 | 8191 | 0 | 8191 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 14) (ge $type.ZoomLevelRange.End 14) }}
| 14 | 0 | 16383 | 0 | 16383 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 15) (ge $type.ZoomLevelRange.End 15) }}
| 15 | 0 | 32767 | 0 | 32767 |
{{- end }}
{{- if and (le $type.ZoomLevelRange.Start 16) (ge $type.ZoomLevelRange.End 16) }}
| 16 | 0 | 65535 | 0 | 65535 |
{{- end }}
{{- end }}
{{- end }}
