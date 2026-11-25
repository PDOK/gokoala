CREATE TABLE dcat_catalog (
    catalog_id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    contact_point TEXT NOT NULL,
    publisher TEXT NOT NULL,
    license_uri TEXT,
    properties TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE dcat_dataset (
    id TEXT PRIMARY KEY,
    catalog_id TEXT NOT NULL REFERENCES dcat_catalog(catalog_id),
    identifier TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL DEFAULT 'dataset',
    access_rights_uri TEXT NOT NULL CHECK (access_rights_uri IN (
        'http://publications.europa.eu/resource/authority/access-right/PUBLIC',
        'http://publications.europa.eu/resource/authority/access-right/RESTRICTED',
        'http://publications.europa.eu/resource/authority/access-right/NON_PUBLIC'
    )),
    title TEXT NOT NULL,
    title_translations TEXT NOT NULL DEFAULT '{}',
    description TEXT NOT NULL,
    description_translations TEXT NOT NULL DEFAULT '{}',
    contact_point TEXT NOT NULL,
    creators TEXT NOT NULL,
    publisher TEXT NOT NULL,
    theme_uris TEXT NOT NULL,
    language_uris TEXT NOT NULL DEFAULT '[]',
    keywords TEXT NOT NULL DEFAULT '[]',
    spatial TEXT,
    temporal TEXT,
    geometry_geojson TEXT,
    bbox TEXT,
    record_time TEXT,
    properties TEXT NOT NULL DEFAULT '{}',
    stac_properties TEXT NOT NULL DEFAULT '{}',
    stac_version TEXT NOT NULL DEFAULT '1.1.0',
    stac_extensions TEXT NOT NULL DEFAULT '[]'
);

CREATE TABLE dcat_distribution (
    distribution_id TEXT PRIMARY KEY,
    dataset_id TEXT NOT NULL REFERENCES dcat_dataset(id) ON DELETE CASCADE,
    asset_key TEXT NOT NULL,
    access_url TEXT NOT NULL,
    download_url TEXT,
    media_type TEXT,
    format_uri TEXT,
    license_uri TEXT NOT NULL,
    title TEXT,
    description TEXT,
    byte_size INTEGER CHECK (byte_size IS NULL OR byte_size >= 0),
    checksum TEXT,
    properties TEXT NOT NULL DEFAULT '{}',
    stac_fields TEXT NOT NULL DEFAULT '{}',
    UNIQUE (dataset_id, asset_key),
    CHECK (download_url IS NULL OR media_type IS NOT NULL)
);

CREATE INDEX dcat_dataset_catalog_id_idx ON dcat_dataset(catalog_id);
CREATE INDEX dcat_distribution_dataset_id_idx ON dcat_distribution(dataset_id);

INSERT OR IGNORE INTO gpkg_contents (table_name, data_type, identifier, description, last_change, srs_id)
VALUES
    ('dcat_catalog', 'attributes', 'dcat_catalog', 'DCAT-AP-NL catalogs', strftime('%Y-%m-%dT%H:%M:%fZ', 'now'), NULL),
    ('dcat_dataset', 'attributes', 'dcat_dataset', 'DCAT-AP-NL datasets', strftime('%Y-%m-%dT%H:%M:%fZ', 'now'), NULL),
    ('dcat_distribution', 'attributes', 'dcat_distribution', 'DCAT-AP-NL distributions', strftime('%Y-%m-%dT%H:%M:%fZ', 'now'), NULL);