# Records database schema

The Records API reads catalog, Dataset, and Distribution rows from a PostgreSQL database or a local GeoPackage. The SQL definitions are in `postgresql.sql` and `geopackage.sql`.

The tables model the core DCAT-AP-NL 3.0 classes:

- `dcat_catalog` stores an OGC Records catalog. Its `catalog_id` is exposed as the OGC collection ID.
- `dcat_dataset` stores one DCAT Dataset and is returned as one OGC API - Records record / STAC Item. `id` is the local API record ID; `identifier` preserves the dataset's DCAT identifier.
- `dcat_distribution` stores one DCAT Distribution and is returned as one STAC Asset of its Dataset. A Dataset may have multiple distributions.

The schema makes DCAT-AP-NL required Dataset fields non-null: `identifier`, `access_rights_uri`, `title`, `description`, `contact_point`, `creators`, `publisher`, and `theme_uris`. Distribution `access_url` and `license_uri` are required. Catalog title, description, contact point, and publisher are required. JSON columns hold structured DCAT support classes such as `vcard:Kind`, `foaf:Agent`, vocabularies, and additional properties.

`title` and `description` contain the primary language text used by OGC Records. `title_translations` and `description_translations` preserve additional language literals. `record_time` uses the OGC API - Records `time` object. `stac_properties` stores STAC Item properties such as `datetime`, `start_datetime`, and `end_datetime`; the Item `datetime` is required by STAC. `stac_version` defaults to `1.1.0`, and `stac_extensions` may list explicitly used extension schema URIs.

Each Distribution's `access_url` or `download_url` becomes the asset `href`; `media_type` becomes the STAC asset `type`; `byte_size` becomes `file:size`; and `stac_fields` carries extension fields scoped to that asset. For example, GeoParquet may use `table:primary_geometry`; raster COGs may use `raster:*` and `proj:*` fields. GeoPackage assets use the same STAC Asset Object and File Info fields. Extension URIs are emitted from recognized `table`, `file`, `raster`, and `proj` prefixes.

Use the appropriate format URI/media type and license for each distribution. CQL filtering is configured per Records collection with the existing GoKoala `filters.properties` and `filters.cql` settings; queryable names must correspond to columns in `dcat_dataset`.

The SQL files implement the relational storage shape. DCAT-AP-NL controlled-vocabulary URI validity and the complete SHACL profile are not enforced by SQL constraints; validate imported metadata against the official [DCAT-AP-NL 3.0 SHACL shapes](https://geonovum.github.io/DCAT-AP-NL30/shapes/) as part of data import.