package ogc

import (
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/common/core"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	"github.com/PDOK/gokoala/internal/ogc/features"
	"github.com/PDOK/gokoala/internal/ogc/geovolumes"
	"github.com/PDOK/gokoala/internal/ogc/processes"
	"github.com/PDOK/gokoala/internal/ogc/styles"
	"github.com/PDOK/gokoala/internal/ogc/tiles"
)

func SetupBuildingBlocks(engine *engine.Engine) {
	// OGC Common Part 1, will always be started
	core.NewCommonCore(engine)

	// OGC Common part 2
	if engine.Config.HasCollections() {
		geospatial.NewCollections(engine)
	}
	// OGC 3D GeoVolumes API
	if engine.Config.OgcAPI.GeoVolumes != nil {
		geovolumes.NewThreeDimensionalGeoVolumes(engine)
	}
	// OGC Tiles API
	if engine.Config.OgcAPI.Tiles != nil {
		tiles.NewTiles(engine)
	}
	// OGC Styles API
	if engine.Config.OgcAPI.Styles != nil {
		styles.NewStyles(engine)
	}
	// OGC Features API
	if engine.Config.OgcAPI.Features != nil {
		features.NewFeatures(engine)
	}
	// OGC Processes API
	if engine.Config.OgcAPI.Processes != nil {
		processes.NewProcesses(engine)
	}
}
