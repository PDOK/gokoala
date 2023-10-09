package geopackage

import (
	"fmt"
	"log"
	"os"

	cloudsqlitevfs "github.com/PDOK/go-cloud-sqlite-vfs"
	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/jmoiron/sqlx"
)

const (
	sqliteDriverName = "sqlite3"
	vfsName          = "cloudbackedvfs"
)

type Genre struct {
	ID   string `db:"GenreId"`
	Name string `db:"Name"`
}

type GeoPackage struct {
	db       *sqlx.DB
	cloudVFS *cloudsqlitevfs.VFS
}

func NewGeoPackage(e *engine.Engine) *GeoPackage {
	gpkg := e.Config.OgcAPI.Features.Datasource.GeoPackage
	if gpkg.Local != nil {
		return newLocalGeoPackage(gpkg.Local)
	} else if gpkg.Cloud != nil {
		return newCloudBackedGeoPackage(gpkg.Cloud)
	}
	log.Fatal("unknown geopackage config encountered")
	return nil
}

func newLocalGeoPackage(gpkg *engine.GeoPackageLocal) *GeoPackage {
	log.Printf("connecting to local GeoPackage: %s", gpkg.File)
	db, err := sqlx.Open(sqliteDriverName, gpkg.File)
	if err != nil {
		log.Fatalf("failed to open GeoPackage: %v", err)
	}
	log.Printf("connected to local GeoPackage: %s", gpkg.File)
	return &GeoPackage{
		db,
		nil,
	}
}

func newCloudBackedGeoPackage(gpkg *engine.GeoPackageCloud) *GeoPackage {
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

	return &GeoPackage{
		db,
		&vfs,
	}
}

func (g *GeoPackage) Close() {
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

func (g *GeoPackage) GetFeatures(collection string, cursor int64, limit int) (*domain.FeatureCollection, domain.Cursor) {
	rows, err := g.db.Queryx("SELECT * FROM genres")
	if err != nil {
		log.Println(err)
	} else {
		genre := Genre{}
		for rows.Next() {
			err := rows.StructScan(&genre)
			if err != nil {
				log.Println(err)
			}
			log.Printf("%#v\n", genre)
		}
	}
	defer rows.Close()

	// TODO: not implemented yet
	log.Printf("TODO: return data from gpkg for collection %s using cursor %d with limt %d",
		collection, cursor, limit)
	return nil, domain.Cursor{}
}

func (g *GeoPackage) GetFeature(collection string, featureID string) *domain.Feature {
	// TODO: not implemented yet
	log.Printf("TODO: return feature %s from gpkg in collection %s", featureID, collection)
	return nil
}
