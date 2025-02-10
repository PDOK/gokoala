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
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	"github.com/twpayne/go-geom"
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

func (g *GeoPackage) Extract(table config.FeatureTable, fields []string, externalFidFields []string, where string, limit int, offset int) ([]t.RawRecord, error) {
	if len(fields) == 0 {
		return nil, errors.New("no fields provided to read from GeoPackage")
	}
	if where != "" {
		where = "where " + where
	}

	// combine field and externalFidFields
	extraFields := fields
	extraFields = append(extraFields, externalFidFields...)

	// TODO we might need WGS84 transformation here of bbox
	query := fmt.Sprintf(`
		select %[3]s as fid,
		    st_minx(castautomagic(%[4]s)) as bbox_minx,
		    st_miny(castautomagic(%[4]s)) as bbox_miny,
		    st_maxx(castautomagic(%[4]s)) as bbox_maxx,
		    st_maxy(castautomagic(%[4]s)) as bbox_maxy,
		    st_geometrytype(castautomagic(%[4]s)) as geom_type,
		    %[1]s -- all feature specific fields and any fields for external_fid
		from %[2]s
		%[5]s
		limit :limit
		offset :offset`, strings.Join(extraFields, ","), table.Name, table.FID, table.Geom, where)

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
		if len(row) != len(fields)+len(externalFidFields)+nrOfStandardFieldsInQuery {
			return nil, fmt.Errorf("unexpected row length (%v)", len(row))
		}
		record, err := mapRowToRawRecord(row, fields, externalFidFields)
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}

func mapRowToRawRecord(row []any, fields []string, externalFidFields []string) (t.RawRecord, error) {
	bbox := row[1:5]

	fid := row[0].(int64)
	if fid <= 0 {
		return t.RawRecord{}, errors.New("encountered negative fid")
	}
	geomType := row[5].(string)
	if geomType == "" {
		return t.RawRecord{}, fmt.Errorf("encountered empty geometry type for fid %d", fid)
	}
	return t.RawRecord{
		FeatureID: fid,
		Bbox: geom.NewBounds(geom.XY).Set(
			bbox[0].(float64),
			bbox[1].(float64),
			bbox[2].(float64),
			bbox[3].(float64),
		),
		GeometryType:      geomType,
		FieldValues:       row[nrOfStandardFieldsInQuery : nrOfStandardFieldsInQuery+len(fields)],
		ExternalFidValues: row[nrOfStandardFieldsInQuery+len(fields) : nrOfStandardFieldsInQuery+len(fields)+len(externalFidFields)],
	}, nil
}
