{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{/* @formatter:off */}}
# {{ .Config.Title }} Features Search / Geocoding

{{ i18n "SearchText" }}

## Search endpoint

The [search]({{ .Config.BaseURL }}/search) endpoint enables text-based geographic searching (geocoding) across multiple OGC API Features collections.
It returns GeoJSON feature collections containing the minimal information needed to display search results to the end user.
To access the complete details of a specific feature, follow the href link provided within its search result.
See the [OpenAPI spec]({{ .Config.BaseURL }}/api?=json) for more details.

The [HTML interface]({{ .Config.BaseURL }}/search?f=html) provides a practical example of how to interact with this API.

### Required Parameters

- The `q` param is required to specify the search term. Wildcards or boolean operators aren't allowed.
- One or more collections using deep object query parameters (e.g., `places[version]=1&addresses[version]=1`) need to be specified.
  A generic query like `?q=Amsterdam` will return **zero** results unless one or more collections are included.

### Optional Parameters

**CRS:** Use the `crs` query parameter to specify the coordinate reference system.

- Default: `http://www.opengis.net/def/crs/OGC/1.3/CRS84`
{{ if .Config.OgcAPI.FeaturesSearch -}}
{{ range $index, $srs := .Config.OgcAPI.FeaturesSearch.CollectionsSRS -}}
- `http://www.opengis.net/def/crs/EPSG/0/{{ trimPrefix "EPSG:" $srs }}`
{{ end -}}
{{- end }}

**Bounding Box:** Use the `bbox` query parameter to spatially filter results.
Use the `bbox-crs` param to specify the CRS of the provided `bbox` coordinates.
{{/* @formatter:on */}}