package extract

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/PDOK/gomagpie/config"
	t "github.com/PDOK/gomagpie/internal/etl/transform"
	"github.com/go-spatial/geom"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
)

const (
	sqliteDriverName = "sqlite3_with_extensions"

	// fid,minx,miny,maxx,maxy,geom_type
	nrOfStandardFieldsInQuery = 6
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

func (g *GeoPackage) Close() {
	_ = g.db.Close()
}

func (g *GeoPackage) Extract(table config.FeatureTable, fields []string, limit int, offset int) ([]t.RawRecord, error) {
	if len(fields) == 0 {
		return nil, errors.New("no fields provided to read from GeoPackage")
	}

	// TODO we might need WGS84 transformation here of bbox
	query := fmt.Sprintf(`
		select %[3]s as fid,
		    st_minx(castautomagic(%[4]s)) as bbox_minx, 
		    st_miny(castautomagic(%[4]s)) as bbox_miny, 
		    st_maxx(castautomagic(%[4]s)) as bbox_maxx, 
		    st_maxy(castautomagic(%[4]s)) as bbox_maxy,
		    st_geometrytype(castautomagic(%[4]s)) as geom_type,
		    %[1]s -- all feature specific fields
		from %[2]s 
		limit :limit 
		offset :offset`, strings.Join(fields, ","), table.Name, table.FID, table.Geom)

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
		if len(row) != len(fields)+nrOfStandardFieldsInQuery {
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
		FieldValues:  row[nrOfStandardFieldsInQuery : nrOfStandardFieldsInQuery+len(fields)],
	}
}
