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

## OGC Compliance validation

In the case of OGC API Features the complaince is [validated on each PR in CI](.github/workflows/e2e-test.yml)
using the OGC [TEAM Engine](https://cite.opengeospatial.org/teamengine/). More specifically using a 
[CLI friendly](https://github.com/PDOK/ets-ogcapi-features10-docker) version of this tool. GoKoala currently
passes all OGC API Features compliance tests.