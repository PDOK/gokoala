package ogc

import (
	"github.com/PDOK/gomagpie/internal/engine"
	"github.com/PDOK/gomagpie/internal/ogc/common/core"
	"github.com/PDOK/gomagpie/internal/ogc/common/geospatial"
)

func SetupBuildingBlocks(engine *engine.Engine) {
	// OGC Common Part 1, will always be started
	core.NewCommonCore(engine)

	// OGC Common part 2
	if engine.Config.HasCollections() {
		geospatial.NewCollections(engine)
	}
}
