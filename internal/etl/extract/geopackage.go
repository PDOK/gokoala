package extract

import (
	"database/sql"
	"os"
	"path"
	"sync"

	t "github.com/PDOK/gomagpie/internal/etl/transform"
	"github.com/mattn/go-sqlite3"
)

const (
	sqliteDriverName = "sqlite3_with_extensions"
)

var once sync.Once

// Load sqlite (with extensions) once.
//
// Extensions are by default expected in /usr/lib. For spatialite you can
// alternatively/optionally set SPATIALITE_LIBRARY_PATH.
func loadDriver() {
	once.Do(func() {
		spatialite := path.Join(os.Getenv("SPATIALITE_LIBRARY_PATH"), "mod_spatialite")
		driver := &sqlite3.SQLiteDriver{Extensions: []string{spatialite}}
		sql.Register(sqliteDriverName, driver)
	})
}

type GeoPackage struct {
}

func NewGeoPackage(path string) *GeoPackage {
	loadDriver()
	g := &GeoPackage{}
	return g
}

func (g *GeoPackage) Extract(featureTable string, fields []string, lastOffset int) ([]t.RawRecord, int, error) {
	return []t.RawRecord{}, lastOffset, nil
}
