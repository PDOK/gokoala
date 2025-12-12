# ADR ruleset

Rulset grabbed on on 03-11-2025
from [ADR repo](https://github.com/developer-overheid-nl/don-static/tree/1ad13cc5e549ad6f9156f87e3f10e792a124f327/assets/adr/2.1).

Slightly modified:

- added `oas3-valid-schema-example: off`. Not part of ADR but standard Spectral functionality. Not compatible with
  default OGC examples.
- set `use-problem-schema` from `warn` to `error` since we want the linter to break in case RFC 9457 problems aren't
  used.

## Known issues

- https://github.com/developer-overheid-nl/don-static/issues/15