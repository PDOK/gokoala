package geopackage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"time"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/engine/util"
	"github.com/PDOK/gokoala/ogc/features/datasources"
	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/gpkg"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	"github.com/qustavo/sqlhooks/v2"

	_ "github.com/mattn/go-sqlite3" // import for side effect (= sqlite3 driver) only
)

const (
	sqliteDriverName = "sqlite3_with_extensions"
	bboxSizeBig      = 10000
)

// Load sqlite extensions once.
//
// Extensions are by default expected in /usr/lib. For spatialite you can
// alternatively/optionally set SPATIALITE_LIBRARY_PATH.
func init() {
	spatialite := path.Join(os.Getenv("SPATIALITE_LIBRARY_PATH"), "mod_spatialite")
	driver := &sqlite3.SQLiteDriver{Extensions: []string{spatialite}}
	sql.Register(sqliteDriverName, sqlhooks.Wrap(driver, &datasources.SQLLog{}))
}

type geoPackageBackend interface {
	getDB() *sqlx.DB
	close()
}

type featureTable struct {
	TableName          string    `db:"table_name"`
	DataType           string    `db:"data_type"` // always 'features'
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

	ColumnsWithDateType map[string]string
}

func (ft featureTable) ColumnsWithDataType() map[string]string {
	return ft.ColumnsWithDateType
}

type GeoPackage struct {
	backend           geoPackageBackend
	preparedStmtCache *PreparedStatementCache

	fidColumn                  string
	featureTableByCollectionID map[string]*featureTable
	queryTimeout               time.Duration
}

func NewGeoPackage(collections engine.GeoSpatialCollections, gpkgConfig engine.GeoPackage) *GeoPackage {
	g := &GeoPackage{}
	g.preparedStmtCache = NewCache()

	switch {
	case gpkgConfig.Local != nil:
		g.backend = newLocalGeoPackage(gpkgConfig.Local)
		g.fidColumn = gpkgConfig.Local.Fid
		g.queryTimeout = gpkgConfig.Local.QueryTimeout
	case gpkgConfig.Cloud != nil:
		g.backend = newCloudBackedGeoPackage(gpkgConfig.Cloud)
		g.fidColumn = gpkgConfig.Cloud.Fid
		g.queryTimeout = gpkgConfig.Cloud.QueryTimeout
	default:
		log.Fatal("unknown GeoPackage config encountered")
	}

	metadata, err := readDriverMetadata(g.backend.getDB())
	if err != nil {
		log.Fatalf("failed to connect with GeoPackage: %v", err)
	}
	log.Println(metadata)

	g.featureTableByCollectionID, err = readGpkgContents(collections, g.backend.getDB())
	if err != nil {
		log.Fatal(err)
	}

	if err = assertIndexesExist(collections, g.featureTableByCollectionID, g.backend.getDB(), g.fidColumn); err != nil {
		log.Fatal(err)
	}

	return g
}

func (g *GeoPackage) Close() {
	g.preparedStmtCache.Close()
	g.backend.close()
}

func (g *GeoPackage) GetFeatureIDs(ctx context.Context, collection string, criteria datasources.FeaturesCriteria) ([]int64, domain.Cursors, error) {
	table, err := g.getFeatureTable(collection)
	if err != nil {
		return nil, domain.Cursors{}, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	stmt, query, queryArgs, err := g.makeFeaturesQuery(queryCtx, table, true, criteria) //nolint:sqlclosecheck // prepared statement is cached, will be closed when evicted from cache
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("failed to create query '%s' error: %w", query, err)
	}

	rows, err := stmt.QueryxContext(queryCtx, queryArgs)
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("failed to execute query '%s' error: %w", query, err)
	}
	defer rows.Close()

	featureIDs, prevNext, err := domain.MapRowsToFeatureIDs(rows)
	if err != nil {
		return nil, domain.Cursors{}, err
	}
	if prevNext == nil {
		return nil, domain.Cursors{}, nil
	}
	return featureIDs, domain.NewCursors(*prevNext, criteria.Cursor.FiltersChecksum), nil
}

