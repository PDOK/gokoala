<p align="center">
<img src="docs/gopher-koala.png" alt="GoKoala" title="GoKoala" width="300" />
</p>

# GoKoala

_Cloud Native OGC APIs server, written in Go._ 

![Build](https://github.com/PDOK/gokoala/actions/workflows/build-and-publish-image.yml/badge.svg) 
[![Go Report Card](https://goreportcard.com/badge/github.com/PDOK/gokoala)](https://goreportcard.com/report/github.com/PDOK/gokoala)
[![GitHub
license](https://img.shields.io/github/license/PDOK/gokoala)](https://github.com/PDOK/gokoala/blob/master/LICENSE)
[![Docker
Pulls](https://img.shields.io/docker/pulls/pdok/gokoala.svg)](https://hub.docker.com/r/pdok/gokoala)

## Description

This server implements modern OGC APIs such as OGC Common Core, OGC Tiles, OGC
Styles. In the future other APIs like OGC Features or OGC Maps may be added. The
goal of this server is to keep a narrow focus and not implement every aspect of 
these APIs, for complex logic this application will delegate to other implementations. 
For example vector tiles hosting is delegated to a vector tile engine or object storage, 
raster map hosting may be delegated to a WMS server, etc.

This application is deliberately not multi-tenant, it exposes an OGC API for
_one_ dataset.

## Features

- [OGC API Common](https://ogcapi.ogc.org/common/) support: Serves landing page
  and conformance declaration. Also serves OpenAPI specification and interactive
  Swagger UI.
  - Comes with default OGC OpenAPI specs out-of-the box with option to overwrite
    with your own custom spec.
- [OGC API Tiles](https://ogcapi.ogc.org/tiles/) support: Serves HTML, JSON and
  TileJSON metadata. Serves as a proxy in front of a vector tiles engine of your
  choosing.
- [OGC API Styles](https://ogcapi.ogc.org/styles/) support: Serves HTML and JSON
  representation of supported styles.
- [OGC API 3D GeoVolumes](https://ogcapi.ogc.org/geovolumes/) support: Serves
  HTML and JSON metadata and functions as a proxy in front of a [3D
  Tiles](https://www.ogc.org/standard/3dtiles/) server of your choosing.

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
   --shutdown-delay value  delay (in seconds) before initiating graceful shutdown (default: 0) [$SHUTDOWN_DELAY]
   --config-file value     reference to YAML configuration file [$CONFIG_FILE]
   --openapi-file value    reference to a (customized) OGC OpenAPI spec for the dynamic parts of your OGC API [$OPENAPI_FILE]
   --resources-dir value   reference to a directory containing static files, like images [$RESOURCES_DIR]
   --help, -h              show help

```

Example:

```docker
docker run -v `pwd`/examples:/examples -p 8080:8080 -it pdok/gokoala --config-file /examples/config_vectortiles.yaml --resources-dir /examples/resources
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

- The `ogc` [package](ogc/README.md) contains logic per specific OGC API
  building block.
- The `engine` package should contain general logic. `ogc` may reference
  `engine`.
  > :warning: The other way around is not allowed!
- The OGC API Specifications are multi-part standards, this means technically
  that parts can be enabled or disabled, the code should reflect this.
- Geospatial related configuration is done through the config file.
- Fail fast, fail hard: do as much pre-processing/validation on startup instead
  of during request handling.
- Assets/templates/etc are explicitly included in the Docker image, see copy
  commands in [Dockerfile](Dockerfile).

### Linting

Install [golangci-lint](https://golangci-lint.run/usage/install/) and run `golangci-lint run`
from the root.

### IntelliJ / GoLand

- Install the [Go
  Template](https://plugins.jetbrains.com/plugin/10581-go-template) plugin
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

- Install the [Go
  Template](https://marketplace.visualstudio.com/items?itemName=jinliming2.vscode-go-template)
  extension
- Open Extension Settings and add the following file patterns:
  - `"*.go.html"`
  - `"*.go.json"`
  - `"*.go.tilejson"`
- Also add `html` and `json` to the list of Go template languages.
- Now you'll have IDE support in the GoKoala templates.

## Misc

### How to Contribute

Make a pull request...

### Contact

Contacting the maintainers can be done through the issue tracker.
