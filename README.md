<p align="center">
<img src="docs/gomagpie.jpeg" alt="Gomagpie" title="Gomagpie" width="300" />
</p>

# Gomagpie

_Location search and geocoding API. To be used in concert with OGC APIs.

[![Build](https://github.com/PDOK/gokoala/actions/workflows/build-and-publish-image.yml/badge.svg)](https://github.com/PDOK/gokoala/actions/workflows/build-and-publish-image.yml)
[![Lint (go)](https://github.com/PDOK/gokoala/actions/workflows/lint-go.yml/badge.svg)](https://github.com/PDOK/gokoala/actions/workflows/lint-go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/PDOK/gokoala)](https://goreportcard.com/report/github.com/PDOK/gokoala)
[![Coverage (go)](https://github.com/PDOK/gokoala/wiki/coverage.svg)](https://raw.githack.com/wiki/PDOK/gokoala/coverage.html)
[![GitHub license](https://img.shields.io/github/license/PDOK/gokoala)](https://github.com/PDOK/gokoala/blob/master/LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/pdok/gokoala.svg)](https://hub.docker.com/r/pdok/gokoala)

## Description

This application offers location search and geocoding.

## Build

```bash
docker build -t pdok/gokoala .
```

## Run

```bash
NAME:
   GoKoala - Cloud Native OGC APIs server, written in Go

USAGE:
   GoKoala [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --host value            bind host for OGC server (default: "0.0.0.0") [$HOST]
   --port value            bind port for OGC server (default: 8080) [$PORT]
   --debug-port value      bind port for debug server (disabled by default), do not expose this port publicly (default: -1) [$DEBUG_PORT]
   --shutdown-delay value  delay (in seconds) before initiating graceful shutdown (e.g. useful in k8s to allow ingress controller to update their endpoints list) (default: 0) [$SHUTDOWN_DELAY]
   --config-file value     reference to YAML configuration file [$CONFIG_FILE]
   --openapi-file value    reference to a (customized) OGC OpenAPI spec for the dynamic parts of your OGC API [$OPENAPI_FILE]
   --enable-trailing-slash allow API calls to URLs with a trailing slash. (default: false) [$ALLOW_TRAILING_SLASH]
   --enable-cors           enable Cross-Origin Resource Sharing (CORS) as required by OGC API specs. Disable if you handle CORS elsewhere. (default: false) [$ENABLE_CORS]
   --help, -h              show help
```

Example (config-file is mandatory):

```docker
docker run -v `pwd`/examples:/examples -p 8080:8080 -it pdok/gokoala --config-file /examples/config_vectortiles.yaml
```

Now open <http://localhost:8080>. See [examples](examples) for more details.

### Observability

#### Health checks

Health endpoint is available on `/health`.

#### Profiling

Besides the main OGC server GoKoala can also start a debug server. This server
binds to localhost and a different port which you must specify using the
`--debug-port` flag. You shouldn't expose this port publicly but only access it
through a tunnel/port-forward. The debug server exposes `/debug` for use by
[pprof](https://go.dev/blog/pprof). For example with `--debug-port 9001`:

- Create a tunnel to the debug server e.g. in k8s: `kubectl port-forward
gokoala-75f59d57f4-4nd6q 9001:9001`
- Create CPU profile: `go tool pprof
http://localhost:9001/debug/pprof/profile?seconds=20`
- Start pprof visualization `go tool pprof -http=":8000" pprofbin <path to pb.gz
file>`
- Open <http://localhost:8000> to explore CPU flamegraphs and such.

A similar flow can be used to profile memory issues.

## Develop

### Build/run as Go application

Make sure [SpatiaLite](https://www.gaia-gis.it/fossil/libspatialite/index), `openssl` and `curl` are installed. 
Also make sure `gcc` or similar is available since the application uses cgo.

```
go build -o gokoala cmd/main.go
./gokoala
```

To troubleshoot, review the [Dockerfile](./Dockerfile) since compilation also happens there.
Optionally set `SPATIALITE_LIBRARY_PATH=/path/to/spatialite` when SpatiaLite isn't found.

### Linting

Install [golangci-lint](https://golangci-lint.run/usage/install/) and run `golangci-lint run`
from the root.

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
- Reopen the project (or restart IDE). Now you'll have full IDE support in the GoKoala templates.

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
- Now you'll have IDE support in the GoKoala templates.

## Misc

### How to Contribute

Make a pull request...

### Contact

Contacting the maintainers can be done through the issue tracker.
