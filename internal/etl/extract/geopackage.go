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
		    st_minx(castautomagic(geom)) as bbox_minx, 
		    st_miny(castautomagic(geom)) as bbox_miny, 
		    st_maxx(castautomagic(geom)) as bbox_maxx, 
		    st_maxy(castautomagic(geom)) as bbox_maxy,
		    st_geometrytype(castautomagic(geom)) as gt, -- alternatively can also be read from gpkg_geometry_columns
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
		if len(row) != 6+len(fields) {
			return nil, fmt.Errorf("unexpected row length (%v)", len(row))
		}
		result = append(result, mapRowToRawRecord(row, fields))
	}
	return result, nil
}

func mapRowToRawRecord(row []any, fields []string) t.RawRecord {
	bbox := row[1:5]

	return t.RawRecord{
		FeatureID: row[0].(int64),
		Bbox: &geom.Extent{
			bbox[0].(float64),
			bbox[1].(float64),
			bbox[2].(float64),
			bbox[3].(float64),
		},
		GeometryType: row[5].(string),
		FieldValues:  row[6 : 6+len(fields)],
	}
}
