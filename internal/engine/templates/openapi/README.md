# OGC OpenAPI specs

We ship OpenAPI specs for the OGC endpoints that are supported out of the box by
GoKoala. Some endpoints or features aren't supported and therefore removed from
the default OGC OpenAPI files (*). These changes are documented below. **Try to keep
this up-to-date!**

The OpenAPI files/templates in this directory are merged to one spec by GoKoala. In
addition, it's possible to provide GoKoala with a custom OpenAPI spec (using a
CLI flag) and overwrite any defaults or specify additional endpoints.

(*) This is also the intent of the OGC: _"An implementation should only include the paths
that are implemented and remove the references to the rest."_ source: OGC API Tiles 1.0 spec.

## Changes

### OGC Common Core (Part 1)

`common.go.json` is based on
[common-1.0](https://developer.ogc.org/api/common/openapi.yaml)

- Changes:
  - Removed default contact details

### OGC Common Core (Part 2)

`common-collections.go.json` is based on
[common-part-2-draft](https://developer.ogc.org/api/common/openapi2.yaml)

- Changes:
  - Removal of OGC Common Part 1 endpoints (landing page, api, conformance), already
    covered by `common.json`
  - Removal of unreferenced responses like "Created", "Updated", etc.
  - Removal of link type regex pattern (_reason: not parsable using current
    OpenAPI Go lib_)
  - Removal of `crs` enum restriction
  - Change values for `f` param from `application/json` to just `json`, same for HTML.
  - Removed default contact details
  - Removed unused responses

### OGC Tiles

`tiles.go.json` is based on
[ogcapi-tiles-1](https://schemas.opengis.net/ogcapi/tiles/part1/1.0/openapi/ogcapi-tiles-1.bundled.json)

- Changes:
  - Removal of OGC Common endpoints (landing page, api, conformance), already
    covered by `common.json`
  - Removal of generic OGC Collection types, already covered in `commons-collections.json`
    - Also removal of OGC Collection endpoints (we don't support these for Tiles at the
    moment), this may change in the future.
    - Removed `extent`, already covered in `commons-collections.json`
  - Removal of OGC Style endpoint (/styles), already - and better - covered by `styles.json`
  - Removal of GeoJSON as tiles format, only MapBox Vector Tiles are supported.
  - Removal of optional parameters for `/tiles` endpoint like datetime (temporal data)
    and crs (on-the-fly re-projection)
  - Changed TileMatrixSet enum values to  "NetherlandsRDNewQuad",
    "EuropeanETRS89_GRS80Quad_Draft", "WebMercatorQuad"
  - Changed `tags` from "server" to "common".
  - Support HTML responses for `/tileMatrixSets/{tileMatrixSetId}` calls
  - Added TileJSON support to `/tiles/{tileMatrixSetId}`. This is allowed in the OGC Tiles spec since it mentions
    "Support for alternative encodings for tileset metadata can be added, such as TileJSON."
  - Remove superfluous `/api/tileMatrixSets`, since it does the same as `/tileMatrixSets`
  - Replaced "EuropeanETRS89_GRS80Quad_Draft" with "EuropeanETRS89_LAEAQuad"
  - Removed unused `datetime` parameter
  - Removed default contact details

### OGC Features

`features.go.json` is based on
[ogcapi-features-1.0.1](https://app.swaggerhub.com/apis/OGC/ogcapi-features-1-example-1/1.0.1) and
[ogcapi-features-2](https://schemas.opengis.net/ogcapi/features/part2/1.0/openapi/ogcapi-features-2.yaml)

- Changes:
  - Removal of OGC Common endpoints (landing page, api, conformance), already
    covered by `common.json`
  - Removal of generic OGC Collection endpoints, already covered in `commons-collections.json`
  - Changed tags from "Data" to "Features"
  - Removed default contact details
  - Removed numberMatched, is optional in the spec and not compatible with cursor-based pagination.
  - Changed examples
    - to use `?f=format` instead of `.format` to be more inline with the OGC spec/docs
    - removed `offset` since we (will) use `cursor` for pagination
  - Added `cursor` query param to feature collection request. Vendor specific parameters need to be explicitly listed
    (the same holds true if you would use offset-based pagination, you would need to list the `offset` param).
  - Added parameters/components/headers from part 2 schema (automatic merge wasn't possible).
  - For `datetime` query param only single datetimes are supported, not intervals.
  - Removed unused responses

- Added:
   - JSON-FG support for conformance class "core", based on a subset of JSON schemas from [JSON-FG release 0.2](https://github.com/opengeospatial/ogc-feat-geo-json/releases/tag/v0.2.1)

### OGC 3D GeoVolumes

`3dgeovolumes.go.json` is based on
[ogcapi-3d-geovolumes-draft-0.0.2](https://raw.githubusercontent.com/opengeospatial/ogcapi-3d-geovolumes/main/standard/openapi/ogcapi-3d-geovolumes-draft-0.0.2.yaml)
and [cologne_lod2](https://demo.ldproxy.net/cologne_lod2/api/?f=json)

- Changes:
  - Removal of OGC Common endpoints (landing page, api, conformance), already
    covered by `common.json`
  - Removed most endpoints only included 3d tiles specific endpoints
  - Removed all unused components
  - Removed default contact details

### OGC Styles

`styles.go.json` is based on
[ogcapi-styles-1](https://developer.ogc.org/api/styles/openapi.yaml)

- Changes:
  - Removal of OGC Common endpoints (landing page, api, conformance), already
    covered by `common.json`
  - Removal of mutating endpoints (POST/PUT/DELETE/etc). We only support
    read-only endpoints
    - Removal of securitySchemes
    - Tidy up `tags`
  - Support HTML responses for `/styles` and `/styles/{styleId}/metadata` calls
  - Add `style-set` and `style-set-entry` schemas from [style-set](https://api.swaggerhub.com/domains/cportele/ogcapi-draft-extensions/1.0.0#/components/schemas/style-set)
  - Removed default contact details
  - Removed resources endpoints `/resources` (and sub endpoints). We do support the resources endpoint but we don't support/allow listing all available resources. There's also [discussion](https://github.com/opengeospatial/ogcapi-styles/issues/12) about the merit of this endpoint
