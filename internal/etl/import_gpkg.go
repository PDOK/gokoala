package etl

import (
	"log"

	"github.com/PDOK/gomagpie/config"
	"github.com/PDOK/gomagpie/internal/etl/extract"
)

func ImportGeoPackage(cfg *config.Config, gpkgPath string, featureTable string, pageSize int,
	synonymsPath string, substitutionsPath string, targetDbConn string) error {

	searchCfg := getSearchConfigForTable(cfg, featureTable)

	offset := 0
	g, err := extract.NewGeoPackage(gpkgPath)
	if err != nil {
		return err
	}
	for {
		batch, err := g.Extract(featureTable, searchCfg.Fields, pageSize, offset)
		if err != nil {
			log.Fatalf("failed importing feature table %s: %v", featureTable, err)
		}
		if len(batch) == 0 {
			break
		}
		offset += pageSize

		println(len(batch))
	}
	//
	// query rows (select + rows.next) to slice of structs, with limit+offset
	// transform data
	// copy data to postgres using pgx.copyfromslice
	println(synonymsPath)      // TODO
	println(substitutionsPath) // TODO
	println(targetDbConn)      // TODO
	return nil
}

func getSearchConfigForTable(cfg *config.Config, featureTable string) *config.Search {
	var searchCfg *config.Search
	for _, coll := range cfg.Collections {
		if coll.ID == featureTable {
			if coll.Search != nil {
				searchCfg = coll.Search
			}
		}
	}
	if searchCfg == nil {
		log.Fatalf("no search configuration found for feature table: %s", featureTable)
	}
	return searchCfg
}
