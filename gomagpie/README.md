<p align="center">
<img src="docs/gomagpie.jpeg" alt="Gomagpie" title="Gomagpie" width="300" />
</p>

# Gomagpie

_Location search and geocoding API. To be used in concert with OGC APIs._

[![Build](https://github.com/PDOK/gomagpie/actions/workflows/build-and-publish-image.yml/badge.svg)](https://github.com/PDOK/gomagpie/actions/workflows/build-and-publish-image.yml)
[![Lint (go)](https://github.com/PDOK/gomagpie/actions/workflows/lint-go.yml/badge.svg)](https://github.com/PDOK/gomagpie/actions/workflows/lint-go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/PDOK/gomagpie)](https://goreportcard.com/report/github.com/PDOK/gomagpie)
[![Coverage (go)](https://github.com/PDOK/gomagpie/wiki/coverage.svg)](https://raw.githack.com/wiki/PDOK/gomagpie/coverage.html)
[![GitHub license](https://img.shields.io/github/license/PDOK/gomagpie)](https://github.com/PDOK/gomagpie/blob/master/LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/pdok/gomagpie.svg)](https://hub.docker.com/r/pdok/gomagpie)

## Description

This application offers location search and geocoding.

## Run

```bash
NAME:
   gomagpie - Run location search and geocoding API, or use as CLI to support the ETL process for this API.

USAGE:
   gomagpie [global options] command [command options]

COMMANDS:
   start-service  Start service to serve location API
   help, h        Shows a list of commands or help for one command
   etl:
     create-search-index  Create empty search index in database
     import-file          Import file into search index

GLOBAL OPTIONS:
   --help, -h  show help
```

Example (config-file is mandatory):

```docker
docker run -v `pwd`/examples:/examples -p 8080:8080 -it pdok/gomagpie start-service \
  --config-file /examples/config.yaml \
  --rewrites-file internal/search/testdata/rewrites.csv \
  --synonyms-file internal/search/testdata/synonyms.csv
```

Now open <http://localhost:8080>or open <http://localhost:8080/api> to check the openAPI specification.

See [examples](examples) for more details.

### Run ETL

Create database using the ETL commands.

```shell
# First create database to populate
docker run --rm --name postgis -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=test_db -p 5432:5432 postgis/postgis

./gomagpie create-search-index --db-name test_db
./gomagpie import-file --db-name test_db \
  --file internal/etl/testdata/addresses-rd.gpkg \
  --feature-table "addresses" \
  --config-file internal/etl/testdata/config.yaml \
  --collection-id "addresses"
```

## Build

```bash
docker build -t pdok/gomagpie .
```

### Observability

#### Health checks

Health endpoint is available on `/health`.

#### Profiling

Besides the main server, Gomagpie can also start a debug server. This server
binds to localhost and a different port which you must specify using the
`--debug-port` flag. You shouldn't expose this port publicly but only access it
through a tunnel/port-forward. The debug server exposes `/debug` for use by
[pprof](https://go.dev/blog/pprof). For example with `--debug-port 9001`:

- Create a tunnel to the debug server e.g. in k8s: `kubectl port-forward
gomagpie-75f59d57f4-4nd6q 9001:9001`
- Create CPU profile: `go tool pprof
http://localhost:9001/debug/pprof/profile?seconds=20`
- Start pprof visualization `go tool pprof -http=":8000" pprofbin <path to pb.gz
file>`
- Open <http://localhost:8000> to explore CPU flamegraphs and such.

A similar flow can be used to profile memory issues.

## Develop

### Build/run as Go application

```
go build -o gomagpie cmd/main.go
./gomagpie
```

To troubleshoot, review the [Dockerfile](./Dockerfile) since compilation also happens there.

### Linting

Install [golangci-lint](https://golangci-lint.run/usage/install/) and run `golangci-lint run`
from the root.

### Unit test

Either run `go test ./...` or `go test -short ./...`

### IntelliJ / GoLand

- Install the [Go Template](https://plugins.jetbrains.com/plugin/10581-go-template) plugin
- Open `Preferences` > `Editor` > `File Types` select `Go Template files` and
  add the following file patterns:
  - `"*.go.html"`
  - `"*.go.json"`
  - `"*.go.tilejson"`
  - `"*.go.xml"`
- Now add template language support by running the
  [setup-jetbrains-gotemplates.sh](hack/setup-jetbrains-gotemplates.sh) script.
- Reopen the project (or restart IDE). Now you'll have full IDE support in the gomagpie templates.

Also:

- Set import order in `Preferences` > `Editor` > `Code Style` > `Go` > `Imports`
  to `goimports` to align with VSCode and goimports usage in golangci-lint.

### VSCode

- Install the [Go Template](https://marketplace.visualstudio.com/items?itemName=jinliming2.vscode-go-template)
  extension
- Open Extension Settings and add the following file patterns:
  - `"*.go.html"`
  - `"*.go.json"`
  - `"*.go.tilejson"`
  - `"*.go.xml"`
- Also add `html`, `json` and `xml` to the list of Go template languages.
- Now you'll have IDE support in the gomagpie templates.

## Misc

### Origin

Gomagpie started as a fork of [GoKoala](https://github.com/PDOK/gokoala).

### How to Contribute

Make a pull request...

### Contact

Contacting the maintainers can be done through the issue tracker.
