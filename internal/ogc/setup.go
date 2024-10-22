package ogc

import (
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/common/core"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
)

func SetupBuildingBlocks(engine *engine.Engine) {
	// OGC Common Part 1, will always be started
	core.NewCommonCore(engine)

	// OGC Common part 2
	if engine.Config.HasCollections() {
		geospatial.NewCollections(engine)
	}
}