func (g *GeoPackage) GetFeaturesByID(ctx context.Context, collection string, featureIDs []int64) (*domain.FeatureCollection, error) {
	table, err := g.getFeatureTable(collection)
	if err != nil {
		return nil, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	fids := map[string]any{"fids": featureIDs}
	query, queryArgs, err := sqlx.Named(fmt.Sprintf("select * from %s where %s in (:fids)", table.TableName, g.fidColumn), fids)
	if err != nil {
		return nil, fmt.Errorf("failed to make features query, error: %w", err)
	}
	query, queryArgs, err = sqlx.In(query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to make IN-clause, error: %w", err)
	}

	rows, err := g.backend.getDB().QueryxContext(queryCtx, g.backend.getDB().Rebind(query), queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query '%s' error: %w", query, err)
	}
	defer rows.Close()

	fc := domain.FeatureCollection{}
	fc.Features, _, err = domain.MapRowsToFeatures(rows, g.fidColumn, table.GeometryColumnName, readGpkgGeometry)
	if err != nil {
		return nil, err
	}
	fc.NumberReturned = len(fc.Features)
	return &fc, nil
}

func (g *GeoPackage) GetFeatures(ctx context.Context, collection string, criteria datasources.FeaturesCriteria) (*domain.FeatureCollection, domain.Cursors, error) {
	table, err := g.getFeatureTable(collection)
	if err != nil {
		return nil, domain.Cursors{}, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	stmt, query, queryArgs, err := g.makeFeaturesQuery(queryCtx, table, false, criteria) //nolint:sqlclosecheck // prepared statement is cached, will be closed when evicted from cache
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("failed to create query '%s' error: %w", query, err)
	}

	rows, err := stmt.QueryxContext(queryCtx, queryArgs)
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("failed to execute query '%s' error: %w", query, err)
	}
	defer rows.Close()

	var prevNext *domain.PrevNextFID
	fc := domain.FeatureCollection{}
	fc.Features, prevNext, err = domain.MapRowsToFeatures(rows, g.fidColumn, table.GeometryColumnName, readGpkgGeometry)
	if err != nil {
		return nil, domain.Cursors{}, err
	}
	if prevNext == nil {
		return nil, domain.Cursors{}, nil
	}
	fc.NumberReturned = len(fc.Features)
	return &fc, domain.NewCursors(*prevNext, criteria.Cursor.FiltersChecksum), nil
}

