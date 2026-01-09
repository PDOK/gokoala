[fake-addresses-crs84.gpkg](fake-addresses-crs84.gpkg) is derived from `etl/testdata/addresses-crs84.gpkg`
but heavily modified to support all sorts of test cases. Especially the street/place fields are modified. 
The geom field does not necessarily reflect the actual place/street. Just consider this data to be
fake/fabricated for the sake of testing. Feel free to add additional records to implement new test cases!

Note the `fake-addresses-crs84.gpkg` geopackage is also used in the examples.
See [docker-compose-feature-search.yaml](../../../../examples/docker-compose-features-search.yaml).