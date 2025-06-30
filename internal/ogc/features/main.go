package features

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/geopackage"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/postgres"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

const (
	templatesDir = "internal/ogc/features/templates/"
)

type Features struct {
	engine                    *engine.Engine
	datasources               map[datasourceKey]ds.Datasource
	axisOrderBySRID           map[int]domain.AxisOrder
	configuredCollections     map[string]config.GeoSpatialCollection
	configuredPropertyFilters map[string]ds.PropertyFiltersWithAllowedValues
	schemas                   map[string]domain.Schema

	html *htmlFeatures
	json *jsonFeatures
}

// NewFeatures Bootstraps OGC API Features logic
func NewFeatures(e *engine.Engine) *Features {
	datasources := createDatasources(e)
	axisOrderBySRID := determineAxisOrder(datasources)
	configuredCollections := cacheConfiguredFeatureCollections(e)
	configuredPropertyFilters := configurePropertyFiltersWithAllowedValues(datasources, configuredCollections)

	schemas := renderSchemas(e, datasources)
	rebuildOpenAPI(e, datasources, configuredPropertyFilters, schemas)

	f := &Features{
		engine:                    e,
		datasources:               datasources,
		axisOrderBySRID:           axisOrderBySRID,
		configuredCollections:     configuredCollections,
		configuredPropertyFilters: configuredPropertyFilters,
		schemas:                   schemas,
		html:                      newHTMLFeatures(e),
		json:                      newJSONFeatures(e),
	}

	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/items", f.Features())
	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/items/{featureId}", f.Feature())
	e.Router.Get(geospatial.CollectionsPath+"/{collectionId}/schema", f.Schema())
	return f
}

type datasourceKey struct {
	srid         int
	collectionID string
}

type datasourceConfig struct {
	collections       config.GeoSpatialCollections
	ds                config.Datasource
	transformOnTheFly bool
}

func createDatasources(e *engine.Engine) map[datasourceKey]ds.Datasource {
	configured := make(map[datasourceKey]*datasourceConfig, len(e.Config.OgcAPI.Features.Collections))

	// configure collection specific datasources first
	configureCollectionDatasources(e, configured)
	// now configure top-level datasources, for the whole dataset. But only when
	// there's no collection-specific datasource already configured
	configureTopLevelDatasources(e, configured)

	if len(configured) == 0 {
		log.Fatal("no datasource(s) configured for OGC API Features, check config")
	}

	// now we have a mapping from collection+projection => desired datasource (the 'configured' map).
	// but the actual datasource connection still needs to be CREATED and associated with these collections.
	// this is what we're going to do now, but in the process we need to make sure no duplicate datasources
	// are instantiated: since multiple collections can point to the same datasource and we only what to have a single
	// datasource/connection-pool serving those collections.
	createdDatasources := make(map[config.Datasource]ds.Datasource)
	result := make(map[datasourceKey]ds.Datasource, len(configured))
	for k, cfg := range configured {
		if cfg == nil {
			continue
		}
		existing, ok := createdDatasources[cfg.ds]
		if !ok {
			// make sure to only create a new datasource when it hasn't already been done before (for another collection)
			created := newDatasource(e, cfg.collections, cfg.ds, cfg.transformOnTheFly)
			createdDatasources[cfg.ds] = created
			result[k] = created
		} else {
			result[k] = existing
		}
	}
	return result
}

