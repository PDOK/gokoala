package ogc

import (
	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/common/core"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	"github.com/PDOK/gokoala/internal/ogc/features"
	"github.com/PDOK/gokoala/internal/ogc/features_search"
	"github.com/PDOK/gokoala/internal/ogc/geovolumes"
	"github.com/PDOK/gokoala/internal/ogc/processes"
	"github.com/PDOK/gokoala/internal/ogc/styles"
	"github.com/PDOK/gokoala/internal/ogc/tiles"
)

func SetupBuildingBlocks(engine *engine.Engine, rewritesFile, synonymsFile string) error {
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
	collectionTypes := geospatial.NewCollectionTypes(nil)
	if engine.Config.OgcAPI.Features != nil {
		f := features.NewFeatures(engine)
		collectionTypes = f.GetCollectionTypes()
	}
	// Features Search API, build on top of the OGC Features API
	if engine.Config.OgcAPI.FeaturesSearch != nil {
		ds := features.CreateDatasources(
			config.NewSearchConfig(engine.Config.OgcAPI.FeaturesSearch), engine.RegisterShutdownHook)
		ao := features.DetermineAxisOrder(ds)
		_, err := features_search.NewSearch(engine, ds, ao, rewritesFile, synonymsFile)
		if err != nil {
			return err
		}
	}
	// OGC Processes API
	if engine.Config.OgcAPI.Processes != nil {
		processes.NewProcesses(engine)
	}

	// OGC Common Part 1, this will always be started
	core.NewCommonCore(engine, core.ExtraConformanceClasses{AttributesConformance: collectionTypes.HasAttributes()})
	// OGC Common part 2
	if engine.Config.HasCollections() {
		geospatial.NewCollections(engine, collectionTypes)
	}
	return nil
}
