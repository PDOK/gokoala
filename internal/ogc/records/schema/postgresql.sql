CREATE TABLE dcat_catalog (
    catalog_id TEXT PRIMARY KEY,
    title JSONB NOT NULL,
    description JSONB NOT NULL,
    contact_point JSONB NOT NULL,
    publisher JSONB NOT NULL,
    license_uri TEXT,
    properties JSONB NOT NULL DEFAULT '{}'::jsonb
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
    title_translations JSONB NOT NULL DEFAULT '{}'::jsonb,
    description TEXT NOT NULL,
    description_translations JSONB NOT NULL DEFAULT '{}'::jsonb,
    contact_point JSONB NOT NULL,
    creators JSONB NOT NULL,
    publisher JSONB NOT NULL,
    theme_uris JSONB NOT NULL,
    language_uris JSONB NOT NULL DEFAULT '[]'::jsonb,
    keywords JSONB NOT NULL DEFAULT '[]'::jsonb,
    spatial JSONB,
    temporal JSONB,
    geometry_geojson JSONB,
    bbox JSONB,
    record_time JSONB,
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    stac_properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    stac_version TEXT NOT NULL DEFAULT '1.1.0',
    stac_extensions JSONB NOT NULL DEFAULT '[]'::jsonb
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
    title JSONB,
    description JSONB,
    byte_size BIGINT CHECK (byte_size IS NULL OR byte_size >= 0),
    checksum JSONB,
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    stac_fields JSONB NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE (dataset_id, asset_key),
    CHECK (download_url IS NULL OR media_type IS NOT NULL)
);

CREATE INDEX dcat_dataset_catalog_id_idx ON dcat_dataset(catalog_id);
CREATE INDEX dcat_distribution_dataset_id_idx ON dcat_distribution(dataset_id);