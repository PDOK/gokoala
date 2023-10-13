package geopackage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/go-spatial/geom/encoding/geojson"
	"github.com/go-spatial/geom/encoding/gpkg"
	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3" // import for side effect (= sqlite3 driver) only
)

const (
	sqliteDriverName = "sqlite3"
	queryGpkgContent = `select
			c.table_name, c.data_type, c.identifier, c.description, c.last_change,
			c.min_x, c.min_y, c.max_x, c.max_y, c.srs_id, gc.column_name, gc.geometry_type_name
		from
			gpkg_contents c join gpkg_geometry_columns gc on c.table_name == gc.table_name
		where
			c.data_type = 'features' and 
			c.min_x is not null`
)

type geoPackageBackend interface {
	getDB() *sqlx.DB
	close()
}

type gpkgFeatureTable struct {
	TableName          string    `db:"table_name"`
	DataType           string    `db:"data_type"`
	Identifier         string    `db:"identifier"`
	Description        string    `db:"description"`
	GeometryColumnName string    `db:"column_name"`
	GeometryType       string    `db:"geometry_type_name"`
	LastChange         time.Time `db:"last_change"`
	MinX               float64   `db:"min_x"` // bbox
	MinY               float64   `db:"min_y"` // bbox
	MaxX               float64   `db:"max_x"` // bbox
	MaxY               float64   `db:"max_y"` // bbox
	SRS                int64     `db:"srs_id"`
}

type GeoPackage struct {
	backend geoPackageBackend

	fidColumn        string
	featureTableByID map[string]*gpkgFeatureTable
	queryTimeout     time.Duration
}

func NewGeoPackage(e *engine.Engine) *GeoPackage {
	gpkgConfig := e.Config.OgcAPI.Features.Datasource.GeoPackage

	g := &GeoPackage{}
	switch {
	case gpkgConfig.Local != nil:
		g.backend = newLocalGeoPackage(gpkgConfig.Local)
		g.fidColumn = gpkgConfig.Local.Fid
		g.queryTimeout = gpkgConfig.Local.GetQueryTimeout()
	case gpkgConfig.Cloud != nil:
		g.backend = newCloudBackedGeoPackage(gpkgConfig.Cloud)
		g.fidColumn = gpkgConfig.Cloud.Fid
		g.queryTimeout = gpkgConfig.Cloud.GetQueryTimeout()
	default:
		log.Fatal("unknown geopackage config encountered")
	}

	featureTables, err := readGpkgContents(g.backend.getDB())
	if err != nil {
		log.Fatal(err)
	}
	g.featureTableByID = featureTables

	return g
}

func (g *GeoPackage) Close() {
	g.backend.close()
}

