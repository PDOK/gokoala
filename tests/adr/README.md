# API Design Rules (ADR).

Run the [ADR linter](https://developer.overheid.nl/kennisbank/apis/api-design-rules/api-design-rules-linter) against
GoKoala, al from withing Docker Compose with Traefik in front since the ADR linter also checks for TLS usage and URL
paths.

Run:

```bash
docker compose up
```

Checkout the [ruleset](.adr/README.md) for details.