{{define "feature_forms"}}
{{/* @formatter:off */}}
## {{ i18n "Filters" }}

**CRS:** Use the `crs` query parameter to specify the coordinate reference system.

- Default: `http://www.opengis.net/def/crs/OGC/1.3/CRS84`
{{ range $index, $srs := .Config.OgcAPI.Features.CollectionSRS $.Params.CollectionID -}}
- `http://www.opengis.net/def/crs/EPSG/0/{{ trimPrefix "EPSG:" $srs }}`
{{ end }}

{{ if and .Params.Metadata .Params.Metadata.TemporalProperties -}}
**Datetime:** Use the `datetime` query parameter for temporal filtering.

- Instant: `datetime=2024-01-01`
- Interval: `datetime=2024-01-01/2024-12-31`
- Open start: `datetime=../2024-12-31`
- Open end: `datetime=2024-01-01/..`

{{- end }}

**Limit:** Use the `limit` query parameter (default: 10, max: {{ .Config.OgcAPI.Features.Limit.Max }}).

{{ if .Params.CQLEnabled -}}
**CQL Filter:** Use the `filter` query parameter with CQL2 text expressions:

```
name = 'FooBar'
name = 'FooBar' AND (width > 5 OR height <= 30)
city NOT IN ('Foo', 'Bar')
```

{{- end }}

{{ if .Params.ConfiguredPropertyFilters -}}
**Property Filters:**

{{ range $pfName, $pf := .Params.ConfiguredPropertyFilters -}}
- `{{ $pfName }}`: {{ if and $pf $pf.AllowedValues }}Allowed values: {{ $pf.AllowedValues | join ", " }}{{ else }}free text{{ end }}
{{ end -}}
{{- end }}
{{end}}

{{/* @formatter:on */}}