func (g *GeoPackage) GetFeatures(ctx context.Context, collection string, cursor int64, limit int) (*domain.FeatureCollection, domain.Cursor, error) {
	featureTable, ok := g.featureTableByID[collection]
	if !ok {
		return nil, domain.Cursor{}, fmt.Errorf("can't query collection '%s' since it doesn't exist in "+
			"geopackage, available in geopackage: %v", collection, engine.Keys(g.featureTableByID))
	}

	query := fmt.Sprintf("select * from %s f where f.%s > ? order by f.%s limit ?",
		featureTable.TableName, g.fidColumn, g.fidColumn)

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	rows, err := g.backend.getDB().QueryxContext(queryCtx, query, cursor, limit)
	if err != nil {
		return nil, domain.Cursor{}, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	result := domain.FeatureCollection{}
	result.Features, err = g.mapRowsToFeatures(rows, g.fidColumn, featureTable.GeometryColumnName)
	if err != nil {
		return nil, domain.Cursor{}, err
	}

	result.NumberReturned = len(result.Features)
	last := result.NumberReturned < limit // we could make this more reliable (by querying one record more), but sufficient for now

	return &result, domain.NewCursor(result.Features, limit, last), nil
}

func (g *GeoPackage) GetFeature(ctx context.Context, collection string, featureID int64) (*domain.Feature, error) {
	gpkgContent, ok := g.featureTableByID[collection]
	if !ok {
		return nil, fmt.Errorf("can't query collection '%s' since it doesn't exist in "+
			"geopackage, available in geopackage: %v", collection, engine.Keys(g.featureTableByID))
	}

	query := fmt.Sprintf("select * from %s f where f.%s = ? limit 1", gpkgContent.TableName, g.fidColumn)

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	rows, err := g.backend.getDB().QueryxContext(queryCtx, query, featureID)
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	features, err := g.mapRowsToFeatures(rows, g.fidColumn, gpkgContent.GeometryColumnName)
	if err != nil {
		return nil, err
	}
	if len(features) != 1 {
		return nil, nil //nolint:nilnil
	}
	return features[0], nil
}

func readGpkgContents(db *sqlx.DB) (map[string]*gpkgFeatureTable, error) {
	rows, err := db.Queryx(queryGpkgContent)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve gpkg_contents using query: %v\n, error: %w", queryGpkgContent, err)
	}
	defer rows.Close()

	result := make(map[string]*gpkgFeatureTable, 10)
	for rows.Next() {
		row := gpkgFeatureTable{}
		if err = rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("failed to read gpkg_contents record, error: %w", err)
		}
		if row.TableName == "" {
			return nil, fmt.Errorf("feature table name is blank, error: %w", err)
		}
		result[row.Identifier] = &row
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no records found in gpkg_contents, can't serve features")
	}

	return result, nil
}

func (g *GeoPackage) mapRowsToFeatures(rows *sqlx.Rows, fidColumn string, geomColumn string) ([]*domain.Feature, error) {
	result := make([]*domain.Feature, 0)
	cols, err := rows.Columns()
	if err != nil {
		return result, err
	}

	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := 0; i < len(cols); i++ {
			valPtrs[i] = &vals[i]
		}
		if err = rows.Scan(valPtrs...); err != nil {
			return result, err
		}

		feature := &domain.Feature{Feature: geojson.Feature{Properties: make(map[string]interface{})}}

		if err = g.mapColumnsToFeature(feature, cols, vals, fidColumn, geomColumn); err != nil {
			return result, err
		}
		result = append(result, feature)
	}
	return result, nil
}

//nolint:cyclop
func (g *GeoPackage) mapColumnsToFeature(feature *domain.Feature, cols []string, vals []interface{}, fidColumn string, geomColumn string) error {
	for i, colName := range cols {
		columnValue := vals[i]
		if columnValue == nil {
			continue
		}

		switch colName {
		case fidColumn:
			feature.ID = columnValue.(int64)

		case geomColumn:
			rawGeom, ok := columnValue.([]byte)
			if !ok {
				return fmt.Errorf("failed to read geometry from %s column in geopackage", geomColumn)
			}
			geom, err := gpkg.DecodeGeometry(rawGeom)
			if err != nil {
				return fmt.Errorf("failed to decode geometry from geopackage: %w", err)
			}
			feature.Geometry = geojson.Geometry{Geometry: geom.Geometry}

		case "minx", "miny", "maxx", "maxy", "min_zoom", "max_zoom":
			// Skip these columns used for bounding box and zoom filtering
			continue

		default:
			// Grab any non-nil, non-id, non-bounding box, & non-geometry column as a tag
			switch v := columnValue.(type) {
			case []uint8:
				asBytes := make([]byte, len(v))
				copy(asBytes, v)
				feature.Properties[colName] = string(asBytes)
			case int64:
				feature.Properties[colName] = v
			case float64:
				feature.Properties[colName] = v
			case time.Time:
				feature.Properties[colName] = v
			case string:
				feature.Properties[colName] = v
			case bool:
				feature.Properties[colName] = v
			default:
				return fmt.Errorf("unexpected type for sqlite column data: %v: %T", cols[i], v)
			}
		}
	}
	return nil
}
