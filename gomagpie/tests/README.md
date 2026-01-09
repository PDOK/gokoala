# End-to-end tests

Besides unit- and integration tests (which are stored near the production code) we also use a couple of end-to-end tests.
These tests are also part of the [CI workflow](../.github/workflows/e2e-test.yml).

## Cypress end-to-end tests

The [cypress](./cypress/) directory holds [end-to-end tests](https://docs.cypress.io/guides/core-concepts/testing-types#What-is-E2E-Testing) written
in Cypress targeted at a running Gomagpie instance.

Run `npm run cypress:headless` in CI, run `npm run cypress:open` to author (new) tests.

