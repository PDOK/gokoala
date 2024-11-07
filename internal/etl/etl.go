package etl

import (
	"fmt"
	"log"
	"strings"

	"github.com/PDOK/gomagpie/config"
	"github.com/PDOK/gomagpie/internal/etl/extract"
	"github.com/PDOK/gomagpie/internal/etl/load"
	t "github.com/PDOK/gomagpie/internal/etl/transform"
)

// Extract - the 'E' in ETL
type Extract interface {
	// Extract raw records from source database to be transformed and loaded into target search index
	Extract(table config.FeatureTable, fields []string, limit int, offset int) ([]t.RawRecord, error)

	// Close connection to source database
	Close()
}

// Load - the 'L' in ETL
type Load interface {
	// Init the target database by creating an empty search index
	Init() error

	// Load records into search index, and in the process transform (the 'T' in ETL) the records
	// from raw source records to the desired target records
	Load(records []t.RawRecord, collection config.GeoSpatialCollection) (int64, error)

	// Close connection to target database
	Close()
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

// ImportFile import source data into target search index using extract-transform-load principle
func ImportFile(cfg *config.Config, filePath string, table config.FeatureTable, pageSize int,
	synonymsPath string, substitutionsPath string, targetDbConn string) error {

	log.Println("start importing")
	collection, err := getCollectionForTable(cfg, table)
	if err != nil {
		return err
	}
	if collection.Search == nil {
		return fmt.Errorf("no search configuration found for feature table: %s", table.Name)
	}

	source, err := newSourceToExtract(filePath)
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := newTargetToLoad(targetDbConn)
	if err != nil {
		return err
	}
	defer target.Close()

	offset := 0
	for {
		records, err := source.Extract(table, collection.Search.Fields, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed extracting source records: %w", err)
		}
		if len(records) == 0 {
			break // no more batches of records to extract
		}
		loaded, err := target.Load(records, collection)
		if err != nil {
			return fmt.Errorf("failed loading records into target: %w", err)
		}
		log.Printf("imported %d records into search index", loaded)
		offset += pageSize
	}
	println(synonymsPath)      // TODO
	println(substitutionsPath) // TODO

	log.Println("done importing")
	return nil
}

func newSourceToExtract(filePath string) (Extract, error) {
	if strings.HasSuffix(filePath, ".gpkg") {
		return extract.NewGeoPackage(filePath)
	}
	// add new sources here (csv, zip, parquet, etc)
	return nil, fmt.Errorf("unsupported source file type: %s", filePath)
}

func newTargetToLoad(dbConn string) (Load, error) {
	if strings.HasPrefix(dbConn, "postgres:") {
		return load.NewPostgis(dbConn)
	}
	// add new targets here (elasticsearch, solr, etc)
	return nil, fmt.Errorf("unsupported target database connection: %s", dbConn)
}

func getCollectionForTable(cfg *config.Config, table config.FeatureTable) (config.GeoSpatialCollection, error) {
	for _, coll := range cfg.Collections {
		if coll.ID == table.Name {
			return coll, nil
		}
	}
	return config.GeoSpatialCollection{}, fmt.Errorf("no configured collection for feature table: %s", table)
}
