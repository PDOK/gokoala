//go:build cgo && !darwin && !windows

package geopackage

import (
	"fmt"
	"log"

	"github.com/PDOK/gokoala/config"

	cloudsqlitevfs "github.com/PDOK/go-cloud-sqlite-vfs"
	"github.com/jmoiron/sqlx"
)

const (
	vfsName = "cloudbackedvfs"
)

// Cloud-Backed SQLite (CBS) GeoPackage in Azure or Google object storage
type cloudGeoPackage struct {
	db       *sqlx.DB
	cloudVFS *cloudsqlitevfs.VFS
}

func newCloudBackedGeoPackage(gpkg *config.GeoPackageCloud) geoPackageBackend {
	cacheDir, err := gpkg.CacheDir()
	if err != nil {
		log.Fatalf("invalid cache dir, error: %v", err)
	}
	cacheSize, err := gpkg.Cache.MaxSizeAsBytes()
	if err != nil {
		log.Fatalf("invalid cache size provided, error: %v", err)
	}

	msg := fmt.Sprintf("Cloud-Backed GeoPackage '%s' in container '%s' on '%s'",
		gpkg.File, gpkg.Container, gpkg.Connection)

	log.Printf("connecting to %s\n", msg)
	vfs, err := cloudsqlitevfs.NewVFS(vfsName, gpkg.Connection, gpkg.User, gpkg.Auth,
		gpkg.Container, cacheDir, cacheSize, gpkg.LogHTTPRequests)
	if err != nil {
		log.Fatalf("failed to connect with %s, error: %v", msg, err)
	}
	log.Printf("connected to %s\n", msg)

	db, err := sqlx.Open(sqliteDriverName, fmt.Sprintf("/%s/%s?vfs=%s&mode=ro", gpkg.Container, gpkg.File, vfsName))
	if err != nil {
		log.Fatalf("failed to open %s, error: %v", msg, err)
	}

	return &cloudGeoPackage{db, &vfs}
}

func (g *cloudGeoPackage) getDB() *sqlx.DB {
	return g.db
}

func (g *cloudGeoPackage) close() {
	err := g.db.Close()
	if err != nil {
		log.Printf("failed to close GeoPackage: %v", err)
	}
	if g.cloudVFS != nil {
		err = g.cloudVFS.Close()
		if err != nil {
			log.Printf("failed to close Cloud-Backed GeoPackage: %v", err)
		}
	}
}
