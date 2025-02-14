package etl

import (
	"fmt"
	"log"
	"strings"

	"github.com/PDOK/gomagpie/config"
	"github.com/PDOK/gomagpie/internal/etl/extract"
	"github.com/PDOK/gomagpie/internal/etl/load"
	t "github.com/PDOK/gomagpie/internal/etl/transform"
	"golang.org/x/text/language"
)

// Extract - the 'E' in ETL. Datasource agnostic interface to extract source data.
type Extract interface {

	// Extract raw records from source database to be transformed and loaded into target search index
	Extract(table config.FeatureTable, fields []string, externaFidFields []string, where string, limit int, offset int) ([]t.RawRecord, error)

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
	Init(index string, lang language.Tag) error

	// Load records into search index
	Load(records []t.SearchIndexRecord, index string) (int64, error)

	// Optimize once ETL is completed (optionally)
	Optimize() error

	// Close connection to target database
	Close()
}

// CreateSearchIndex creates empty search index in target database
func CreateSearchIndex(dbConn string, searchIndex string, lang language.Tag) error {
	db, err := newTargetToLoad(dbConn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Init(searchIndex, lang)
}

// ImportFile import source data into target search index using extract-transform-load principle
//
//nolint:funlen
func ImportFile(collection config.GeoSpatialCollection, searchIndex string, filePath string, substitutionsFile string,
	synonymsFile string, table config.FeatureTable, pageSize int, dbConn string) error {

	details := fmt.Sprintf("file %s (feature table '%s', collection '%s') into search index %s", filePath, table.Name, collection.ID, searchIndex)
	log.Printf("start import of %s", details)

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

	transformer, err := newTransformer(substitutionsFile, synonymsFile)
	if err != nil {
		return err
	}

	// import records in batches depending on page size
	offset := 0
	for {
		log.Println("---")
		log.Printf("extracting source records from offset %d", offset)
		externalFidFields := []string{}
		if collection.Search.ETL.ExternalFid != nil {
			externalFidFields = collection.Search.ETL.ExternalFid.Fields
		}
		sourceRecords, err := source.Extract(table, collection.Search.Fields, externalFidFields, collection.Search.ETL.Filter, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed extracting source records: %w", err)
		}
		sourceRecordCount := len(sourceRecords)
		if sourceRecordCount == 0 {
			break // no more batches of records to extract
		}
		log.Printf("extracted %d source records, starting transform", sourceRecordCount)
		targetRecords, err := transformer.Transform(sourceRecords, collection)
		if err != nil {
			return fmt.Errorf("failed to transform raw records to search index records: %w", err)
		}
		log.Printf("transform completed, %d source records transformed into %d target records", sourceRecordCount, len(targetRecords))
		loaded, err := target.Load(targetRecords, searchIndex)
		if err != nil {
			return fmt.Errorf("failed loading records into target: %w", err)
		}
		log.Printf("loaded %d records into target search index: '%s'", loaded, searchIndex)
		offset += pageSize
	}
	log.Printf("completed import of %s", details)

	log.Println("start optimizing")
	if err = target.Optimize(); err != nil {
		return fmt.Errorf("failed optimizing: %w", err)
	}
	log.Println("completed optimizing")
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
		return load.NewPostgres(dbConn)
	}
	// add new targets here (elasticsearch, solr, etc)
	return nil, fmt.Errorf("unsupported target database connection: %s", dbConn)
}

func newTransformer(substitutionsFile string, synonymsFile string) (Transform, error) {
	return t.NewTransformer(substitutionsFile, synonymsFile)
}
