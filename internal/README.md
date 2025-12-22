- The `ogc` [package](ogc/README.md) contains logic per specific OGC API
  building block.
- The `engine` package should contain general logic. `ogc` may reference
  `engine`. **The other way around is not allowed!**
- The `etl` package contains extract-transform-load logic to create the search index.