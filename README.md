<p align="center">
<img src="docs/gopher-koala.png" alt="GoKoala" title="GoKoala" width="300" />
</p>

# GoKoala

_Cloud Native OGC APIs server, written in Go._ 

[![Build](https://github.com/PDOK/gokoala/actions/workflows/build-and-publish-image.yml/badge.svg)](https://github.com/PDOK/gokoala/actions/workflows/build-and-publish-image.yml)
[![Lint](https://github.com/PDOK/gokoala/actions/workflows/lint.yml/badge.svg)](https://github.com/PDOK/gokoala/actions/workflows/lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/PDOK/gokoala)](https://goreportcard.com/report/github.com/PDOK/gokoala)
[![GitHub
license](https://img.shields.io/github/license/PDOK/gokoala)](https://github.com/PDOK/gokoala/blob/master/LICENSE)
[![Docker
Pulls](https://img.shields.io/docker/pulls/pdok/gokoala.svg)](https://hub.docker.com/r/pdok/gokoala)

## Description

This server implements modern OGC APIs such as Common, Tiles, Styles. The goal of 
this server is to keep a narrow focus and not implement every aspect of 
these APIs, for complex logic this application will delegate to other implementations. 
For example vector tiles hosting is delegated to a vector tile engine or object storage, 
raster map hosting may be delegated to a WMS server, etc.

This application is deliberately not multi-tenant, it exposes an OGC API for
_one_ dataset.

## Features

- [OGC API Common](https://ogcapi.ogc.org/common/) serves landing page
  and conformance declaration. Also serves OpenAPI specification and interactive
  Swagger UI.
  - Comes with default OGC OpenAPI specs out-of-the box with option to overwrite
    with your own custom spec.
- [OGC API Tiles](https://ogcapi.ogc.org/tiles/) serves HTML, JSON and
  TileJSON metadata. Act as a proxy in front of a vector tiles engine of your
  choosing. Currently 3 projections (RD, ETRS89 and WebMercator) are supported.
- [OGC API Styles](https://ogcapi.ogc.org/styles/) serves HTML and JSON representation of supported styles.
- [OGC API 3D GeoVolumes](https://ogcapi.ogc.org/geovolumes/) serves HTML and JSON metadata and functions as a proxy 
  in front of a [3D Tiles](https://www.ogc.org/standard/3dtiles/) server of your choosing.
- [OGC API Processes](https://ogcapi.ogc.org/processes/) act as a passthrough proxy to an OGC API Processes 
  implementation of your choosing, but enables the use of OGC API Common functionality.
- [OGC API Features](https://ogcapi.ogc.org/features/) _in development_.

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
   --allow-trailing-slash  support API calls to URLs with a trailing slash (default: false) [$ALLOW_TRAILING_SLASH]
   --help, -h              show help
```

Example (config-file is mandatory):

```docker
docker run -v `pwd`/examples:/examples -p 8080:8080 -it pdok/gokoala --config-file /examples/config_vectortiles.yaml
```

Now open <http://localhost:8080>

### Configuration file

The configuration file consists of a general section and a section
per OGC API building block (tiles, styles, etc). See [example configuration
files](examples/) for details. You can reference environment variables in the
configuration file. For example to use the `MY_SERVER` env var:

```yaml
ogcApi:
  tiles:
    title: My Dataset
    tileServer: https://${MY_SERVER}/foo/bar
```

### OpenAPI spec

GoKoala ships with OGC OpenAPI support out of the box, see [OpenAPI
specs](assets/openapi-specs/README.md) for details. You can overwrite or extend
the defaults by providing your own spec using the `openapi-file` CLI flag.

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

Design principles:

- Performance and scalability are key!
- The `ogc` [package](ogc/README.md) contains logic per specific OGC API
  building block.
- The `engine` package should contain general logic. `ogc` may reference
  `engine`.
  > :warning: The other way around is not allowed!
- Geospatial related configuration is done through the config file, technical
  configuration (host/port/etc) is done through CLI flags/env variables.
- Fail fast, fail hard: do as much pre-processing/validation on startup instead
  of during request handling.
- Assets/templates/etc should be explicitly included in the Docker image, see COPY
  commands in [Dockerfile](Dockerfile).

### Linting

Install [golangci-lint](https://golangci-lint.run/usage/install/) and run `golangci-lint run`
from the root.

### Webcomponents

GoKoala includes a [vector tile viewer](webcomponents/vectortile-view-component) which is available 
as a Web Component for embedding in HTML pages. To use the vector tile viewer locally when running 
GoKoala outside Docker execute: `hack/build-local-viewer.sh`. This will build the viewer and add 
it to the GoKoala assets.

Note this is only required for local development. When running GoKoala as a container this is
already being taken care of when building the Docker container image.

### IntelliJ / GoLand

- Install the [Go Template](https://plugins.jetbrains.com/plugin/10581-go-template) plugin
- Open `Preferences` > `Editor` > `File Types` select `Go Template files` and
  add the following file patterns:
  - `"*.go.html"`
  - `"*.go.json"`
  - `"*.go.tilejson"`
- Now add template language support by running the
  [setup-jetbrains-gotemplates.sh](hack/setup-jetbrains-gotemplates.sh) script.
- Now you'll have full IDE support in the GoKoala templates.

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
- Also add `html` and `json` to the list of Go template languages.
- Now you'll have IDE support in the GoKoala templates.

### OGC compliance validation

Use the OGC [TEAM Engine](https://cite.opengeospatial.org/teamengine/) to validate 
compliance when available. In the case of OGC API Features follow these steps:

- Run `docker run -p 8081:8080 ogccite/ets-ogcapi-features10`
- Open http://localhost:8081/teamengine/
- Start GoKoala. 
  - When running Docker in a VM (like on macOS) make sure to start GoKoala with base url: http://host.docker.internal:8080.
- Start a new test session in the TEAM Engine against http://localhost:8080 (or http://host.docker.internal:8080).
  - More details in the [features conformance test suite](https://opengeospatial.github.io/ets-ogcapi-features10/).
- Publish test results HTML report in [docs](./docs/ogc-features-test-report) and list below.
  - Test results on [27-09-2023](https://htmlpreview.github.io/?https://github.com/PDOK/gokoala/blob/master/docs/ogc-features-test-report/20230927.html)

## Misc

### How to Contribute

Make a pull request...

### Contact

Contacting the maintainers can be done through the issue tracker.
