package extract

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	t "github.com/PDOK/gomagpie/internal/etl/transform"
	"github.com/go-spatial/geom"
	"github.com/jmoiron/sqlx"
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
	db *sqlx.DB
}

func NewGeoPackage(path string) (*GeoPackage, error) {
	loadDriver()

	conn := fmt.Sprintf("file:%s?mode=ro", path)
	db, err := sqlx.Open(sqliteDriverName, conn)
	if err != nil {
		return nil, err
	}
	return &GeoPackage{db}, nil
}

func (g *GeoPackage) Close() error {
	return g.db.Close()
}

// TODO: configure fid column, get geom column, prepare
func (g *GeoPackage) Extract(featureTable string, fields []string, limit int, offset int) ([]t.RawRecord, error) {
	query := fmt.Sprintf(`
		select fid,
		    st_minx(castautomagic(geom)) as minx, 
		    st_miny(castautomagic(geom)) as miny, 
		    st_maxx(castautomagic(geom)) as maxx, 
		    st_maxy(castautomagic(geom)) as maxy,
		    %s -- all feature specific fields
		from %s 
		limit :limit 
		offset :offset`, strings.Join(fields, ","), featureTable)

	rows, err := g.db.NamedQuery(query, map[string]any{"limit": limit, "offset": offset})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []t.RawRecord
	for rows.Next() {
		var row []any
		if row, err = rows.SliceScan(); err != nil {
			return nil, err
		}
		result = append(result, mapRowToRawRecord(row, fields))
	}
	return result, nil
}

func mapRowToRawRecord(row []any, fields []string) t.RawRecord {
	bbox := row[1:5]
	record := t.RawRecord{
		FeatureID: row[0].(int64),
		Bbox: geom.Extent{
			bbox[0].(float64),
			bbox[1].(float64),
			bbox[2].(float64),
			bbox[3].(float64),
		},
		FieldValues: row[5 : 5+len(fields)],
	}
	return record
}
