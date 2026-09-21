
# Search Addresses Features Search

Geocodeerservice om naar <i>features</i> te zoeken in één of meerdere collecties op basis van vrije tekst, zoals een adres of plaatsnaam.

## Search

Use the `search` endpoint to find features by location or text query.

**CRS:** Use the `crs` query parameter to specify the coordinate reference system.

- Default: `http://www.opengis.net/def/crs/OGC/1.3/CRS84`
- `http://www.opengis.net/def/crs/EPSG/0/28992`
- `http://www.opengis.net/def/crs/EPSG/0/3857`
- `http://www.opengis.net/def/crs/EPSG/0/4258`


**Bounding Box:** Use the `bbox` query parameter to spatially filter results.

The search API returns GeoJSON feature collections. Browse results interactively via the HTML interface.
