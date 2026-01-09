# OGC API

OGC APIs are constructed by different building blocks. These building blocks
are composed of the different [OGC API standards](https://ogcapi.ogc.org/).
Each OGC building block resides in its own Go package.

- package `common` represents OGC API Common.
- package `features` represents OGC API Features. It's the most extensive building block.
- package `feature_search` is **NOT** part of the OGC API standard, but build on top of OGC API Features.
  It's an API for geocoding and searching features across multiple collections. The need for this
  functionality is also expressed in an
  OGC [feature request](https://github.com/opengeospatial/ogcapi-features/issues/939). This building block was
  previously known as [gomagpie](https://github.com/PDOK/gomagpie).
- package `geovolumes` represents OGC API 3D GeoVolumes / 3D Tiles.
- package `processes` represents OGC API Processes.
- package `styles` represents OGC API Styles.
- package `tiles` represents OGC API Tiles.

When coding, we try to use the naming convention as used by the OGC, so it is clear
which specification or part is referred to in code.

## Coding

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

See [README](../../README.md) in the root.

#### Tip: handling JSON

When generating JSON arrays using templates you need to be aware of trailing
commas. The last element in an array must not contain a comma. To prevent this,
either:

- Add the comma in front of array items
- Use the index of a `range` to check array position and place the comma based
  on the index
- The most comprehensive solution is to use:

```gotemplate
{{ $first := true }}
{{ range $_, $element := .}}
{{if not $first}}, {{else}} {{$first = false}} {{end}}
{{$element.Name}}
{{end}}
```
