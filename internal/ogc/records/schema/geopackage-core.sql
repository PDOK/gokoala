PRAGMA application_id = 1196444487;
PRAGMA user_version = 10300;
PRAGMA foreign_keys = ON;

CREATE TABLE gpkg_spatial_ref_sys (
    srs_name TEXT NOT NULL,
    srs_id INTEGER NOT NULL PRIMARY KEY,
    organization TEXT NOT NULL,
    organization_coordsys_id INTEGER NOT NULL,
    definition TEXT NOT NULL,
    description TEXT
);

INSERT INTO gpkg_spatial_ref_sys (srs_name, srs_id, organization, organization_coordsys_id, definition, description)
VALUES
    ('Undefined Cartesian SRS', -1, 'NONE', -1, 'undefined', 'undefined Cartesian coordinate reference system'),
    ('Undefined Geographic SRS', 0, 'NONE', 0, 'undefined', 'undefined geographic coordinate reference system'),
    ('WGS 84 geodetic', 4326, 'EPSG', 4326, 'GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563]],PRIMEM["Greenwich",0],UNIT["degree",0.0174532925199433]]', 'longitude/latitude coordinates in decimal degrees on the WGS 84 spheroid');

CREATE TABLE gpkg_contents (
    table_name TEXT NOT NULL PRIMARY KEY,
    data_type TEXT NOT NULL CHECK (data_type IN ('features', 'tiles', 'attributes', '2d-gridded-coverage')),
    identifier TEXT UNIQUE,
    description TEXT DEFAULT '',
    last_change DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    min_x DOUBLE,
    min_y DOUBLE,
    max_x DOUBLE,
    max_y DOUBLE,
    srs_id INTEGER,
    CONSTRAINT fk_gc_r_srs_id FOREIGN KEY (srs_id) REFERENCES gpkg_spatial_ref_sys(srs_id)
);

INSERT INTO gpkg_contents (table_name, data_type, identifier, description, last_change, srs_id)
VALUES
    ('dcat_catalog', 'attributes', 'dcat_catalog', 'DCAT-AP-NL catalogs', '2025-01-01T00:00:00.000Z', NULL),
    ('dcat_dataset', 'attributes', 'dcat_dataset', 'DCAT-AP-NL datasets', '2025-01-01T00:00:00.000Z', NULL),
    ('dcat_distribution', 'attributes', 'dcat_distribution', 'DCAT-AP-NL distributions', '2025-01-01T00:00:00.000Z', NULL);