package geopackage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/geopackage/encoding"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkt"
	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	"github.com/qustavo/sqlhooks/v2"
)

const (
	sqliteDriverName = "sqlite3_with_extensions"
	selectAll        = "*"
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
		sql.Register(sqliteDriverName, sqlhooks.Wrap(driver, datasources.NewSQLLogFromEnv()))
	})
}

// geoPackageBackend abstraction over different kinds of GeoPackages, e.g. local file or cloud-backed sqlite.
type geoPackageBackend interface {
	getDB() *sqlx.DB
	close()
}

type GeoPackage struct {
	backend           geoPackageBackend
	preparedStmtCache *PreparedStatementCache

	fidColumn                     string
	externalFidColumn             string
	featureTableByCollectionID    map[string]*featureTable
	propertyFiltersByCollectionID map[string]datasources.PropertyFiltersWithAllowedValues
	propertiesByCollectionID      map[string]*config.FeatureProperties
	queryTimeout                  time.Duration
	maxBBoxSizeToUseWithRTree     int
}

func NewGeoPackage(collections config.GeoSpatialCollections, gpkgConfig config.GeoPackage) *GeoPackage {
	loadDriver()

	g := &GeoPackage{}
	g.preparedStmtCache = NewCache()
	g.propertiesByCollectionID = cacheFeatureProperties(collections)
	warmUp := false

	switch {
	case gpkgConfig.Local != nil:
		g.backend = newLocalGeoPackage(gpkgConfig.Local)
		g.fidColumn = gpkgConfig.Local.Fid
		g.externalFidColumn = gpkgConfig.Local.ExternalFid
		g.queryTimeout = gpkgConfig.Local.QueryTimeout.Duration
		g.maxBBoxSizeToUseWithRTree = gpkgConfig.Local.MaxBBoxSizeToUseWithRTree
	case gpkgConfig.Cloud != nil:
		g.backend = newCloudBackedGeoPackage(gpkgConfig.Cloud)
		g.fidColumn = gpkgConfig.Cloud.Fid
		g.externalFidColumn = gpkgConfig.Cloud.ExternalFid
		g.queryTimeout = gpkgConfig.Cloud.QueryTimeout.Duration
		g.maxBBoxSizeToUseWithRTree = gpkgConfig.Cloud.MaxBBoxSizeToUseWithRTree
		warmUp = gpkgConfig.Cloud.Cache.WarmUp
	default:
		log.Fatal("unknown GeoPackage config encountered")
	}

	g.featureTableByCollectionID, g.propertyFiltersByCollectionID = readMetadata(
		g.backend.getDB(), collections, g.fidColumn, g.externalFidColumn)

	if err := assertIndexesExist(collections, g.featureTableByCollectionID, g.backend.getDB(), g.fidColumn); err != nil {
		log.Fatal(err)
	}
	if warmUp {
		// perform warmup async since it can take a long time
		go func() {
			if err := warmUpFeatureTables(collections, g.featureTableByCollectionID, g.backend.getDB()); err != nil {
				log.Fatal(err)
			}
		}()
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

	propConfig := g.propertiesByCollectionID[collection]
	stmt, query, queryArgs, err := g.makeFeaturesQuery(queryCtx, propConfig, table, true, -1, criteria) //nolint:sqlclosecheck // prepared statement is cached, will be closed when evicted from cache
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("failed to create query '%s' error: %w", query, err)
	}

	rows, err := stmt.QueryxContext(queryCtx, queryArgs)
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("failed to execute query '%s' error: %w", query, err)
	}
	defer rows.Close()

	featureIDs, prevNext, err := domain.MapRowsToFeatureIDs(queryCtx, rows)
	if err != nil {
		return nil, domain.Cursors{}, err
	}
	if prevNext == nil {
		return nil, domain.Cursors{}, nil
	}
	return featureIDs, domain.NewCursors(*prevNext, criteria.Cursor.FiltersChecksum), queryCtx.Err()
}

