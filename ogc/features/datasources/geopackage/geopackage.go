package geopackage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"strings"
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

func (g *GeoPackage) GetFeatureTableMetadata(collection string) (datasources.FeatureTableMetadata, error) {
	val, ok := g.featureTableByCollectionID[collection]
	if !ok {
		return nil, fmt.Errorf("no metadata for %s", collection)
	}
	return val, nil
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
		g.queryTimeout = gpkgConfig.Local.QueryTimeout
	case gpkgConfig.Cloud != nil:
		g.backend = newCloudBackedGeoPackage(gpkgConfig.Cloud)
		g.fidColumn = gpkgConfig.Cloud.Fid
		g.queryTimeout = gpkgConfig.Cloud.QueryTimeout
	default:
		log.Fatal("unknown geopackage config encountered")
	}

	metadata, err := readDriverMetadata(g.backend.getDB())
	if err != nil {
		log.Fatalf("failed to connect with geopackage: %v", err)
	}
	log.Println(metadata)

	featureTables, err := readGpkgContents(collections, g.backend.getDB())
	if err != nil {
		log.Fatal(err)
	}
	g.featureTableByCollectionID = featureTables

	// assert that an index named <table>_spatial_idx exists on each feature table with the given columns
	g.assertIndexExistOnFeatureTables("_spatial_idx",
		strings.Join([]string{g.fidColumn, "minx", "maxx", "miny", "maxy"}, ","))

	return g
}

func (g *GeoPackage) Close() {
	g.backend.close()
}