func (g *GeoPackage) GetFeature(ctx context.Context, collection string, featureID int64) (*domain.Feature, error) {
	table, err := g.getFeatureTable(collection)
	if err != nil {
		return nil, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	query := fmt.Sprintf("select * from %s f where f.%s = :fid limit 1", table.TableName, g.fidColumn)
	rows, err := g.backend.getDB().NamedQueryContext(queryCtx, query, map[string]any{"fid": featureID})
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	features, _, err := domain.MapRowsToFeatures(rows, g.fidColumn, table.GeometryColumnName, readGpkgGeometry)
	if err != nil {
		return nil, err
	}
	if len(features) != 1 {
		return nil, nil
	}
	return features[0], nil
}

func (g *GeoPackage) GetFeatureTableMetadata(collection string) (datasources.FeatureTableMetadata, error) {
	val, ok := g.featureTableByCollectionID[collection]
	if !ok {
		return nil, fmt.Errorf("no metadata for %s", collection)
	}
	return val, nil
}

// Build specific features queries based on the given options.
// Make sure to use SQL bind variables and return named params: https://jmoiron.github.io/sqlx/#namedParams
func (g *GeoPackage) makeFeaturesQuery(ctx context.Context, table *featureTable, onlyFIDs bool,
	criteria datasources.FeaturesCriteria) (stmt *sqlx.NamedStmt, query string, queryArgs map[string]any, err error) {

	// make query
	if criteria.Bbox != nil {
		query, queryArgs, err = g.makeBboxQuery(table, onlyFIDs, criteria)
		if err != nil {
			return
		}
	} else {
		query, queryArgs = g.makeDefaultQuery(table, criteria)
	}
	// lookup prepared statement for given query, or create new one
	stmt, err = g.preparedStmtCache.Lookup(ctx, g.backend.getDB(), query)
	return
}

func (g *GeoPackage) makeDefaultQuery(table *featureTable, criteria datasources.FeaturesCriteria) (string, map[string]any) {
	pfClause, pfNamedParams := propertyFiltersToSQL(criteria.PropertyFilters)

	defaultQuery := fmt.Sprintf(`
with 
    next as (select * from %[1]s where %[2]s >= :fid %[3]s order by %[2]s asc limit :limit + 1),
    prev as (select * from %[1]s where %[2]s < :fid %[3]s order by %[2]s desc limit :limit),
    nextprev as (select * from next union all select * from prev),
    nextprevfeat as (select *, lag(%[2]s, :limit) over (order by %[2]s) as prevfid, lead(%[2]s, :limit) over (order by %[2]s) as nextfid from nextprev)
select * from nextprevfeat where %[2]s >= :fid %[3]s limit :limit
`, table.TableName, g.fidColumn, pfClause) // don't add user input here, use named params for user input!

	namedParams := map[string]any{
		"fid":   criteria.Cursor.FID,
		"limit": criteria.Limit,
	}
	maps.Copy(namedParams, pfNamedParams)
	return defaultQuery, namedParams
}

func (g *GeoPackage) makeBboxQuery(table *featureTable, onlyFIDs bool, criteria datasources.FeaturesCriteria) (string, map[string]any, error) {
	selectClause := "*"
	if onlyFIDs {
		selectClause = fmt.Sprintf("%s, prevfid, nextfid", g.fidColumn)
	}

	pfClause, pfNamedParams := propertyFiltersToSQL(criteria.PropertyFilters)

	bboxQuery := fmt.Sprintf(`
with 
     given_bbox as (select geomfromtext(:bboxWkt, :bboxSrid)),
     bbox_size as (select iif(count(id) < %[3]d, 'small', 'big') as bbox_size
                     from (select id from rtree_%[1]s_%[4]s
                           where minx <= :maxx and maxx >= :minx and miny <= :maxy and maxy >= :miny
                           limit %[3]d)),
     next_bbox_rtree as (select f.*
                         from %[1]s f inner join rtree_%[1]s_%[4]s rf on f.%[2]s = rf.id
                         where rf.minx <= :maxx and rf.maxx >= :minx and rf.miny <= :maxy and rf.maxy >= :miny
                           and st_intersects((select * from given_bbox), castautomagic(f.%[4]s)) = 1
                           and f.%[2]s >= :fid %[6]s
                         order by f.%[2]s asc 
                         limit (select iif(bbox_size == 'small', :limit + 1, 0) from bbox_size)),
     next_bbox_btree as (select f.*
                         from %[1]s f
                         where f.minx <= :maxx and f.maxx >= :minx and f.miny <= :maxy and f.maxy >= :miny
                           and st_intersects((select * from given_bbox), castautomagic(f.%[4]s)) = 1
                           and f.%[2]s >= :fid %[6]s
                         order by f.%[2]s asc 
                         limit (select iif(bbox_size == 'big', :limit + 1, 0) from bbox_size)),
     next as (select * from next_bbox_rtree union all select * from next_bbox_btree),
     prev_bbox_rtree as (select f.*
                         from %[1]s f inner join rtree_%[1]s_%[4]s rf on f.%[2]s = rf.id
                         where rf.minx <= :maxx and rf.maxx >= :minx and rf.miny <= :maxy and rf.maxy >= :miny
                           and st_intersects((select * from given_bbox), castautomagic(f.%[4]s)) = 1
                           and f.%[2]s < :fid %[6]s
                         order by f.%[2]s desc 
                         limit (select iif(bbox_size == 'small', :limit, 0) from bbox_size)),
     prev_bbox_btree as (select f.*
                         from %[1]s f
                         where f.minx <= :maxx and f.maxx >= :minx and f.miny <= :maxy and f.maxy >= :miny
                           and st_intersects((select * from given_bbox), castautomagic(f.%[4]s)) = 1
                           and f.%[2]s < :fid %[6]s
                         order by f.%[2]s desc 
                         limit (select iif(bbox_size == 'big', :limit, 0) from bbox_size)),
     prev as (select * from prev_bbox_rtree union all select * from prev_bbox_btree),
     nextprev as (select * from next union all select * from prev),
     nextprevfeat as (select *, lag(%[2]s, :limit) over (order by %[2]s) as prevfid, lead(%[2]s, :limit) over (order by %[2]s) as nextfid from nextprev)
select %[5]s from nextprevfeat where %[2]s >= :fid %[6]s limit :limit
`, table.TableName, g.fidColumn, bboxSizeBig, table.GeometryColumnName, selectClause, pfClause) // don't add user input here, use named params for user input!

	bboxAsWKT, err := wkt.EncodeString(criteria.Bbox)
	if err != nil {
		return "", nil, err
	}
	namedParams := map[string]any{
		"fid":      criteria.Cursor.FID,
		"limit":    criteria.Limit,
		"bboxWkt":  bboxAsWKT,
		"maxx":     criteria.Bbox.MaxX(),
		"minx":     criteria.Bbox.MinX(),
		"maxy":     criteria.Bbox.MaxY(),
		"miny":     criteria.Bbox.MinY(),
		"bboxSrid": criteria.InputSRID}
	maps.Copy(namedParams, pfNamedParams)
	return bboxQuery, namedParams, nil
}

func (g *GeoPackage) getFeatureTable(collection string) (*featureTable, error) {
	table, ok := g.featureTableByCollectionID[collection]
	if !ok {
		return nil, fmt.Errorf("can't query collection '%s' since it doesn't exist in "+
			"geopackage, available in geopackage: %v", collection, util.Keys(g.featureTableByCollectionID))
	}
	return table, nil
}

func readGpkgGeometry(rawGeom []byte) (geom.Geometry, error) {
	geometry, err := gpkg.DecodeGeometry(rawGeom)
	if err != nil {
		return nil, err
	}
	return geometry.Geometry, nil
}

func propertyFiltersToSQL(pf map[string]string) (sql string, namedParams map[string]any) {
	namedParams = make(map[string]any)
	if len(pf) > 0 {
		position := 0
		for k, v := range pf {
			position++
			namedParam := fmt.Sprintf("pf%d", position)
			// column name in double quotes in case it is a reserved keyword
			// also: we don't currently support LIKE since wildcard searches don't use the index
			sql += fmt.Sprintf(" and \"%s\" = :%s", k, namedParam)
			namedParams[namedParam] = v
		}
	}
	return sql, namedParams
}
