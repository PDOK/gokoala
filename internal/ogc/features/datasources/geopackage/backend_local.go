package geopackage

import (
	"fmt"
	"log"

	"github.com/PDOK/gokoala/config"
	"github.com/jmoiron/sqlx"
)

// GeoPackage on local disk
type localGeoPackage struct {
	db *sqlx.DB
}

func newLocalGeoPackage(gpkg *config.GeoPackageLocal) geoPackageBackend {
	inMemCacheSize, err := gpkg.InMemoryCacheSizeSqlite()
	if err != nil {
		log.Fatalf("invalid in-memory cache size provided, error: %v", err)
	}
	conn := fmt.Sprintf("file:%s?mode=ro&_cache_size=%d", gpkg.File, inMemCacheSize)
	db, err := sqlx.Open(sqliteDriverName, conn)
	if err != nil {
		log.Fatalf("failed to open GeoPackage: %v", err)
	}
	log.Printf("connected to local GeoPackage: %s", gpkg.File)

	return &localGeoPackage{db}
}

func (g *localGeoPackage) getDB() *sqlx.DB {
	return g.db
}

func (g *localGeoPackage) close() {
	err := g.db.Close()
	if err != nil {
		log.Printf("failed to close GeoPackage: %v", err)
	}
}
