package geopackage

import (
	"fmt"
	"log"
	"os"
	"time"

	cloudsqlitevfs "github.com/PDOK/go-cloud-sqlite-vfs"
	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/jmoiron/sqlx"
)

const (
	sqliteDriverName = "sqlite3"
	vfsName          = "cloudbackedvfs"
	queryGpkgContent = `select
							c.table_name, c.data_type, c.identifier, c.description, c.last_change,
							c.min_x, c.min_y, c.max_x, c.max_y, c.srs_id, gc.column_name, gc.geometry_type_name
						from
							gpkg_contents c join gpkg_geometry_columns gc on c.table_name == gc.table_name
						where
							c.data_type = 'features' and 
							c.min_x is not null`
)

type Genre struct { // FIXME remove me once features are implemented
	ID   string `db:"GenreId"`
	Name string `db:"Name"`
}

type GpkgContent struct {
	TableName          string    `db:"table_name"`
	DataType           string    `db:"data_type"`
	Identifier         string    `db:"identifier"`
	Description        string    `db:"description"`
	GeometryColumnName string    `db:"column_name"`
	GeometryType       string    `db:"geometry_type_name"`
	LastChange         time.Time `db:"last_change"`
	MinX               float64   `db:"min_x"`  // bbox
	MinY               float64   `db:"min_y"`  // bbox
	MaxX               float64   `db:"max_x"`  // bbox
	MaxY               float64   `db:"max_y"`  // bbox
	SrsId              int64     `db:"srs_id"` //nolint:revive

	Columns *[]GpkgColumn
}

type GpkgColumn struct {
	Cid          int    `db:"cid"`
	Name         string `db:"name"`
	DataType     string `db:"type"`
	NotNull      int    `db:"notnull"`
	DefaultValue int    `db:"dflt_value"`
	PrimaryKey   int    `db:"pk"`
}

type GeoPackage struct {
	db              *sqlx.DB
	cloudVFS        *cloudsqlitevfs.VFS
	gpkgContentByID map[string]*GpkgContent
}

func NewGeoPackage(e *engine.Engine) *GeoPackage {
	gpkgConfig := e.Config.OgcAPI.Features.Datasource.GeoPackage

	var gpkg *GeoPackage
	if gpkgConfig.Local != nil {
		gpkg = newLocalGeoPackage(gpkgConfig.Local)
	} else if gpkgConfig.Cloud != nil {
		gpkg = newCloudBackedGeoPackage(gpkgConfig.Cloud)
	} else {
		log.Fatal("unknown geopackage config encountered")
	}

	content, err := readGpkgContents(gpkg.db)
	if err != nil {
		log.Fatal(err)
	}
	gpkg.gpkgContentByID = content
	return gpkg
}

func newLocalGeoPackage(gpkg *engine.GeoPackageLocal) *GeoPackage {
	if _, err := os.Stat(gpkg.File); os.IsNotExist(err) {
		log.Fatalf("failed to locate GeoPackage: %s", gpkg.File)
	}

	db, err := sqlx.Open(sqliteDriverName, gpkg.File)
	if err != nil {
		log.Fatalf("failed to open GeoPackage: %v", err)
	}
	log.Printf("connected to local GeoPackage: %s", gpkg.File)

	return &GeoPackage{db, nil, nil}
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

	return &GeoPackage{db, &vfs, nil}
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
	result := domain.FeatureCollection{}

	gpkgContent, ok := g.gpkgContentByID[collection]
	if !ok {
		log.Printf("can't query collection '%s' since it doesn't exist in geopackage, "+
			"available in geopackage: %v\n", collection, engine.Keys(g.gpkgContentByID))
		return nil, domain.Cursor{}
	}

	log.Print(gpkgContent)
	query := ""

	rows, err := g.db.Queryx(query)
	if err != nil {
		log.Printf("failed to query features using query: %v\n, error: %v", query, err)
	}
	defer rows.Close()

	return &result, domain.Cursor{}
}

func (g *GeoPackage) GetFeature(collection string, featureID string) *domain.Feature {
	// TODO: not implemented yet
	log.Printf("TODO: return feature %s from gpkg in collection %s", featureID, collection)
	return nil
}

func readGpkgContents(db *sqlx.DB) (map[string]*GpkgContent, error) {
	rows, err := db.Queryx(queryGpkgContent)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve gpkg_contents using query: %v\n, error: %w", queryGpkgContent, err)
	}
	defer rows.Close()

	result := make(map[string]*GpkgContent, 10)
	for rows.Next() {
		// read a gpkg_contents record
		row := GpkgContent{}
		if err = rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("failed to read gpkg_contents record, error: %w", err)
		}

		// read metadata of table mentioned in gpkg_contents record
		var columns []GpkgColumn
		if err = db.Select(&columns, `pragma table_info("?");`, row.TableName); err != nil {
			return nil, fmt.Errorf("failed to read columns of table %s, error: %w", row.TableName, err)
		}
		row.Columns = &columns

		result[row.Identifier] = &row
	}

	if err := rows.Err(); err != nil {
		log.Panic(err)
	}
	if len(result) == 0 {
		log.Panic("no records found in gpkg_contents, can't serve features")
	}

	return result, nil
}
