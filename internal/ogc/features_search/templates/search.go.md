{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
# {{ .Config.Title }} Features Search

{{ i18n "SearchText" }}

## Search

Use the `search` endpoint to find features by location or text query.

**CRS:** Use the `crs` query parameter to specify the coordinate reference system.

- Default: `http://www.opengis.net/def/crs/OGC/1.3/CRS84`
{{ if .Config.OgcAPI.FeaturesSearch -}}
{{ range $index, $srs := .Config.OgcAPI.FeaturesSearch.CollectionsSRS -}}
- `http://www.opengis.net/def/crs/EPSG/0/{{ trimPrefix "EPSG:" $srs }}`
{{ end -}}
{{- end }}

**Bounding Box:** Use the `bbox` query parameter to spatially filter results.

The search API returns GeoJSON feature collections. Browse results interactively via the HTML interface.
{{/* @formatter:on */}}