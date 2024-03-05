# End-to-end tests

Besides unit- and integration tests (which are stored near the production code) we also use a couple of end-to-end tests.
These tests are also part of the CI workflow.

## Cypress end-to-end tests

The [cypress](./cypress/) directory holds [end-to-end tests](https://docs.cypress.io/guides/core-concepts/testing-types#What-is-E2E-Testing) written
in Cypress targeted at a running GoKoala instance.

> NOTE: The [viewer](../viewer/cypress) also contains Cypress tests, these are only focussed on viewer/map components.

Run `npm run cypress:headless` in CI, run `npm run cypress:open` to author (new) tests.

## Cloud-backed GeoPackage smoke test

See [OGC API Features example](../examples) involving the `config_features_azure.yaml` config file.

## OGC Compliance validation

Will be automated in the future, currently [executed manually](../README.md).