package geopackage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/features/datasources"
	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/go-spatial/geom"
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

type featureTable struct {
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

	fidColumn                  string
	featureTableByCollectionID map[string]*featureTable
	queryTimeout               time.Duration
}

func NewGeoPackage(collections engine.GeoSpatialCollections, gpkgConfig engine.GeoPackage) *GeoPackage {
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

	featureTables, err := readGpkgContents(collections, g.backend.getDB())
	if err != nil {
		log.Fatal(err)
	}
	g.featureTableByCollectionID = featureTables

	return g
}

func (g *GeoPackage) Close() {
	g.backend.close()
}

func (g *GeoPackage) GetFeatures(ctx context.Context, collection string, options datasources.FeatureOptions) (*domain.FeatureCollection, domain.Cursor, error) {
	table, ok := g.featureTableByCollectionID[collection]
	if !ok {
		return nil, domain.Cursor{}, fmt.Errorf("can't query collection '%s' since it doesn't exist in "+
			"geopackage, available in geopackage: %v", collection, engine.Keys(g.featureTableByCollectionID))
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	query, queryArgs := g.makeFeaturesQuery(table, options)
	stmt, err := g.backend.getDB().PreparexContext(ctx, query)
	if err != nil {
		return nil, domain.Cursor{}, err
	}
	defer stmt.Close()

	rows, err := g.backend.getDB().QueryxContext(queryCtx, query, queryArgs...)
	if err != nil {
		return nil, domain.Cursor{}, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	var nextPrev *domain.NextPrevID
	result := domain.FeatureCollection{}
	result.Features, nextPrev, err = domain.MapRowsToFeatures(rows, g.fidColumn, table.GeometryColumnName, readGpkgGeometry)
	if err != nil {
		return nil, domain.Cursor{}, err
	}
	if nextPrev == nil {
		return nil, domain.Cursor{}, fmt.Errorf("failed to get prev/next cursor")
	}

	result.NumberReturned = len(result.Features)
	return &result, domain.NewCursor(result.Features, *nextPrev), nil
}

func (g *GeoPackage) GetFeature(ctx context.Context, collection string, featureID int64) (*domain.Feature, error) {
	table, ok := g.featureTableByCollectionID[collection]
	if !ok {
		return nil, fmt.Errorf("can't query collection '%s' since it doesn't exist in "+
			"geopackage, available in geopackage: %v", collection, engine.Keys(g.featureTableByCollectionID))
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	query := fmt.Sprintf("select * from %s f where f.%s = ? limit 1", table.TableName, g.fidColumn)
	stmt, err := g.backend.getDB().PreparexContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(queryCtx, featureID)
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	features, _, err := domain.MapRowsToFeatures(rows, g.fidColumn, table.GeometryColumnName, readGpkgGeometry)
	if err != nil {
		return nil, err
	}
	if len(features) != 1 {
		return nil, nil //nolint:nilnil
	}
	return features[0], nil
}

func (g *GeoPackage) makeFeaturesQuery(table *featureTable, opt datasources.FeatureOptions) (string, []any) {
	// we don't yet support extra filters but when we do (as in the case of bbox or part3 filtering)
	// they need to be included in the next/prev handling as well
	extraFilters := ""

	nextPrevCTE := fmt.Sprintf(`
next as (select * from %[1]s where %[2]s >= %[3]d %[5]s order by %[2]s asc limit %[4]d + 1),
prev as (select * from %[1]s where %[2]s < %[3]d %[5]s order by %[2]s desc limit %[4]d),
nextprev as (select * from next union all select * from prev),
featuretable as (select *, lag(%[2]s, %[4]d) over (order by %[2]s) as prevcursor, lead(%[2]s, %[4]d) over (order by %[2]s) as nextcursor from nextprev)
`, table.TableName, g.fidColumn, opt.Cursor, opt.Limit, extraFilters)

	if opt.Bbox != nil {
		// TODO create bbox query
		bboxQuery := fmt.Sprintf(`with %s <bbox query here>`, nextPrevCTE)
		return bboxQuery, []any{opt.Cursor, opt.Limit, opt.Bbox}
	}
	if opt.Filter != "" {
		// TODO create part3 filter query
		filterQuery := fmt.Sprintf(`with %s <filter query here>`, nextPrevCTE)
		return filterQuery, []any{opt.Cursor, opt.Limit, opt.Filter}
	}
	defaultQuery := fmt.Sprintf(`with %s select * from featuretable where %s >= ? limit ?`, nextPrevCTE, g.fidColumn)
	return defaultQuery, []any{opt.Cursor, opt.Limit}
}

// Read gpkg_contents table. This table contains metadata about feature tables. The result is a mapping from
// collection ID -> feature table metadata. We match each feature table to the collection ID by looking at the
// 'identifier' column. Also in case there's no exact match between 'collection ID' and 'identifier' we use
// the explicitly configured 'datasource ID'
func readGpkgContents(collections engine.GeoSpatialCollections, db *sqlx.DB) (map[string]*featureTable, error) {
	rows, err := db.Queryx(queryGpkgContent)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve gpkg_contents using query: %v\n, error: %w", queryGpkgContent, err)
	}
	defer rows.Close()

	result := make(map[string]*featureTable, 10)
	for rows.Next() {
		row := featureTable{}
		if err = rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("failed to read gpkg_contents record, error: %w", err)
		}
		if row.TableName == "" {
			return nil, fmt.Errorf("feature table name is blank, error: %w", err)
		}

		if len(collections) == 0 {
			result[row.Identifier] = &row
		} else {
			for _, collection := range collections {
				if row.Identifier == collection.ID {
					result[collection.ID] = &row
					break
				} else if collection.Features != nil && collection.Features.DatasourceID != nil &&
					row.Identifier == *collection.Features.DatasourceID {
					result[collection.ID] = &row
					break
				}
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no records found in gpkg_contents, can't serve features")
	}

	return result, nil
}

func readGpkgGeometry(rawGeom []byte) (geom.Geometry, error) {
	geometry, err := gpkg.DecodeGeometry(rawGeom)
	if err != nil {
		return nil, err
	}
	return geometry.Geometry, nil
}
