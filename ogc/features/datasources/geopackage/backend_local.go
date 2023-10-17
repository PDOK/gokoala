package geopackage

import (
	"log"

	"github.com/PDOK/gokoala/engine"
	"github.com/jmoiron/sqlx"
)

// GeoPackage on local disk
type localGeoPackage struct {
	db *sqlx.DB
}

func newLocalGeoPackage(gpkg *engine.GeoPackageLocal) geoPackageBackend {
	db, err := sqlx.Open(sqliteDriverName, gpkg.File)
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
