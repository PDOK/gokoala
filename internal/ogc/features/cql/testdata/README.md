# Testdata

## OGC examples as testdata

The [ogc](./ogc) directory contains CQL2 Text examples from OGC.
Source: [OGC spec on GitHub](https://github.com/opengeospatial/ogcapi-features/tree/64ac2d892b877b711a4570336cb9d42e2afb4ef8/cql2/standard/schema/examples/text)

Note:

- The files with expected results are suffixed with `_expected_<datasource>.txt`. We provide these (not OGC).
- The files with expected errors are suffixed with `_expected_error_<datasource>.txt`. Also provided by us (not OGC).

Where `<datasource>` is one of: `gpkg`, `postgres`, etc.