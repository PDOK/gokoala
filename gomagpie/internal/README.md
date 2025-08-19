- The `engine` package should contain general logic.

The other packages may reference `engine`, the other way around is not allowed!
- The `ogc` [package](ogc/README.md) contains logic per specific OGC API building block. Which is primarily OGC API Common in the case of Gomagpie.
- The `search` package contains front facing location search and geocoding logic.
- The `etl` package contains extract-transform-load logic to create the search index.