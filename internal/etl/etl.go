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

// Extract - the 'E' in ETL. Datasource agnostic interface to extract source data.
type Extract interface {

	// Extract raw records from source database to be transformed and loaded into target search index
	Extract(table config.FeatureTable, fields []string, where string, limit int, offset int) ([]t.RawRecord, error)

	// Close connection to source database
	Close()
}

// Transform - the 'T' in ETL. Logic to transform raw records to search index records
type Transform interface {

	// Transform each raw record in one or more search records depending on the given configuration
	Transform(records []t.RawRecord, collection config.GeoSpatialCollection) ([]t.SearchIndexRecord, error)
}

// Load - the 'L' in ETL. Datasource agnostic interface to load data into target database.
type Load interface {

	// Init the target database by creating an empty search index
	Init(index string) error

	// Load records into search index
	Load(records []t.SearchIndexRecord, index string) (int64, error)

	// Close connection to target database
	Close()
}

// CreateSearchIndex creates empty search index in target database
func CreateSearchIndex(dbConn string, searchIndex string) error {
	db, err := newTargetToLoad(dbConn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Init(searchIndex)
}

// ImportFile import source data into target search index using extract-transform-load principle
func ImportFile(cfg *config.Config, searchIndex string, filePath string, table config.FeatureTable,
	pageSize int, dbConn string) error {

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

	target, err := newTargetToLoad(dbConn)
	if err != nil {
		return err
	}
	defer target.Close()

	transformer := newTransformer()

	// import records in batches depending on page size
	offset := 0
	for {
		sourceRecords, err := source.Extract(table, collection.Search.Fields, collection.Search.ETLFilter, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed extracting source records: %w", err)
		}
		if len(sourceRecords) == 0 {
			break // no more batches of records to extract
		}
		targetRecords, err := transformer.Transform(sourceRecords, collection)
		if err != nil {
			return fmt.Errorf("failed to transform raw records to search index records: %w", err)
		}
		loaded, err := target.Load(targetRecords, searchIndex)
		if err != nil {
			return fmt.Errorf("failed loading records into target: %w", err)
		}
		log.Printf("imported %d records into search index", loaded)
		offset += pageSize
	}

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

func newTransformer() Transform {
	return t.Transformer{}
}

func getCollectionForTable(cfg *config.Config, table config.FeatureTable) (config.GeoSpatialCollection, error) {
	for _, coll := range cfg.Collections {
		if coll.ID == table.Name {
			return coll, nil
		}
	}
	return config.GeoSpatialCollection{}, fmt.Errorf("no configured collection for feature table: %s", table)
}
