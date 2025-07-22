package geopackage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"sync"

	"github.com/PDOK/gokoala/config"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/common"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/geopackage/encoding"
	d "github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	"github.com/qustavo/sqlhooks/v2"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkt"
)

const (
	sqliteDriverName = "sqlite3_with_extensions"

	sqlNamedParamSymbol = ":"
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
		sql.Register(sqliteDriverName, sqlhooks.Wrap(driver, ds.NewSQLLogFromEnv()))
	})
}

// geoPackageBackend abstraction over different kinds of GeoPackages, e.g. local file or cloud-backed sqlite.
type geoPackageBackend interface {
	getDB() *sqlx.DB
	close()
}

type GeoPackage struct {
	common.DatasourceCommon

	backend           geoPackageBackend
	preparedStmtCache *PreparedStatementCache

	maxBBoxSizeToUseWithRTree int
}

func NewGeoPackage(collections config.GeoSpatialCollections, gpkgConfig config.GeoPackage,
	transformOnTheFly bool, maxDecimals int, forceUTC bool) (*GeoPackage, error) {

	loadDriver()
	if transformOnTheFly {
		return nil, errors.New("on the fly reprojection/transformation is currently not supported for GeoPackages")
	}

	g := &GeoPackage{
		DatasourceCommon: common.DatasourceCommon{
			TransformOnTheFly:        transformOnTheFly,
			MaxDecimals:              maxDecimals,
			ForceUTC:                 forceUTC,
			PropertiesByCollectionID: collections.FeaturePropertiesByID(),
		},
		preparedStmtCache: NewCache(),
	}

	warmUp := false
	switch {
	case gpkgConfig.Local != nil:
		g.backend = newLocalGeoPackage(gpkgConfig.Local)
		g.DatasourceCommon.FidColumn = gpkgConfig.Local.Fid
		g.DatasourceCommon.ExternalFidColumn = gpkgConfig.Local.ExternalFid
		g.DatasourceCommon.QueryTimeout = gpkgConfig.Local.QueryTimeout.Duration
		g.maxBBoxSizeToUseWithRTree = gpkgConfig.Local.MaxBBoxSizeToUseWithRTree
	case gpkgConfig.Cloud != nil:
		g.backend = newCloudBackedGeoPackage(gpkgConfig.Cloud)
		g.DatasourceCommon.FidColumn = gpkgConfig.Cloud.Fid
		g.DatasourceCommon.ExternalFidColumn = gpkgConfig.Cloud.ExternalFid
		g.DatasourceCommon.QueryTimeout = gpkgConfig.Cloud.QueryTimeout.Duration
		g.maxBBoxSizeToUseWithRTree = gpkgConfig.Cloud.MaxBBoxSizeToUseWithRTree
		warmUp = gpkgConfig.Cloud.Cache.WarmUp
	default:
		return nil, errors.New("unknown GeoPackage config encountered")
	}

	g.FeatureTableByCollectionID, g.PropertyFiltersByCollectionID = readMetadata(
		g.backend.getDB(), collections, g.FidColumn, g.ExternalFidColumn)

	if err := assertIndexesExist(collections, g.FeatureTableByCollectionID, g.backend.getDB(), g.FidColumn); err != nil {
		return nil, err
	}
	if warmUp {
		// perform warmup async since it can take a long time
		go func() {
			if err := warmUpFeatureTables(collections, g.FeatureTableByCollectionID, g.backend.getDB()); err != nil {
				log.Fatal(err)
			}
		}()
	}
	return g, nil
}

func (g *GeoPackage) Close() {
	g.preparedStmtCache.Close()
	g.backend.close()
}

