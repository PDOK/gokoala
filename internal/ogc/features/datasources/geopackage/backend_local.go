package geopackage

import (
	"fmt"
	"log"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/jmoiron/sqlx"
)

// GeoPackage on local disk
type localGeoPackage struct {
	db *sqlx.DB
}

func newLocalGeoPackage(gpkg *config.GeoPackageLocal) geoPackageBackend {
	if gpkg.Init != nil {
		downloadGeoPackage(gpkg)
	}
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

func downloadGeoPackage(gpkg *config.GeoPackageLocal) {
	url := *gpkg.Init.Download.URL
	log.Printf("start download of GeoPackage: %s", url.String())
	downloadTime, err := engine.Download(url, gpkg.File, gpkg.Init.Parallelism, gpkg.Init.TLSSkipVerify,
		gpkg.Init.RetryDelay.Duration, gpkg.Init.RetryMaxDelay.Duration, gpkg.Init.MaxRetries)
	if err != nil {
		log.Fatalf("failed to download GeoPackage: %v", err)
	}
	log.Printf("succesfully downloaded GeoPackage to %s in %s", gpkg.File, downloadTime.Round(time.Second))
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