func (g *GeoPackage) GetFeatureIDs(ctx context.Context, collection string, criteria datasources.FeaturesCriteria) ([]int64, domain.Cursors, error) {
	table, err := g.getFeatureTable(collection)
	if err != nil {
		return nil, domain.Cursors{}, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, g.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	query, queryArgs, err := g.makeFeaturesQuery(table, true, criteria)
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("failed to make features query, error: %w", err)
	}

	stmt, err := g.backend.getDB().PrepareNamedContext(queryCtx, query)
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("failed to prepare query '%s' error: %w", query, err)
	}
	defer stmt.Close()

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

	query, queryArgs, err := g.makeFeaturesQuery(table, false, criteria)
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("failed to make features query, error: %w", err)
	}

	stmt, err := g.backend.getDB().PrepareNamedContext(queryCtx, query)
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("failed to prepare query '%s' error: %w", query, err)
	}
	defer stmt.Close()

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
	stmt, err := g.backend.getDB().PrepareNamedContext(queryCtx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(queryCtx, map[string]any{"fid": featureID})
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

// Build specific features queries based on the given options.
// Make sure to use SQL bind variables and return named params: https://jmoiron.github.io/sqlx/#namedParams
func (g *GeoPackage) makeFeaturesQuery(table *featureTable, onlyFIDs bool, criteria datasources.FeaturesCriteria) (string, map[string]any, error) {
	if criteria.Bbox != nil {
		return g.makeBboxQuery(table, onlyFIDs, criteria)
	}
	return g.makeDefaultQuery(table, criteria)
}

func (g *GeoPackage) makeDefaultQuery(table *featureTable, criteria datasources.FeaturesCriteria) (string, map[string]any, error) {
	pfClause, pfNamedParams := propertyFiltersToSQL(criteria.PropertyFilters)

	defaultQuery := fmt.Sprintf(`
with 
    next as (select * from %[1]s where %[2]s >= :fid %[3]s order by %[2]s asc limit :limit + 1),
    prev as (select * from %[1]s where %[2]s < :fid %[3]s order by %[2]s desc limit :limit),
    nextprev as (select * from next union all select * from prev),
    nextprevfeat as (select *, lag(%[2]s, :limit) over (order by %[2]s) as prevfid, lead(%[2]s, :limit) over (order by %[2]s) as nextfid from nextprev)
select * from nextprevfeat where %[2]s >= :fid %[3]s limit :limit
`, table.TableName, g.fidColumn, pfClause)

	namedParams := map[string]any{
		"fid":   criteria.Cursor.FID,
		"limit": criteria.Limit,
	}
	maps.Copy(namedParams, pfNamedParams)
	return defaultQuery, namedParams, nil
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
`, table.TableName, g.fidColumn, bboxSizeBig, table.GeometryColumnName, selectClause, pfClause)

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

// Read metadata about gpkg and sqlite driver
func readDriverMetadata(db *sqlx.DB) (string, error) {
	type pragma struct {
		UserVersion string `db:"user_version"`
	}
	type metadata struct {
		Sqlite     string `db:"sqlite"`
		Spatialite string `db:"spatialite"`
		Arch       string `db:"arch"`
	}

	var m metadata
	err := db.QueryRowx(`
select sqlite_version() as sqlite, 
spatialite_version() as spatialite,  
spatialite_target_cpu() as arch`).StructScan(&m)
	if err != nil {
		return "", err
	}

	var gpkgVersion pragma
	_ = db.QueryRowx(`pragma user_version`).StructScan(&gpkgVersion)
	if gpkgVersion.UserVersion == "" {
		gpkgVersion.UserVersion = "unknown"
	}

	return fmt.Sprintf("geopackage version: %s, sqlite version: %s, spatialite version: %s on %s",
		gpkgVersion.UserVersion, m.Sqlite, m.Spatialite, m.Arch), nil
}

// Assert that an index on each feature table exists with the given suffix and covering the given columns, in the given order.
func (g *GeoPackage) assertIndexExistOnFeatureTables(expectedIndexNameSuffix string, expectedIndexColumns string) {
	for _, collection := range g.featureTableByCollectionID {
		expectedIndexName := collection.TableName + expectedIndexNameSuffix
		var actualIndexColumns string

		query := fmt.Sprintf(`
select group_concat(name) 
from pragma_index_info('%s') 
order by name asc`, expectedIndexName)

		err := g.backend.getDB().QueryRowx(query).Scan(&actualIndexColumns)
		if err != nil {
			log.Fatalf("missing index: failed to read index '%s' from table '%s'",
				expectedIndexName, collection.TableName)
		}
		if expectedIndexColumns != actualIndexColumns {
			log.Fatalf("incorrect index: expected index '%s' with columns '%s' to exist on table '%s', found indexed columns '%s'",
				expectedIndexName, expectedIndexColumns, collection.TableName, actualIndexColumns)
		}
	}
}

func (g *GeoPackage) getFeatureTable(collection string) (*featureTable, error) {
	table, ok := g.featureTableByCollectionID[collection]
	if !ok {
		return nil, fmt.Errorf("can't query collection '%s' since it doesn't exist in "+
			"geopackage, available in geopackage: %v", collection, util.Keys(g.featureTableByCollectionID))
	}
	return table, nil
}

// Read gpkg_contents table. This table contains metadata about feature tables. The result is a mapping from
// collection ID -> feature table metadata. We match each feature table to the collection ID by looking at the
// 'identifier' column. Also in case there's no exact match between 'collection ID' and 'identifier' we use
// the explicitly configured table name.
func readGpkgContents(collections engine.GeoSpatialCollections, db *sqlx.DB) (map[string]*featureTable, error) {
	query := `
select
	c.table_name, c.data_type, c.identifier, c.description, c.last_change,
	c.min_x, c.min_y, c.max_x, c.max_y, c.srs_id, gc.column_name, gc.geometry_type_name
from
	gpkg_contents c join gpkg_geometry_columns gc on c.table_name == gc.table_name
where
	c.data_type = 'features' and 
	c.min_x is not null`

	rows, err := db.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve gpkg_contents using query: %v\n, error: %w", query, err)
	}
	defer rows.Close()

	result := make(map[string]*featureTable, 10)
	for rows.Next() {
		row := featureTable{
			ColumnsWithDateType: make(map[string]string),
		}
		if err = rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("failed to read gpkg_contents record, error: %w", err)
		}
		if row.TableName == "" {
			return nil, fmt.Errorf("feature table name is blank, error: %w", err)
		}
		if err = readFeatureTableInfo(db, row); err != nil {
			return nil, fmt.Errorf("failed to read feature table metadata, error: %w", err)
		}

		if len(collections) == 0 {
			result[row.Identifier] = &row
		} else {
			for _, collection := range collections {
				if row.Identifier == collection.ID {
					result[collection.ID] = &row
					break
				} else if hasMatchingTableName(collection, row) {
					result[collection.ID] = &row
					break
				}
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no records found in gpkg_contents, can't serve features")
	}
	return result, nil
}

func readFeatureTableInfo(db *sqlx.DB, table featureTable) error {
	rows, err := db.Queryx(fmt.Sprintf("select name, type from pragma_table_info('%s')", table.TableName))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var colName, colType string
		err = rows.Scan(&colName, &colType)
		if err != nil {
			return err
		}
		table.ColumnsWithDateType[colName] = colType
	}
	return nil
}

func hasMatchingTableName(collection engine.GeoSpatialCollection, row featureTable) bool {
	return collection.Features != nil && collection.Features.TableName != nil &&
		row.Identifier == *collection.Features.TableName
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
			sql += fmt.Sprintf(" and %s = :%s", k, namedParam)
			namedParams[namedParam] = v
		}
	}
	return sql, namedParams
}
