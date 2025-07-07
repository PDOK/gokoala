# Postgres testdata

Execute `docker compose up` to get a running Postgres/PostGIS instance filed with testdata. The testdata
is extracted from GeoPackages. This way we only have to store the testdata once and can re-use it in all 
datasources supported by GoKoala. There's one Postgres database with a schema for each imported GeoPackage.

Note: the unit tests use this Docker Compose file through [Testcontainers](https://golang.testcontainers.org/).
See [setup_test.go](../../../setup_test.go).

# How to add new testdata?

- Add a GeoPackage to this codebase. Make sure it contains only testdata, no sensitive information.
- Create new schema in `create-schemas.sql`. Choose a valid name (e.g., no '-' chars)
- Add `ogr2ogr` import command to `docker-compose.yaml` and reference the newly added GeoPackage.