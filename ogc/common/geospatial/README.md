# Geospatial data resources / collections.

For GoKoala devs: If you want to implement collections support in one of the OGC building blocks 
in GoKoala (see `ogc` package) you'll need to perform the following tasks:

Config:
- Expand / add yaml tag in `engine.Config.OgcAPI` to allow users to configure collections

OpenAPI
- Materialize the collections as API endpoints by looping over the collection in the OpenAPI template 
  for that specific OGC building block. For example for OGC tiles you'll need to 
  create `/collection/{collectionId}/tiles` endpoints in OpenAPI. Note `/collection/{collectionId}` endpoint
  are already implemented in OpenAPI by this package.

Responses:
- Expand the `collections` and `collection` [templates](./templates). 
- Implement collection support in the given OGC building block code by implementing the `CollectionContent` interface.

Testing:
- Add unit tests

