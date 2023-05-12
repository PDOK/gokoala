# Examples

Checkout the examples below to see how GoKoala works.

## Example OGC API Tiles

This example uses vector tiles from the [PDOK BGT dataset](https://www.pdok.nl/introductie/-/article/basisregistratie-grootschalige-topografie-bgt-).

- Start GoKoala as specified in the root [README](../README.md#run) 
  and provide `config_vectortiles.yaml` as the config file.
- Open http://localhost:8080 to explore the landing page
- Call http://localhost:8080/tiles/NetherlandsRDNewQuad/12/2235/2031.pbf to download a specific tile

## Example OGC API 3D GeoVolumes

This example uses 3D tiles of New York.

- Start GoKoala as specified in the root [README](../README.md#run)
  and provide `config_3d.yaml` as the config file.
- Open http://localhost:8080 to explore the landing page
- Call http://localhost:8080/collections/NewYork/3dtiles/6/0/1.b3dm to download a specific 3D tile

## Example multiple OGC APIs for a single collection

This example demonstrates that you can have a collection (NewYork in this case) that offers
multiple OGC APIs (both OGC API 3D GeoVolumes and OGC API Features in this example).

To keep the config DRY we use YAML anchors+aliases to reference common metadata for a collection.

- Start GoKoala as specified in the root [README](../README.md#run)
  and provide `config_multiple_ogc_apis_single_collection.yaml` as the config file.
- Open http://localhost:8080 to explore the landing page
- Call http://localhost:8080/collections/NewYork/ to view the collection

