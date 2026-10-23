package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const defaultOutput = "examples/resources/records.gpkg"

func main() {
	output := flag.String("output", defaultOutput, "path for the generated GeoPackage")
	force := flag.Bool("force", false, "replace the output file if it already exists")
	flag.Parse()

	if err := build(*output, *force); err != nil {
		log.Fatal(err)
	}
}

func build(output string, force bool) (err error) {
	_, statErr := os.Stat(output)
	if statErr == nil && !force {
		return fmt.Errorf("%s already exists; pass -force to replace it", output)
	}
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("check output %s: %w", output, statErr)
	}
	if statErr == nil {
		if err := os.Remove(output); err != nil {
			return fmt.Errorf("remove existing output %s: %w", output, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	db, err := sql.Open("sqlite3", output)
	if err != nil {
		return fmt.Errorf("open output GeoPackage: %w", err)
	}
	db.SetMaxOpenConns(1)
	defer func() {
		if closeErr := db.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close output GeoPackage: %w", closeErr)
		}
	}()

	for _, path := range []string{
		"internal/ogc/records/schema/geopackage-core.sql",
		"internal/ogc/records/schema/geopackage.sql",
		"internal/ogc/records/schema/example-data.sql",
	} {
		sqlText, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if _, err := db.Exec(string(sqlText)); err != nil {
			return fmt.Errorf("execute %s: %w", path, err)
		}
	}

	var applicationID, userVersion int
	if err := db.QueryRow("PRAGMA application_id").Scan(&applicationID); err != nil {
		return fmt.Errorf("verify GeoPackage application_id: %w", err)
	}
	if err := db.QueryRow("PRAGMA user_version").Scan(&userVersion); err != nil {
		return fmt.Errorf("verify GeoPackage user_version: %w", err)
	}
	if applicationID != 1196444487 || userVersion != 10300 {
		return fmt.Errorf("invalid GeoPackage identifiers: application_id=%d user_version=%d", applicationID, userVersion)
	}

	var catalogs, datasets, distributions int
	if err := db.QueryRow("SELECT COUNT(*) FROM dcat_catalog").Scan(&catalogs); err != nil {
		return fmt.Errorf("verify catalog rows: %w", err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM dcat_dataset").Scan(&datasets); err != nil {
		return fmt.Errorf("verify dataset rows: %w", err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM dcat_distribution").Scan(&distributions); err != nil {
		return fmt.Errorf("verify distribution rows: %w", err)
	}
	if catalogs != 1 || datasets != 1 || distributions != 1 {
		return fmt.Errorf("unexpected fixture rows: catalogs=%d datasets=%d distributions=%d", catalogs, datasets, distributions)
	}
	log.Printf("generated %s with one catalog, dataset, and distribution", output)
	return nil
}