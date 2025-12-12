# Datasource common

This package holds common/shared data structures and logic between the different data sources.

This package should stay data source agnostic, so no imports of `sqlx`, `pgx` are allowed
and no GeoPackage, Postgres, Mongo, Elastic, etc specific logic is allowed.