# ADR ruleset

## ADR 2.1.0

Ruleset grabbed on 03-11-2025
from [ADR repo](https://github.com/developer-overheid-nl/don-static/tree/1ad13cc5e549ad6f9156f87e3f10e792a124f327/assets/adr/2.1).

Slightly modified:

- added `oas3-valid-schema-example: off`. Not part of ADR but standard Spectral functionality. Not compatible with
  default OGC examples.
- set `use-problem-schema` from `warn` to `error` since we want the linter to break in case RFC 9457 problems aren't
  used.

### Known issues

- https://github.com/developer-overheid-nl/don-static/issues/15

## ADR 2.2.0

Ruleset grabbed on 05-10-2026
from [ADR documentation page](https://gitdocumentatie.logius.nl/publicatie/api/adr/2.2.0/media/linter.yaml)

Slightly modified:

- set `oas3-valid-schema-example` from `error` to `off`. Not compatible with default OGC examples.
- set `nlgov:paths-kebab-case` from `error` to `off`. Not compatible with default OGC API paths.
- set `nlgov:query-keys-camel-case` from `error` to `off`. Not compatible with default OGC API query parameters.