func (g *GeoPackage) GetFeaturesByID(ctx context.Context, collection string, featureIDs []int64,
	axisOrder domain.AxisOrder, profile domain.Profile) (*domain.FeatureCollection, error) {

	table, err := g.getFeatureTable(collection)
	if err != nil {
		return nil, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	propConfig := g.propertiesByCollectionID[collection]
	selectClause := g.selectColumns(table, axisOrder, propConfig, false)
	fids := map[string]any{"fids": featureIDs}

	query, queryArgs, err := sqlx.Named(fmt.Sprintf("select %s from %s where %s in (:fids)", selectClause, table.TableName, g.fidColumn), fids)
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
	fc.Features, _, err = domain.MapRowsToFeatures(queryCtx, rows, g.fidColumn, g.externalFidColumn, table.GeometryColumnName,
		propConfig, table.Schema, mapGpkgGeometry, profile.MapRelationUsingProfile)
	if err != nil {
		return nil, err
	}
	fc.NumberReturned = len(fc.Features)
	return &fc, queryCtx.Err()
}

func (g *GeoPackage) GetFeatures(ctx context.Context, collection string, criteria datasources.FeaturesCriteria,
	axisOrder domain.AxisOrder, profile domain.Profile) (*domain.FeatureCollection, domain.Cursors, error) {

	table, err := g.getFeatureTable(collection)
	if err != nil {
		return nil, domain.Cursors{}, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	propConfig := g.propertiesByCollectionID[collection]
	stmt, query, queryArgs, err := g.makeFeaturesQuery(queryCtx, propConfig, table, false, axisOrder, criteria) //nolint:sqlclosecheck // prepared statement is cached, will be closed when evicted from cache
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
	fc.Features, prevNext, err = domain.MapRowsToFeatures(queryCtx, rows, g.fidColumn, g.externalFidColumn, table.GeometryColumnName,
		propConfig, table.Schema, mapGpkgGeometry, profile.MapRelationUsingProfile)
	if err != nil {
		return nil, domain.Cursors{}, err
	}
	if prevNext == nil {
		return nil, domain.Cursors{}, nil
	}
	fc.NumberReturned = len(fc.Features)
	return &fc, domain.NewCursors(*prevNext, criteria.Cursor.FiltersChecksum), queryCtx.Err()
}

func (g *GeoPackage) GetFeature(ctx context.Context, collection string, featureID any,
	axisOrder domain.AxisOrder, profile domain.Profile) (*domain.Feature, error) {

	table, err := g.getFeatureTable(collection)
	if err != nil {
		return nil, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	var fidColumn string
	switch featureID.(type) {
	case int64:
		if g.externalFidColumn != "" {
			// Features should be retrieved by UUID
			log.Println("feature requested by int while external fid column is defined")
			return nil, nil
		}
		fidColumn = g.fidColumn
	case uuid.UUID:
		if g.externalFidColumn == "" {
			// Features should be retrieved by int64
			log.Println("feature requested by UUID while external fid column is not defined")
			return nil, nil
		}
		fidColumn = g.externalFidColumn
	}

	propConfig := g.propertiesByCollectionID[collection]
	selectClause := g.selectColumns(table, axisOrder, propConfig, false)

	query := fmt.Sprintf("select %s from %s f where f.%s = :fid limit 1", selectClause, table.TableName, fidColumn)
	rows, err := g.backend.getDB().NamedQueryContext(queryCtx, query, map[string]any{"fid": featureID})
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	features, _, err := domain.MapRowsToFeatures(queryCtx, rows, g.fidColumn, g.externalFidColumn, table.GeometryColumnName,
		propConfig, table.Schema, mapGpkgGeometry, profile.MapRelationUsingProfile)
	if err != nil {
		return nil, err
	}
	if len(features) != 1 {
		return nil, nil
	}
	return features[0], queryCtx.Err()
}

func (g *GeoPackage) GetSchema(collection string) (*domain.Schema, error) {
	table, err := g.getFeatureTable(collection)
	if err != nil {
		return nil, err
	}
	return table.Schema, nil
}

func (g *GeoPackage) GetPropertyFiltersWithAllowedValues(collection string) datasources.PropertyFiltersWithAllowedValues {
	return g.propertyFiltersByCollectionID[collection]
}

// Build specific features queries based on the given options.
// Make sure to use SQL bind variables and return named params: https://jmoiron.github.io/sqlx/#namedParams
func (g *GeoPackage) makeFeaturesQuery(ctx context.Context, propConfig *config.FeatureProperties, table *featureTable,
	onlyFIDs bool, axisOrder domain.AxisOrder, criteria datasources.FeaturesCriteria) (stmt *sqlx.NamedStmt, query string, queryArgs map[string]any, err error) {

	var selectClause string
	if onlyFIDs {
		selectClause = columnsToSQL([]string{g.fidColumn, domain.PrevFid, domain.NextFid})
	} else {
		selectClause = g.selectColumns(table, axisOrder, propConfig, true)
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

func (g *GeoPackage) makeDefaultQuery(table *featureTable, selectClause string, criteria datasources.FeaturesCriteria) (string, map[string]any) {
	pfClause, pfNamedParams := propertyFiltersToSQL(criteria.PropertyFilters)
	temporalClause, temporalNamedParams := temporalCriteriaToSQL(criteria.TemporalCriteria)

	defaultQuery := fmt.Sprintf(`
with
    next as (select * from "%[1]s" where "%[2]s" >= :fid %[3]s %[4]s order by %[2]s asc limit :limit + 1),
    prev as (select * from "%[1]s" where "%[2]s" < :fid %[3]s %[4]s order by %[2]s desc limit :limit),
    nextprev as (select * from next union all select * from prev),
    nextprevfeat as (select *, lag("%[2]s", :limit) over (order by %[2]s) as %[6]s, lead("%[2]s", :limit) over (order by "%[2]s") as %[7]s from nextprev)
select %[5]s from nextprevfeat where "%[2]s" >= :fid %[3]s %[4]s limit :limit
`, table.TableName, g.fidColumn, temporalClause, pfClause, selectClause, domain.PrevFid, domain.NextFid) // don't add user input here, use named params for user input!

	namedParams := map[string]any{
		"fid":   criteria.Cursor.FID,
		"limit": criteria.Limit,
	}
	maps.Copy(namedParams, pfNamedParams)
	maps.Copy(namedParams, temporalNamedParams)
	return defaultQuery, namedParams
}

func (g *GeoPackage) makeBboxQuery(table *featureTable, selectClause string, criteria datasources.FeaturesCriteria) (string, map[string]any, error) {
	btreeIndexHint := fmt.Sprintf("indexed by \"%s_spatial_idx\"", table.TableName)

	pfClause, pfNamedParams := propertyFiltersToSQL(criteria.PropertyFilters)
	if pfClause != "" {
		// don't force btree index when using property filter, let SQLite decide
		// whether to use the BTree index or the property filter index
		btreeIndexHint = ""
	}
	temporalClause, temporalNamedParams := temporalCriteriaToSQL(criteria.TemporalCriteria)

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
`, table.TableName, g.fidColumn, g.maxBBoxSizeToUseWithRTree, table.GeometryColumnName,
		selectClause, temporalClause, pfClause, btreeIndexHint, domain.PrevFid, domain.NextFid) // don't add user input here, use named params for user input!

	bboxAsWKT, err := wkt.Marshal(criteria.Bbox.Polygon())
	if err != nil {
		return "", nil, err
	}
	namedParams := map[string]any{
		"fid":            criteria.Cursor.FID,
		"limit":          criteria.Limit,
		"bboxWkt":        bboxAsWKT,
		domain.MaxxField: criteria.Bbox.Max(0),
		domain.MinxField: criteria.Bbox.Min(0),
		domain.MaxyField: criteria.Bbox.Max(1),
		domain.MinyField: criteria.Bbox.Min(1),
		"bboxSrid":       criteria.InputSRID}
	maps.Copy(namedParams, pfNamedParams)
	maps.Copy(namedParams, temporalNamedParams)
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

// selectColumns build select clause
func (g *GeoPackage) selectColumns(table *featureTable, axisOrder domain.AxisOrder,
	propConfig *config.FeatureProperties, includePrevNext bool) string {

	columns := orderedmap.New[string, struct{}]() // map (actually a set) to prevent accidental duplicate columns
	switch {
	case propConfig != nil:
		// select columns in a specific order (we need an ordered map for this purpose!)
		for _, prop := range propConfig.Properties {
			if prop != table.GeometryColumnName {
				columns.Set(prop, struct{}{})
			}
		}
		if !propConfig.PropertiesExcludeUnknown {
			// select missing columns according to the table schema
			for _, field := range table.Schema.Fields {
				if field.Name != table.GeometryColumnName {
					_, ok := columns.Get(field.Name)
					if !ok {
						columns.Set(field.Name, struct{}{})
					}
				}
			}
		}
	case table.Schema != nil:
		// select all columns according to the table schema
		for _, field := range table.Schema.Fields {
			if field.Name != table.GeometryColumnName {
				columns.Set(field.Name, struct{}{})
			}
		}
	default:
		log.Println("Warning: table doesn't have a schema. Can't select columns by name, selecting all")
		return selectAll
	}

	columns.Set(g.fidColumn, struct{}{})
	if includePrevNext {
		columns.Set(domain.PrevFid, struct{}{})
		columns.Set(domain.NextFid, struct{}{})
	}

	result := columnsToSQL(slices.Collect(columns.KeysFromOldest()))

	// Add the geometry column. GeoPackage geometries are stored in WKB format and WKB is always XY.
	// So swap coordinates when needed. This requires casting to a SpatiaLite geometry first, executing
	// the swap and then casting back to a GeoPackage geometry.
	if axisOrder == domain.AxisOrderYX {
		result += fmt.Sprintf(", asgpb(swapcoords(castautomagic(\"%[1]s\"))) as \"%[1]s\"", table.GeometryColumnName)
	} else {
		result += fmt.Sprintf(", \"%s\"", table.GeometryColumnName)
	}

	return result
}

// mapGpkgGeometry GeoPackage specific way to read geometries
func mapGpkgGeometry(rawGeom []byte) (geom.T, error) {
	geomWithMetadata, err := encoding.DecodeGeometry(rawGeom)
	if err != nil {
		return nil, err
	}
	if geomWithMetadata == nil || geomWithMetadata.Geometry.Empty() {
		return nil, nil
	}
	return geomWithMetadata.Geometry, nil
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

func temporalCriteriaToSQL(temporalCriteria datasources.TemporalCriteria) (sql string, namedParams map[string]any) {
	namedParams = make(map[string]any)
	if !temporalCriteria.ReferenceDate.IsZero() {
		namedParams["referenceDate"] = temporalCriteria.ReferenceDate
		startDate := temporalCriteria.StartDateProperty
		endDate := temporalCriteria.EndDateProperty
		sql = fmt.Sprintf(" and \"%[1]s\" <= :referenceDate and (\"%[2]s\" >= :referenceDate or \"%[2]s\" is null)", startDate, endDate)
	}
	return sql, namedParams
}

func cacheFeatureProperties(collections config.GeoSpatialCollections) map[string]*config.FeatureProperties {
	result := make(map[string]*config.FeatureProperties)
	for _, collection := range collections {
		if collection.Features == nil {
			continue
		}
		result[collection.ID] = collection.Features.FeatureProperties
	}
	return result
}

func columnsToSQL(columns []string) string {
	return fmt.Sprintf("\"%s\"", strings.Join(columns, `", "`))
}
