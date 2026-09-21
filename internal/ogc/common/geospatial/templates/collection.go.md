{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
# {{ .Config.Title }} {{ if and .Params.Collection.Metadata .Params.Collection.Metadata.Title }}{{ .Params.Collection.Metadata.Title }}{{ else }}{{ .Params.Collection.ID }}{{ end }}

{{ if and .Params.Collection.Metadata .Params.Collection.Metadata.Description -}}
    {{ .Params.Collection.Metadata.Description }}
{{- end }}
{{ if and .Params.Collection.Metadata .Params.Collection.Metadata.Keywords -}}
**{{ i18n "Keywords" }}:** {{ .Params.Collection.Metadata.Keywords | join ", " }}
{{- end }}
{{ if and .Params.Collection.Metadata .Params.Collection.Metadata.LastUpdated -}}
**{{ i18n "LastUpdated" }}:** {{ toDate "2006-01-02T15:04:05Z07:00" .Params.Collection.Metadata.LastUpdated | date "2006-01-02" }}
{{- end }}
{{ if and .Params.Collection.Metadata .Params.Collection.Metadata.Extent (ne $.Params.GeomType "none") -}}
**{{ i18n "GeographicExtent" }}{{ if .Params.Collection.Metadata.Extent.Srs }} ({{ .Params.Collection.Metadata.Extent.Srs }}){{ else }} ([CRS84](http://www.opengis.net/def/crs/OGC/1.3/CRS84)){{ end }}:** {{ .Params.Collection.Metadata.Extent.Bbox | join ", " }}
{{- end }}
{{ if and .Params.Collection.Metadata .Params.Collection.Metadata.Extent .Params.Collection.Metadata.Extent.Interval -}}
**{{ i18n "TemporalExtent" }} ([ISO-8601](http://www.opengis.net/def/uom/ISO-8601/0/Gregorian)):** {{ toDate "2006-01-02T15:04:05Z" ((first .Params.Collection.Metadata.Extent.Interval) | replace "\"" "") | date "2006-01-02" }} / {{ if not (contains "null" (last .Params.Collection.Metadata.Extent.Interval)) }}{{ toDate "2006-01-02T15:04:05Z" ((last .Params.Collection.Metadata.Extent.Interval) | replace "\"" "") | date "2006-01-02" }}{{ else }}..{{ end }}
{{- end }}

## Available Resources

{{ if and .Config.OgcAPI.GeoVolumes .Config.OgcAPI.GeoVolumes.Collections -}}
{{ if .Config.OgcAPI.GeoVolumes.Collections.ContainsID .Params.Collection.ID -}}
### 3D GeoVolumes
{{ if hasfield .Params.Collection "IsDtm" -}}
    {{ if not .Params.Collection.IsDtm -}}
- [3D Tiles]({{ .Config.BaseURL }}/collections/{{ .Params.Collection.ID }}/3dtiles)
    {{- end }}
    {{ if .Params.Collection.IsDtm -}}
- [Quantized Mesh DTM]({{ .Config.BaseURL }}/collections/{{ .Params.Collection.ID }}/quantized-mesh)
    {{- end }}
{{- end }}
{{ if hasfield .Params.Collection "URL3DViewer" -}}
    {{ if .Params.Collection.URL3DViewer -}}
- [3D Viewer]({{ .Params.Collection.URL3DViewer }})
   {{- end }}
{{- end }}
{{- end }}
{{- end }}

{{ if and .Config.OgcAPI.Tiles .Config.OgcAPI.Tiles.Collections -}}
{{ if .Config.OgcAPI.Tiles.Collections.ContainsID .Params.Collection.ID -}}
### Tiles
[View Tiles]({{ .Config.BaseURL }}/collections/{{ .Params.Collection.ID }}/tiles)
{{- end }}
{{- end }}

{{ if and .Config.OgcAPI.Features .Config.OgcAPI.Features.Collections -}}
{{ if .Config.OgcAPI.Features.Collections.ContainsID .Params.Collection.ID -}}
{{ if ne $.Params.GeomType "none" -}}
### Features
[Browse Features]({{ .Config.BaseURL }}/collections/{{ .Params.Collection.ID }}/items) 
[Schema]({{ .Config.BaseURL }}/collections/{{ .Params.Collection.ID }}/schema)

Available formats:
{{ range $availableFormat := $.Params.Type.AvailableFormats -}}
- [{{ $availableFormat.Name }}]({{ $.Config.BaseURL }}/collections/{{ $.Params.Collection.ID }}/items?f={{ $availableFormat.Key }})
{{ end -}}
{{- end }}

{{ if $.Config.OgcAPI.Features.SupportsNonGeoData -}}
{{ if eq $.Params.GeomType "none" -}}
### Attributes
[Browse Attributes]({{ .Config.BaseURL }}/collections/{{ .Params.Collection.ID }}/items) 
[Schema]({{ .Config.BaseURL }}/collections/{{ .Params.Collection.ID }}/schema)
{{- end }}
{{- end }}

{{- end }}
{{- end }}

{{ if and .Params.Collection.Links .Params.Collection.Links.Downloads -}}
### Downloads
{{ range $link := .Params.Collection.Links.Downloads -}}
- [{{ $link.Name }}{{ if $link.Size }} ({{ $link.Size }}){{ end }}]({{ $link.AssetURL }})
{{ end -}}
{{- end }}

{{/* @formatter:on */}}