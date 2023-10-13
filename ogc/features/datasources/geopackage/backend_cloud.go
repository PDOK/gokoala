//go:build cgo && !darwin

package geopackage

import (
	"fmt"
	"log"
	"os"

	cloudsqlitevfs "github.com/PDOK/go-cloud-sqlite-vfs"
	"github.com/PDOK/gokoala/engine"
	"github.com/jmoiron/sqlx"
)

const vfsName = "cloudbackedvfs"

// Cloud-Backed SQLite (CBS) GeoPackage in Azure or Google object storage
type cloudGeoPackage struct {
	db       *sqlx.DB
	cloudVFS *cloudsqlitevfs.VFS
}

func newCloudBackedGeoPackage(gpkg *engine.GeoPackageCloud) geoPackageBackend {
	cacheDir := os.TempDir()
	if gpkg.Cache != nil {
		cacheDir = *gpkg.Cache
	}

	log.Printf("connecting to Cloud-Backed GeoPackage: %s\n", gpkg.Connection)
	vfs, err := cloudsqlitevfs.NewVFS(vfsName, gpkg.Connection, gpkg.User, gpkg.Auth, gpkg.Container, cacheDir)
	if err != nil {
		log.Fatalf("failed to connect with Cloud-Backed GeoPackage: %v", err)
	}
	log.Printf("connected to Cloud-Backed GeoPackage: %s\n", gpkg.Connection)

	db, err := sqlx.Open(sqliteDriverName, fmt.Sprintf("/%s/%s?vfs=%s", gpkg.Container, gpkg.File, vfsName))
	if err != nil {
		log.Fatalf("failed to open Cloud-Backed GeoPackage: %v", err)
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
