# OGC OpenAPI specs

We ship OpenAPI specs for the OGC endpoints that are supported out of the box by GoKoala. We strive to fully conform to 
the OGC specs but some endpoints or features aren't supported and therefore removed from the default OGC OpenAPI files. 
This is also the intent of the OGC: _"An implementation should only include the paths that are implemented and remove 
the references to the rest."_ source: `OGC API Tiles 1.0 spec`.

The OpenAPI files/templates in this directory are merged into one spec by GoKoala. In addition, it's possible to provide 
GoKoala with a custom OpenAPI spec (using a CLI flag) and overwrite any defaults or specify additional endpoints.

## Sources

While the OpenAPI specs/templates are modified to match the capabilities of GoKoala, it might be useful to now their origins:

- OGC Common Core (Part 1): `common.go.json` is based on [common-1.0](https://developer.ogc.org/api/common/openapi.yaml)
- OGC Common Core (Part 2): `common-collections.go.json` is based on [common-part-2-draft](https://developer.ogc.org/api/common/openapi2.yaml)
- OGC Tiles: `tiles.go.json` is based on [ogcapi-tiles-1](https://schemas.opengis.net/ogcapi/tiles/part1/1.0/openapi/ogcapi-tiles-1.bundled.json)
- OGC Features: `features.go.json` is based on [ogcapi-features-1.0.1](https://app.swaggerhub.com/apis/OGC/ogcapi-features-1-example-1/1.0.1) and [ogcapi-features-2](https://schemas.opengis.net/ogcapi/features/part2/1.0/openapi/ogcapi-features-2.yaml). Part 5 was added manually based on HTML spec.
- OGC 3D GeoVolumes: `3dgeovolumes.go.json` is based on [ogcapi-3d-geovolumes-draft-0.0.2](https://raw.githubusercontent.com/opengeospatial/ogcapi-3d-geovolumes/main/standard/openapi/ogcapi-3d-geovolumes-draft-0.0.2.yaml) and [cologne_lod2](https://demo.ldproxy.net/cologne_lod2/api/?f=json)
- OGC Styles: `styles.go.json` is based on [ogcapi-styles-1](https://developer.ogc.org/api/styles/openapi.yaml)

Note: See the Git history of this file for more details. We stopped documenting every change to the specs 
since there were just too many and the benefit was lost.