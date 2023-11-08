package geopackage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/engine/util"
	"github.com/PDOK/gokoala/ogc/features/datasources"
	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/gpkg"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"

	_ "github.com/mattn/go-sqlite3" // import for side effect (= sqlite3 driver) only
)

const (
	sqliteDriverName = "sqlite3_with_extensions"
	queryGpkgContent = `select
			c.table_name, c.data_type, c.identifier, c.description, c.last_change,
			c.min_x, c.min_y, c.max_x, c.max_y, c.srs_id, gc.column_name, gc.geometry_type_name
		from
			gpkg_contents c join gpkg_geometry_columns gc on c.table_name == gc.table_name
		where
			c.data_type = 'features' and 
			c.min_x is not null`
	bboxSizeBig = 10000
)

// Load sqlite extensions once.
//
// Extensions are expected in /usr/lib. On Linux you can alternatively point LD_LIBRARY_PATH to
// another directory holding the extensions. On Darwin DYLD_LIBRARY_PATH is used for the same purpose.
func init() {
	sql.Register(sqliteDriverName, &sqlite3.SQLiteDriver{Extensions: []string{
		"mod_spatialite",
	}})
}

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

	// TODO validate btree index is present

	return g
}

func (g *GeoPackage) Close() {
	g.backend.close()
}

