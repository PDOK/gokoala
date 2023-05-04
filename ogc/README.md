# OGC API

OGC APIs are constructed by different building blocks. These building blocks
are composed of the different [OGC API standards](https://ogcapi.ogc.org/) as
defined by the OGC.

## Parts

The following parts are or will be supported.

- [X] common
  - part 1: core
  - part 2: geospatial collections
- [ ] features
- [x] geovolumes (3D GeoVolumes, DRAFT)
- [ ] maps
- [X] styles
- [X] tiles

## Coding

### Naming convention

When coding we will try to use the naming convention as that is used by the OGC.
So it is clean which specification or part is referred to in the code.

### Templates

We use templates to generate static/pre-defined API responses based on
the given GoKoala configuration file. Lots of OGC API responses can be
statically generated. Generation happens at startup and results are served
from memory when an API request is received. Benefits of this approach are:

- Lightning fast responses to API calls since everything is served from memory
- Fail fast since validation is performed during startup

#### Duplication

We will have duplication between JSON and HTML templates: that's ok. They're
different representations of the same data. Don't try to be clever and
"optimize" it. The duplication is pretty obvious/visible since the files only
differ by extension, so it's clear any changes need to be done in both
representations. Having independent files keeps the templates simple and
flexible.

#### IDE support

See [README](../README.md) in the root.

#### Tip: handling JSON

When generating JSON arrays using templates you need to be aware of trailing
commas. The last element in an array must not contain a comma. To prevent this,
either:

- Add the comma in front of array items
- Use the index of a `range` to check array position and place the comma based
  on the index
- The most comprehensive solution is to use:

```jinja
{{ $first := true }}
{{ range $_, $element := .}}
    {{if not $first}}, {{else}} {{$first = false}} {{end}}
    {{$element.Name}}
{{end}}
```
