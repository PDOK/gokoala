package etl

import (
	"fmt"
	"log"

	"github.com/PDOK/gomagpie/config"
	"github.com/PDOK/gomagpie/internal/etl/extract"
	"github.com/PDOK/gomagpie/internal/etl/load"
	t "github.com/PDOK/gomagpie/internal/etl/transform"
)

// Extract the 'E' in ETL
type Extract interface {
	// Extract raw records from source database to be transformed and loaded into target search indexs
	Extract(featureTable string, fields []string, limit int, offset int) ([]t.RawRecord, error)

	// Close connection to source database
	Close() error
}

// Load the 'L' in ETL
type Load interface {
	// Init the target database by creating an empty search index
	Init() error

	// Load records into search index, and in the process transform (the 'T' in ETL) the records
	// from raw source records to the desired target records
	Load(records []t.RawRecord, collection config.GeoSpatialCollection) (int64, error)

	// Close connection to target database
	Close() error
}

// CreateSearchIndex creates empty search index in target database
func CreateSearchIndex(dbConn string) error {
	db, err := load.NewPostgis(dbConn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Init()
}

// ImportGeoPackage import source data into target search index using extract-transform-load principle
func ImportGeoPackage(cfg *config.Config, gpkgPath string, featureTable string, pageSize int,
	synonymsPath string, substitutionsPath string, targetDbConn string) error {

	log.Printf("start importing GeoPackage %s into Postgres search index", gpkgPath)
	collection, err := getCollectionForTable(cfg, featureTable)
	if err != nil {
		return err
	}
	if collection.Search == nil {
		return fmt.Errorf("no search configuration found for feature table: %s", featureTable)
	}

	source, err := extract.NewGeoPackage(gpkgPath)
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := load.NewPostgis(targetDbConn)
	if err != nil {
		return err
	}
	defer target.Close()

	offset := 0
	for {
		batch, err := source.Extract(featureTable, collection.Search.Fields, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed reading source: %w", err)
		}
		if len(batch) == 0 {
			break // no more batches of records to load into search index
		}
		loaded, err := target.Load(batch, collection)
		if err != nil {
			return fmt.Errorf("failed importing into target: %w", err)
		}
		log.Printf("imported %d records from GeoPackage into Postgres search index", loaded)
		offset += pageSize
	}
	println(synonymsPath)      // TODO
	println(substitutionsPath) // TODO

	log.Printf("done importing GeoPackage %s into Postgres search index", gpkgPath)
	return nil
}

func getCollectionForTable(cfg *config.Config, table string) (config.GeoSpatialCollection, error) {
	for _, coll := range cfg.Collections {
		if coll.ID == table {
			return coll, nil
		}
	}
	return config.GeoSpatialCollection{}, fmt.Errorf("no configured collection for feature table: %s", table)
}
