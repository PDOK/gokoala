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
	if gpkg.Download != nil {
		downloadGeoPackage(gpkg)
	}
	conn := fmt.Sprintf("file:%s?mode=ro&_cache_size=%d", gpkg.File, gpkg.InMemoryCacheSize)
	db, err := sqlx.Open(sqliteDriverName, conn)
	if err != nil {
		log.Fatalf("failed to open GeoPackage: %v", err)
	}
	log.Printf("connected to local GeoPackage: %s", gpkg.File)

	return &localGeoPackage{db}
}

func downloadGeoPackage(gpkg *config.GeoPackageLocal) {
	url := *gpkg.Download.From.URL
	log.Printf("start download of GeoPackage: %s", url.String())
	downloadTime, err := engine.Download(url, gpkg.File, gpkg.Download.Parallelism, gpkg.Download.TLSSkipVerify,
		gpkg.Download.Timeout.Duration, gpkg.Download.RetryDelay.Duration, gpkg.Download.RetryMaxDelay.Duration, gpkg.Download.MaxRetries)
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
