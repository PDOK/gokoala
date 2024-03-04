# End-to-end tests

Besides unit- and integration tests (which are stored near the production code) we also use various end-to-end tests:

## Cypress end-to-end tests

The [cypress](./cypress/) directory holds [end-to-end tests](https://docs.cypress.io/guides/core-concepts/testing-types#What-is-E2E-Testing) written
in Cypress to a running GoKoala instance.

> NOTE: The [viewer](../viewer/cypress) also contains Cypress tests, these are only focussed on viewer/map components.

## Cloud-backed GeoPackage test

See [OGC API Features example](../examples).

## OGC Compliance validation

Will be automated in the future, currently [executed manually](../README.md).