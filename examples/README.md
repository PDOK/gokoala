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

