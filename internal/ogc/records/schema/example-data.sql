INSERT INTO dcat_catalog (catalog_id, title, description, contact_point, publisher)
VALUES (
    'geoparquet-example',
    '{"en":"GeoParquet example catalog"}',
    '{"en":"A catalog containing a GeoParquet example dataset."}',
    '{"name":"Open Geospatial Consortium"}',
    '{"name":"Open Geospatial Consortium"}'
);

INSERT INTO dcat_dataset (
    id, catalog_id, identifier, access_rights_uri, title, description, contact_point,
    creators, publisher, theme_uris, language_uris, keywords, record_time, stac_properties
)
VALUES (
    'geoparquet-example-dataset',
    'geoparquet-example',
    'https://example.org/datasets/geoparquet-example',
    'http://publications.europa.eu/resource/authority/access-right/PUBLIC',
    'GeoParquet example dataset',
    'An example GeoParquet file published by the Open Geospatial Consortium.',
    '{"name":"Open Geospatial Consortium"}',
    '[{"name":"Open Geospatial Consortium"}]',
    '{"name":"Open Geospatial Consortium"}',
    '["http://publications.europa.eu/resource/authority/data-theme/TECH"]',
    '["http://publications.europa.eu/resource/authority/language/ENG"]',
    '["GeoParquet","example"]',
    '{"timestamp":"2025-01-01T00:00:00Z"}',
    '{"datetime":"2025-01-01T00:00:00Z"}'
);

INSERT INTO dcat_distribution (
    distribution_id, dataset_id, asset_key, access_url, media_type, checksum, license_uri, title, description, stac_fields
)
VALUES (
    'geoparquet-example-file',
    'geoparquet-example-dataset',
    'geoparquet',
    'https://raw.githubusercontent.com/opengeospatial/geoparquet/main/examples/example.parquet',
    'application/vnd.apache.parquet',
    'sha256:example-checksum',
    'https://www.apache.org/licenses/LICENSE-2.0',
    'GeoParquet example.parquet',
    'Example GeoParquet file from the Open Geospatial Consortium.',
    '{"table:primary_geometry":"geometry"}'
);