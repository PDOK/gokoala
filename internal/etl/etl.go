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

	// Extract raw records from the source database to be transformed and loaded into the target search index
	Extract(table config.FeatureTable, fields []string, externaFidFields []string,
		where string, limit int, offset int) ([]t.RawRecord, error)

	// Close connection to the source database
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
	Init(index string, srid int, lang language.Tag) error

	// Get the current version of a collection loaded in the search index
	GetVersion(collectionID string, index string) (string, error)

	// PreLoad hook to execute logic before loading records into the search index.
	// For example, by creating tables or partitions
	PreLoad(collectionID string, index string) error

	// Load records into the search index. Returns the number of records loaded.
	// Assumes the index is already initialized.
	Load(records []t.SearchIndexRecord) (int64, error)

	// PostLoad hook to execute logic after loading records into the search index.
	// For example, by switching partitions or rebuilding indexes.
	PostLoad(collectionID string, index string, collectionVersion string) error

	// Optimize once ETL is completed (optional)
	Optimize() error

	// Close connection to the target database
	Close()
}

// CreateSearchIndex creates an empty search index in the target database
func CreateSearchIndex(dbConn string, searchIndex string, srid int, lang language.Tag) error {
	db, err := newTargetToLoad(dbConn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Init(searchIndex, srid, lang)
}

// GetVersion returns the current version of a collection in the target search index
func GetVersion(dbConn string, collectionID string, searchIndex string) (string, error) {
	db, err := newTargetToLoad(dbConn)
	if err != nil {
		return "", err
	}
	defer db.Close()
	return db.GetVersion(collectionID, searchIndex)
}

// ImportFile import source data into the target search index using extract-transform-load principle
//
//nolint:funlen
func ImportFile(collection config.GeoSpatialCollection, searchIndex string, collectionVersion string, filePath string,
	tables []config.FeatureTable, pageSize int, skipOptimize bool, dbConn string) error {

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
	transformer := t.NewTransformer()

	// pre-load
	if err = target.PreLoad(collection.ID, searchIndex); err != nil {
		return err
	}

	// import each feature table
	for _, table := range tables {
		details := fmt.Sprintf("file %s (feature table '%s', collection '%s') into search index %s", filePath, table.Name, collection.ID, searchIndex)
		log.Printf("start import of %s", details)
		if collection.Search == nil {
			return fmt.Errorf("no search configuration found for feature table: %s", table.Name)
		}

		// import records in batches depending on page size
		offset := 0
		for {
			log.Println("---")
			log.Printf("extracting source records from offset %d", offset)
			var externalFidFields []string
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
			loaded, err := target.Load(targetRecords)
			if err != nil {
				return fmt.Errorf("failed loading records into target: %w", err)
			}
			log.Printf("loaded %d records into target search index: '%s'", loaded, searchIndex)
			offset += pageSize
		}
		log.Printf("completed import of %s", details)
	}

	// post-load
	if err = target.PostLoad(collection.ID, searchIndex, collectionVersion); err != nil {
		return err
	}

	if !skipOptimize {
		log.Println("start optimizing")
		if err = target.Optimize(); err != nil {
			return err
		}
		log.Println("completed optimizing")
	}
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