func determineAxisOrder(datasources map[datasourceKey]ds.Datasource) map[int]domain.AxisOrder {
	log.Println("start determining axis order for all configured CRS's")
	order := map[int]domain.AxisOrder{
		domain.WGS84SRID: domain.AxisOrderXY, // We know CRS84 is XY, see https://spatialreference.org/ref/ogc/CRS84/
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	for key := range datasources {
		mu.Lock()
		_, exists := order[key.srid]
		mu.Unlock()

		if !exists {
			wg.Add(1)

			// use goroutine to avoid blocking on GetAxisOrder(). The mutex is necessary
			// to avoid race conditions on the map.
			go func() {
				defer wg.Done()

				axisOrder, err := GetAxisOrder(domain.SRID(key.srid))
				if err != nil {
					log.Printf("Warning: failed to determine whether EPSG:%d needs "+
						"swap of X/Y axis: %v. Defaulting to XY order.", key.srid, err)
					axisOrder = domain.AxisOrderXY
				}

				mu.Lock()
				order[key.srid] = axisOrder
				mu.Unlock()
			}()
		}
	}
	wg.Wait()

	log.Println("done determining axis order for all configured CRSs")
	return order
}

func cacheConfiguredFeatureCollections(e *engine.Engine) map[string]config.GeoSpatialCollection {
	result := make(map[string]config.GeoSpatialCollection)
	for _, collection := range e.Config.OgcAPI.Features.Collections {
		result[collection.ID] = collection
	}
	return result
}

func configurePropertyFiltersWithAllowedValues(datasources map[datasourceKey]ds.Datasource,
	collections map[string]config.GeoSpatialCollection) map[string]ds.PropertyFiltersWithAllowedValues {

	result := make(map[string]ds.PropertyFiltersWithAllowedValues)
	for k, datasource := range datasources {
		result[k.collectionID] = datasource.GetPropertyFiltersWithAllowedValues(k.collectionID)
	}

	// sanity check to make sure datasources return all configured property filters.
	for _, collection := range collections {
		actual := len(result[collection.ID])
		if collection.Features != nil && collection.Features.Filters.Properties != nil {
			expected := len(collection.Features.Filters.Properties)
			if expected != actual {
				log.Fatalf("number of property filters received from datasource for collection '%s' does not "+
					"match the number of configured property filters. Expected filters: %d, got from datasource: %d",
					collection.ID, expected, actual)
			}
		}
	}
	return result
}

// configureTopLevelDatasources configures top-level datasources - in one or multiple CRS's - which can be
// used by one or multiple collections (e.g., one GPKG that holds an entire dataset)
func configureTopLevelDatasources(e *engine.Engine, result map[datasourceKey]*datasourceConfig) {
	cfg := e.Config.OgcAPI.Features
	if cfg.Datasources == nil {
		return
	}
	// Ahead-of-time WGS84
	var defaultDS *datasourceConfig
	for _, coll := range cfg.Collections {
		key := datasourceKey{srid: domain.WGS84SRID, collectionID: coll.ID}
		if result[key] == nil {
			if defaultDS == nil {
				defaultDS = &datasourceConfig{cfg.Collections, cfg.Datasources.DefaultWGS84, false}
			}
			result[key] = defaultDS
		}
	}

	// Ahead-of-time additional SRSs
	for _, additional := range cfg.Datasources.Additional {
		for _, coll := range cfg.Collections {
			srid, err := domain.EpsgToSrid(additional.Srs)
			if err != nil {
				log.Fatal(err)
			}
			key := datasourceKey{srid: srid.GetOrDefault(), collectionID: coll.ID}
			if result[key] == nil {
				result[key] = &datasourceConfig{cfg.Collections, additional.Datasource, false}
			}
		}
	}

	// On-the-fly SRSs -- add these as last since we prefer ahead-of-time projections
	for _, otf := range cfg.Datasources.OnTheFly {
		for _, coll := range cfg.Collections {
			for _, srs := range otf.SupportedSrs {
				srid, err := domain.EpsgToSrid(srs.Srs)
				if err != nil {
					log.Fatal(err)
				}
				key := datasourceKey{srid: srid.GetOrDefault(), collectionID: coll.ID}
				if result[key] == nil {
					result[key] = &datasourceConfig{cfg.Collections, otf.Datasource, true}
				}
			}
		}
	}
}

// configureCollectionDatasources configures datasources - in one or multiple CRS's - which are specific
// to a certain collection (e.g., a separate GPKG per collection)
func configureCollectionDatasources(e *engine.Engine, result map[datasourceKey]*datasourceConfig) {
	cfg := e.Config.OgcAPI.Features
	for _, coll := range cfg.Collections {
		if coll.Features == nil || coll.Features.Datasources == nil {
			continue
		}
		// Ahead-of-time WGS84
		defaultDS := &datasourceConfig{cfg.Collections, coll.Features.Datasources.DefaultWGS84, false}
		result[datasourceKey{srid: domain.WGS84SRID, collectionID: coll.ID}] = defaultDS

		// Ahead-of-time additional SRSs
		for _, additional := range coll.Features.Datasources.Additional {
			srid, err := domain.EpsgToSrid(additional.Srs)
			if err != nil {
				log.Fatal(err)
			}
			additionalDS := &datasourceConfig{cfg.Collections, additional.Datasource, false}
			result[datasourceKey{srid: srid.GetOrDefault(), collectionID: coll.ID}] = additionalDS
		}

		// On-the-fly SRSs -- add these as last since we prefer ahead-of-time projections
		for _, otf := range coll.Features.Datasources.OnTheFly {
			for _, srs := range otf.SupportedSrs {
				srid, err := domain.EpsgToSrid(srs.Srs)
				if err != nil {
					log.Fatal(err)
				}
				additionalDS := &datasourceConfig{cfg.Collections, otf.Datasource, true}
				result[datasourceKey{srid: srid.GetOrDefault(), collectionID: coll.ID}] = additionalDS
			}
		}
	}
}

func newDatasource(e *engine.Engine, collections config.GeoSpatialCollections,
	dsConfig config.Datasource, transformOnTheFly bool) ds.Datasource {
	var datasource ds.Datasource
	var err error
	switch {
	case dsConfig.GeoPackage != nil:
		datasource, err = geopackage.NewGeoPackage(collections, *dsConfig.GeoPackage, transformOnTheFly)
	case dsConfig.Postgres != nil:
		datasource, err = postgres.NewPostgres(collections, *dsConfig.Postgres, transformOnTheFly)
	default:
		log.Fatal("got unknown datasource type")
	}
	if err != nil {
		log.Fatal(err)
	}
	e.RegisterShutdownHook(datasource.Close)
	return datasource
}

func handleCollectionNotFound(w http.ResponseWriter, collectionID string) {
	msg := fmt.Sprintf("collection %s doesn't exist in this features service", collectionID)
	log.Println(msg)
	engine.RenderProblem(engine.ProblemNotFound, w, msg)
}