func (g *GeoPackage) GetFeatureIDs(ctx context.Context, collection string, criteria ds.FeaturesCriteria) ([]int64, d.Cursors, error) {
	table, err := g.GetFeatureTable(collection)
	if err != nil {
		return nil, d.Cursors{}, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.QueryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	propConfig := g.PropertiesByCollectionID[collection]
	stmt, query, queryArgs, err := g.makeFeaturesQuery(queryCtx, propConfig, table, true, -1, criteria) //nolint:sqlclosecheck // prepared statement is cached, will be closed when evicted from cache
	if err != nil {
		return nil, d.Cursors{}, fmt.Errorf("failed to create query '%s' error: %w", query, err)
	}

	rows, err := stmt.QueryxContext(queryCtx, queryArgs)
	if err != nil {
		return nil, d.Cursors{}, fmt.Errorf("failed to execute query '%s' error: %w", query, err)
	}
	defer rows.Close()

	featureIDs, prevNext, err := common.MapRowsToFeatureIDs(queryCtx, FromSqlxRows(rows))
	if err != nil {
		return nil, d.Cursors{}, err
	}
	if prevNext == nil {
		return nil, d.Cursors{}, nil
	}
	return featureIDs, d.NewCursors(*prevNext, criteria.Cursor.FiltersChecksum), queryCtx.Err()
}

func (g *GeoPackage) GetFeaturesByID(ctx context.Context, collection string, featureIDs []int64,
	axisOrder d.AxisOrder, profile d.Profile) (*d.FeatureCollection, error) {

	table, err := g.GetFeatureTable(collection)
	if err != nil {
		return nil, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.QueryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	propConfig := g.PropertiesByCollectionID[collection]
	selectClause := g.SelectColumns(table, axisOrder, selectGpkgGeometry, propConfig, false)
	fids := map[string]any{"fids": featureIDs}

	query, queryArgs, err := sqlx.Named(fmt.Sprintf("select %s from %s where %s in (:fids)",
		selectClause, table.TableName, g.FidColumn), fids)
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

	fc := d.FeatureCollection{}
	fc.Features, _, err = common.MapRowsToFeatures(queryCtx, FromSqlxRows(rows),
		g.FidColumn, g.ExternalFidColumn, table.GeometryColumnName,
		propConfig, table.Schema, mapGpkgGeometry, profile.MapRelationUsingProfile,
		common.FormatOpts{MaxDecimals: g.MaxDecimals, ForceUTC: g.ForceUTC})
	if err != nil {
		return nil, err
	}
	fc.NumberReturned = len(fc.Features)
	return &fc, queryCtx.Err()
}

func (g *GeoPackage) GetFeatures(ctx context.Context, collection string, criteria ds.FeaturesCriteria,
	axisOrder d.AxisOrder, profile d.Profile) (*d.FeatureCollection, d.Cursors, error) {

	table, err := g.GetFeatureTable(collection)
	if err != nil {
		return nil, d.Cursors{}, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.QueryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	propConfig := g.PropertiesByCollectionID[collection]
	stmt, query, queryArgs, err := g.makeFeaturesQuery(queryCtx, propConfig, table, false, axisOrder, criteria) //nolint:sqlclosecheck // prepared statement is cached, will be closed when evicted from cache
	if err != nil {
		return nil, d.Cursors{}, fmt.Errorf("failed to create query '%s' error: %w", query, err)
	}

	rows, err := stmt.QueryxContext(queryCtx, queryArgs)
	if err != nil {
		return nil, d.Cursors{}, fmt.Errorf("failed to execute query '%s' error: %w", query, err)
	}
	defer rows.Close()

	var prevNext *d.PrevNextFID
	fc := d.FeatureCollection{}
	fc.Features, prevNext, err = common.MapRowsToFeatures(queryCtx, FromSqlxRows(rows),
		g.FidColumn, g.ExternalFidColumn, table.GeometryColumnName,
		propConfig, table.Schema, mapGpkgGeometry, profile.MapRelationUsingProfile,
		common.FormatOpts{MaxDecimals: g.MaxDecimals, ForceUTC: g.ForceUTC})
	if err != nil {
		return nil, d.Cursors{}, err
	}
	if prevNext == nil {
		return nil, d.Cursors{}, nil
	}
	fc.NumberReturned = len(fc.Features)
	return &fc, d.NewCursors(*prevNext, criteria.Cursor.FiltersChecksum), queryCtx.Err()
}

func (g *GeoPackage) GetFeature(ctx context.Context, collection string, featureID any,
	_ d.SRID, axisOrder d.AxisOrder, profile d.Profile) (*d.Feature, error) {

	table, err := g.GetFeatureTable(collection)
	if err != nil {
		return nil, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.QueryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	var fidColumn string
	switch featureID.(type) {
	case int64:
		if g.ExternalFidColumn != "" {
			// Features should be retrieved by UUID
			log.Println("feature requested by int while external fid column is defined")
			return nil, nil
		}
		fidColumn = g.FidColumn
	case uuid.UUID:
		if g.ExternalFidColumn == "" {
			// Features should be retrieved by int64
			log.Println("feature requested by UUID while external fid column is not defined")
			return nil, nil
		}
		fidColumn = g.ExternalFidColumn
	}

	propConfig := g.PropertiesByCollectionID[collection]
	selectClause := g.SelectColumns(table, axisOrder, selectGpkgGeometry, propConfig, false)

	query := fmt.Sprintf(`select %s from "%s" where "%s" = :fid limit 1`, selectClause, table.TableName, fidColumn)
	rows, err := g.backend.getDB().NamedQueryContext(queryCtx, query, map[string]any{"fid": featureID})
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	features, _, err := common.MapRowsToFeatures(queryCtx, FromSqlxRows(rows),
		g.FidColumn, g.ExternalFidColumn, table.GeometryColumnName,
		propConfig, table.Schema, mapGpkgGeometry, profile.MapRelationUsingProfile,
		common.FormatOpts{MaxDecimals: g.MaxDecimals, ForceUTC: g.ForceUTC})
	if err != nil {
		return nil, err
	}
	if len(features) != 1 {
		return nil, nil
	}
	return features[0], queryCtx.Err()
}

// Build specific features queries based on the given options.
// Make sure to use SQL bind variables and return named params: https://jmoiron.github.io/sqlx/#namedParams
func (g *GeoPackage) makeFeaturesQuery(ctx context.Context, propConfig *config.FeatureProperties, table *common.FeatureTable,
	onlyFIDs bool, axisOrder d.AxisOrder, criteria ds.FeaturesCriteria) (stmt *sqlx.NamedStmt, query string, queryArgs map[string]any, err error) {

	var selectClause string
	if onlyFIDs {
		selectClause = common.ColumnsToSQL([]string{g.FidColumn, d.PrevFid, d.NextFid})
	} else {
		selectClause = g.SelectColumns(table, axisOrder, selectGpkgGeometry, propConfig, true)
	}

	// make query
	if criteria.Bbox != nil {
		query, queryArgs, err = g.makeBboxQuery(table, selectClause, criteria)
		if err != nil {
			return
		}
	} else {
		query, queryArgs = g.makeDefaultQuery(table, selectClause, criteria)
	}
	// lookup prepared statement for given query, or create new one
	stmt, err = g.preparedStmtCache.Lookup(ctx, g.backend.getDB(), query)
	return
}

func (g *GeoPackage) makeDefaultQuery(table *common.FeatureTable, selectClause string, criteria ds.FeaturesCriteria) (string, map[string]any) {
	pfClause, pfNamedParams := common.PropertyFiltersToSQL(criteria.PropertyFilters, sqlNamedParamSymbol)
	temporalClause, temporalNamedParams := common.TemporalCriteriaToSQL(criteria.TemporalCriteria, sqlNamedParamSymbol)

	defaultQuery := fmt.Sprintf(`
with
    next as (select * from "%[1]s" where "%[2]s" >= :fid %[3]s %[4]s order by %[2]s asc limit :limit + 1),
    prev as (select * from "%[1]s" where "%[2]s" < :fid %[3]s %[4]s order by %[2]s desc limit :limit),
    nextprev as (select * from next union all select * from prev),
    nextprevfeat as (select *, lag("%[2]s", :limit) over (order by %[2]s) as %[6]s, lead("%[2]s", :limit) over (order by "%[2]s") as %[7]s from nextprev)
select %[5]s from nextprevfeat where "%[2]s" >= :fid %[3]s %[4]s limit :limit
`, table.TableName, g.FidColumn, temporalClause, pfClause, selectClause, d.PrevFid, d.NextFid) // don't add user input here, use named params for user input!

	namedParams := map[string]any{
		"fid":   criteria.Cursor.FID,
		"limit": criteria.Limit,
	}
	maps.Copy(namedParams, pfNamedParams)
	maps.Copy(namedParams, temporalNamedParams)
	return defaultQuery, namedParams
}

func (g *GeoPackage) makeBboxQuery(table *common.FeatureTable, selectClause string, criteria ds.FeaturesCriteria) (string, map[string]any, error) {
	btreeIndexHint := fmt.Sprintf("indexed by \"%s_spatial_idx\"", table.TableName)

	pfClause, pfNamedParams := common.PropertyFiltersToSQL(criteria.PropertyFilters, sqlNamedParamSymbol)
	if pfClause != "" {
		// don't force btree index when using property filter, let SQLite decide
		// whether to use the BTree index or the property filter index
		btreeIndexHint = ""
	}
	temporalClause, temporalNamedParams := common.TemporalCriteriaToSQL(criteria.TemporalCriteria, sqlNamedParamSymbol)

	bboxQuery := fmt.Sprintf(`
with
     given_bbox as (select geomfromtext(:bboxWkt, :bboxSrid)),
     bbox_size as (select iif(count(id) < %[3]d, 'small', 'big') as bbox_size
                     from (select id from rtree_%[1]s_%[4]s
                           where minx <= :maxx and maxx >= :minx and miny <= :maxy and maxy >= :miny
                           limit %[3]d)),
     next_bbox_rtree as (select f.*
                         from "%[1]s" f inner join rtree_%[1]s_%[4]s rf on f."%[2]s" = rf.id
                         where rf.minx <= :maxx and rf.maxx >= :minx and rf.miny <= :maxy and rf.maxy >= :miny
                           and st_intersects((select * from given_bbox), castautomagic(f.%[4]s)) = 1
                           and f."%[2]s" >= :fid %[6]s %[7]s
                         order by f."%[2]s" asc
                         limit (select iif(bbox_size == 'small', :limit + 1, 0) from bbox_size)),
     next_bbox_btree as (select f.*
                         from "%[1]s" f %[8]s
                         where f.minx <= :maxx and f.maxx >= :minx and f.miny <= :maxy and f.maxy >= :miny
                           and st_intersects((select * from given_bbox), castautomagic(f.%[4]s)) = 1
                           and f."%[2]s" >= :fid %[6]s %[7]s
                         order by f."%[2]s" asc
                         limit (select iif(bbox_size == 'big', :limit + 1, 0) from bbox_size)),
     next as (select * from next_bbox_rtree union all select * from next_bbox_btree),
     prev_bbox_rtree as (select f.*
                         from "%[1]s" f inner join rtree_%[1]s_%[4]s rf on f."%[2]s" = rf.id
                         where rf.minx <= :maxx and rf.maxx >= :minx and rf.miny <= :maxy and rf.maxy >= :miny
                           and st_intersects((select * from given_bbox), castautomagic(f.%[4]s)) = 1
                           and f."%[2]s" < :fid %[6]s %[7]s
                         order by f."%[2]s" desc
                         limit (select iif(bbox_size == 'small', :limit, 0) from bbox_size)),
     prev_bbox_btree as (select f.*
                         from "%[1]s" f %[8]s
                         where f.minx <= :maxx and f.maxx >= :minx and f.miny <= :maxy and f.maxy >= :miny
                           and st_intersects((select * from given_bbox), castautomagic(f.%[4]s)) = 1
                           and f."%[2]s" < :fid %[6]s %[7]s
                         order by f."%[2]s" desc
                         limit (select iif(bbox_size == 'big', :limit, 0) from bbox_size)),
     prev as (select * from prev_bbox_rtree union all select * from prev_bbox_btree),
     nextprev as (select * from next union all select * from prev),
     nextprevfeat as (select *, lag("%[2]s", :limit) over (order by "%[2]s") as %[9]s, lead("%[2]s", :limit) over (order by "%[2]s") as %[10]s from nextprev)
select %[5]s from nextprevfeat where "%[2]s" >= :fid %[6]s %[7]s limit :limit
`, table.TableName, g.FidColumn, g.maxBBoxSizeToUseWithRTree, table.GeometryColumnName,
		selectClause, temporalClause, pfClause, btreeIndexHint, d.PrevFid, d.NextFid) // don't add user input here, use named params for user input!

	bboxAsWKT, err := wkt.Marshal(criteria.Bbox.Polygon())
	if err != nil {
		return "", nil, err
	}
	namedParams := map[string]any{
		"fid":       criteria.Cursor.FID,
		"limit":     criteria.Limit,
		"bboxWkt":   bboxAsWKT,
		d.MaxxField: criteria.Bbox.Max(0),
		d.MinxField: criteria.Bbox.Min(0),
		d.MaxyField: criteria.Bbox.Max(1),
		d.MinyField: criteria.Bbox.Min(1),
		"bboxSrid":  criteria.InputSRID}
	maps.Copy(namedParams, pfNamedParams)
	maps.Copy(namedParams, temporalNamedParams)
	return bboxQuery, namedParams, nil
}

// mapGpkgGeometry GeoPackage specific way to read geometries into a geom.T
func mapGpkgGeometry(columnValue any) (geom.T, error) {
	rawGeom, ok := columnValue.([]byte)
	if !ok {
		return nil, errors.New("failed to cast GeoPackage geom to bytes")
	}
	geomWithMetadata, err := encoding.DecodeGeometry(rawGeom)
	if err != nil {
		return nil, err
	}
	if geomWithMetadata == nil || geomWithMetadata.Geometry.Empty() {
		return nil, nil
	}
	return geomWithMetadata.Geometry, nil
}

// selectGpkgGeometry GeoPackage specific way to select geometry
// and take d.AxisOrder into account.
func selectGpkgGeometry(axisOrder d.AxisOrder, table *common.FeatureTable) string {
	if axisOrder == d.AxisOrderYX {
		// Add the geometry column. GeoPackage geometries are stored in WKB format and WKB is always XY.
		// So swap coordinates when needed. This requires casting to a SpatiaLite geometry first, executing
		// the swap and then casting back to a GeoPackage geometry.
		return fmt.Sprintf(", asgpb(swapcoords(castautomagic(\"%[1]s\"))) as \"%[1]s\"", table.GeometryColumnName)
	}
	return fmt.Sprintf(", \"%s\"", table.GeometryColumnName)
}
