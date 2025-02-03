# End-to-end tests

Besides unit- and integration tests (which are stored near the production code) we also use a couple of end-to-end tests.
These tests are also part of the [CI workflow](../.github/workflows/e2e-test.yml).

## Cypress end-to-end tests

The [cypress](./cypress/) directory holds [end-to-end tests](https://docs.cypress.io/guides/core-concepts/testing-types#What-is-E2E-Testing) written
in Cypress targeted at a running GoKoala instance.

> NOTE: The [viewer](../viewer/cypress) also contains Cypress tests, these are only focussed on viewer/map components.

Run `npm run cypress:headless` in CI, run `npm run cypress:open` to author (new) tests.

## Cloud-backed GeoPackage smoke test

See [OGC API Features example](../examples) involving the `config_features_azure.yaml` config file.

## OGC Compliance Validation

GoKoala currently passes all OGC API Features and OGC API Tiles compliance tests.

These are [validated on each PR in CI](.github/workflows/e2e-test.yml) using the OGC [TEAM Engine](https://github.com/opengeospatial/teamengine). 
More specifically using a CLI friendly version of this tool:

- https://github.com/PDOK/ets-ogcapi-features10-docker
- https://github.com/PDOK/ets-ogcapi-tiles10-docker