func (g *GeoPackage) GetFeatures(ctx context.Context, collection string, options datasources.FeatureOptions) (*domain.FeatureCollection, domain.Cursors, error) {
	table, ok := g.featureTableByCollectionID[collection]
	if !ok {
		return nil, domain.Cursors{}, fmt.Errorf("can't query collection '%s' since it doesn't exist in "+
			"geopackage, available in geopackage: %v", collection, util.Keys(g.featureTableByCollectionID))
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	query, queryArgs := g.makeFeaturesQuery(table, options)
	stmt, err := g.backend.getDB().PreparexContext(ctx, query)
	if err != nil {
		return nil, domain.Cursors{}, err
	}
	defer stmt.Close()

	log.Println(query)     // TODO
	log.Println(queryArgs) // TODO

	rows, err := g.backend.getDB().QueryxContext(queryCtx, query, queryArgs...)
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	var nextPrev *domain.PrevNextFID
	result := domain.FeatureCollection{}
	result.Features, nextPrev, err = domain.MapRowsToFeatures(rows, g.fidColumn, table.GeometryColumnName, readGpkgGeometry)
	if err != nil {
		return nil, domain.Cursors{}, err
	}
	if nextPrev == nil {
		return nil, domain.Cursors{}, nil
	}

	result.NumberReturned = len(result.Features)
	return &result, domain.NewCursors(*nextPrev, options.Cursor.FiltersChecksum), nil
}

func (g *GeoPackage) GetFeature(ctx context.Context, collection string, featureID int64) (*domain.Feature, error) {
	table, ok := g.featureTableByCollectionID[collection]
	if !ok {
		return nil, fmt.Errorf("can't query collection '%s' since it doesn't exist in "+
			"geopackage, available in geopackage: %v", collection, util.Keys(g.featureTableByCollectionID))
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

// Build specific features queries based on the given options.
// Make sure to use SQL bind variables! We prefer $N style bind variables for reusability and broader compatibility.
func (g *GeoPackage) makeFeaturesQuery(table *featureTable, opt datasources.FeatureOptions) (string, []any) {
	if opt.Bbox != nil {
		return g.makeBboxQuery(table, opt)
	}
	if opt.Filter != "" {
		return g.makeFiltersQuery(opt)
	}
	return g.makeDefaultQuery(table, opt)
}

func (g *GeoPackage) makeDefaultQuery(table *featureTable, opt datasources.FeatureOptions) (string, []any) {
	defaultQuery := fmt.Sprintf(`
with 
	next as (select * from %[1]s where %[2]s >= $1 order by %[2]s asc limit $2 + 1),
	prev as (select * from %[1]s where %[2]s < $1 order by %[2]s desc limit $2),
	nextprev as (select * from next union all select * from prev),
	nextprevfeat as (select *, lag(%[2]s, $2) over (order by %[2]s) as prevfid, lead(%[2]s, $2) over (order by %[2]s) as nextfid from nextprev)
select * from nextprevfeat where %[2]s >= $1 limit $2
`, table.TableName, g.fidColumn)

	return defaultQuery, []any{opt.Cursor.FID, opt.Limit}
}

func (g *GeoPackage) makeBboxQuery(table *featureTable, opt datasources.FeatureOptions) (string, []any) {
	bboxQuery := fmt.Sprintf(`
with const as (select iif(count(id) < %[3]d, 'small', 'big') as bbox_size
               from (select id from rtree_%[1]s_geom
                     where minx <= $3 and maxx >= $4 and miny <= $5 and maxy >= $6
                     limit %[3]d)),
     next_bbox_rtree as (select f.*
						 from %[1]s f inner join rtree_%[1]s_geom rf on f.%[2]s = rf.id
						 where rf.minx <= $3 and rf.maxx >= $4 and rf.miny <= $5 and rf.maxy >= $6
						   and st_intersects(geomfromtext("polygon (($4 $6,$3 $6,$3 $5,$4 $5,$4 $6))", $7), castautomagic(f.geom)) = 1
						   and f.%[2]s >= $1 
						 order by f.%[2]s asc 
						 limit (select iif(bbox_size == 'small', $2 + 1, 0) from const)),
     next_bbox_btree as (select f.*
                         from %[1]s f indexed by %[1]s_spatial_idx
                         where f.minx <= $3 and f.maxx >= $4 and f.miny <= $5 and f.maxy >= $6
                           and st_intersects(geomfromtext("polygon (($4 $6,$3 $6,$3 $5,$4 $5,$4 $6))", $7), castautomagic(f.geom)) = 1
                           and f.%[2]s >= $1 
					     order by f.%[2]s asc 
                         limit (select iif(bbox_size == 'big', $2 + 1, 0) from const)),
     next as (select * from next_bbox_rtree union all select * from next_bbox_btree),
     prev_bbox_rtree as (select f.*
                         from %[1]s f inner join rtree_%[1]s_geom rf on f.%[2]s = rf.id
                         where rf.minx <= $3 and rf.maxx >= $4 and rf.miny <= $5 and rf.maxy >= $6
                           and st_intersects(geomfromtext("polygon (($4 $6,$3 $6,$3 $5,$4 $5,$4 $6))", $7), castautomagic(f.geom)) = 1
                           and f.%[2]s < $1 
                         order by f.%[2]s desc 
                         limit (select iif(bbox_size == 'small', $2, 0) from const)),
     prev_bbox_btree as (select f.*
                         from %[1]s f indexed by %[1]s_spatial_idx
                         where f.minx <= $3 and f.maxx >= $4 and f.miny <= $5 and f.maxy >= $6
                           and st_intersects(geomfromtext("polygon (($4 $6,$3 $6,$3 $5,$4 $5,$4 $6))", $7), castautomagic(f.geom)) = 1
                           and f.%[2]s < $1 
                         order by f.%[2]s desc 
                         limit (select iif(bbox_size == 'big', $2, 0) from const)),
     prev as (select * from prev_bbox_rtree union all select * from prev_bbox_btree),
     nextprev as (select * from next union all select * from prev),
     nextprevfeat as (select *, lag(%[2]s, $2) over (order by %[2]s) as prevfid, lead(%[2]s, $2) over (order by %[2]s) as nextfid from nextprev)
select * from nextprevfeat where %[2]s >= $1 limit $2
`, table.TableName, g.fidColumn, bboxSizeBig)

	return bboxQuery, []any{opt.Cursor.FID, opt.Limit, opt.Bbox.MaxX(), opt.Bbox.MinX(), opt.Bbox.MaxY(), opt.Bbox.MinY(), opt.BboxCrs}
}

func (g *GeoPackage) makeFiltersQuery(opt datasources.FeatureOptions) (string, []any) {
	// TODO create part3/CQL filter query
	filterQuery := `with <filter query here>`
	return filterQuery, []any{opt.Cursor.FID, opt.Limit, opt.Filter}
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
				} else if hasMatchingDatasourceID(collection, row) {
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

func hasMatchingDatasourceID(collection engine.GeoSpatialCollection, row featureTable) bool {
	return collection.Features != nil && collection.Features.DatasourceID != nil &&
		row.Identifier == *collection.Features.DatasourceID
}

func readGpkgGeometry(rawGeom []byte) (geom.Geometry, error) {
	geometry, err := gpkg.DecodeGeometry(rawGeom)
	if err != nil {
		return nil, err
	}
	return geometry.Geometry, nil
}
