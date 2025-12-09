package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"

	"github.com/PDOK/gokoala/config"
	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/common"
	d "github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkt"
	pgxgeom "github.com/twpayne/pgx-geom"
	pgxuuid "github.com/vgarvardt/pgx-google-uuid/v5"
)

const (
	// https://github.com/jackc/pgx/issues/387#issuecomment-1107666716
	pgxNamedParamSymbol = "@"
)

type Postgres struct {
	common.DatasourceCommon

	db         *pgxpool.Pool
	schemaName string
}

func NewPostgres(collections config.GeoSpatialCollections, pgConfig config.Postgres,
	transformOnTheFly bool, maxDecimals int, forceUTC bool) (*Postgres, error) {

	if !transformOnTheFly {
		return nil, errors.New("ahead-of-time transformed features are currently not " +
			"supported for postgresql, reprojection/transformation is always applied")
	}

	pgxConfig, err := pgxpool.ParseConfig(pgConfig.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// enable SQL logging when appropriate environment variable (LOG_SQL=true) is set
	if sl := NewSQLLogFromEnv(); sl.LogSQL {
		pgxConfig.ConnConfig.Tracer = sl.Tracer
	}

	// set connection to read-only for safety since we (should) never write to Postgres.
	pgxConfig.ConnConfig.RuntimeParams["default_transaction_read_only"] = "on"

	pgxConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// add support for github.com/google/uuid <-> PostGIS conversions
		pgxuuid.Register(conn.TypeMap())
		// add support for Go <-> PostGIS conversions
		return pgxgeom.Register(ctx, conn)
	}

	ctx := context.Background()
	db, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	log.Printf("connecting to database '%s' as user '%s' on server: %s",
		pgConfig.DatabaseName, pgConfig.User, pgConfig.Host)
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to connect with database: %w", err)
	}

	pg := &Postgres{
		DatasourceCommon: common.DatasourceCommon{
			TransformOnTheFly:        transformOnTheFly,
			FidColumn:                pgConfig.Fid,
			ExternalFidColumn:        pgConfig.ExternalFid,
			QueryTimeout:             pgConfig.QueryTimeout.Duration,
			MaxDecimals:              maxDecimals,
			ForceUTC:                 forceUTC,
			PropertiesByCollectionID: collections.FeaturePropertiesByID(),
			RelationsByCollectionID:  collections.FeatureRelationsByID(),
		},
		db:         db,
		schemaName: pgConfig.Schema,
	}

	pg.TableByCollectionID, pg.PropertyFiltersByCollectionID = readMetadata(
		db, collections, pg.FidColumn, pg.ExternalFidColumn, pg.schemaName)

	if err = assertIndexesExist(collections, pg.TableByCollectionID, db, *pgConfig.SpatialIndexRequired); err != nil {
		return nil, err
	}

	return pg, nil
}

func (pg *Postgres) Close() {
	pg.db.Close()
}

func (pg *Postgres) GetFeatureIDs(_ context.Context, _ string, _ ds.FeaturesCriteria) ([]int64, d.Cursors, error) {
	return []int64{}, d.Cursors{}, errors.New("not implemented since the postgres datasource currently " +
		"only support on-the-fly transformation/reprojection, use GetFeatures() to get features in every supported CRS")
}

func (pg *Postgres) GetFeaturesByID(_ context.Context, _ string, _ []int64, _ d.AxisOrder, _ d.Profile) (*d.FeatureCollection, error) {
	return &d.FeatureCollection{}, errors.New("not implemented since the postgres datasource currently " +
		"only support on-the-fly transformation/reprojection, use GetFeatures() to get features in every supported CRS")
}

func (pg *Postgres) GetFeatures(ctx context.Context, collection string, criteria ds.FeaturesCriteria,
	axisOrder d.AxisOrder, profile d.Profile) (*d.FeatureCollection, d.Cursors, error) {

	table, err := pg.CollectionToTable(collection)
	if err != nil {
		return nil, d.Cursors{}, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, pg.QueryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	propConfig := pg.PropertiesByCollectionID[collection]
	query, queryArgs, err := pg.makeFeaturesQuery(propConfig, table, false, axisOrder, criteria)
	if err != nil {
		return nil, d.Cursors{}, fmt.Errorf("failed to create query '%s' error: %w", query, err)
	}

	rows, err := pg.db.Query(queryCtx, query, queryArgs)
	if err != nil {
		return nil, d.Cursors{}, fmt.Errorf("failed to execute query '%s' error: %w", query, err)
	}
	defer rows.Close()

	var prevNext *d.PrevNextFID
	fc := d.FeatureCollection{}
	fc.Features, prevNext, err = common.MapRowsToFeatures(queryCtx, FromPgxRows(rows),
		pg.FidColumn, pg.ExternalFidColumn, table.GeometryColumnName,
		propConfig, table.Schema, mapPostGISGeometry, profile.MapRelationUsingProfile,
		common.FormatOpts{MaxDecimals: pg.MaxDecimals, ForceUTC: pg.ForceUTC})
	if err != nil {
		return nil, d.Cursors{}, err
	}
	if prevNext == nil {
		return nil, d.Cursors{}, nil
	}
	fc.NumberReturned = len(fc.Features)

	return &fc, d.NewCursors(*prevNext, criteria.Cursor.FiltersChecksum), queryCtx.Err()
}

func (pg *Postgres) GetFeature(ctx context.Context, collection string, featureID any,
	outputSRID d.SRID, axisOrder d.AxisOrder, profile d.Profile) (*d.Feature, error) {

	table, err := pg.CollectionToTable(collection)
	if err != nil {
		return nil, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, pg.QueryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	var fidTypeCast string
	var fidColumn string
	switch featureID.(type) {
	case int64:
		if pg.ExternalFidColumn != "" {
			// Features should be retrieved by UUID
			log.Println("feature requested by int while external fid column is defined")

			return nil, nil
		}
		fidColumn = pg.FidColumn
		fidTypeCast = "::bigint" // always compare as 64-bits integer, regardless of numeric type in schema
	case uuid.UUID:
		if pg.ExternalFidColumn == "" {
			// Features should be retrieved by int64
			log.Println("feature requested by UUID while external fid column is not defined")

			return nil, nil
		}
		fidColumn = pg.ExternalFidColumn
	}

	propConfig := pg.PropertiesByCollectionID[collection]
	selectClause := pg.SelectColumns(table, axisOrder, selectPostGISGeometry, propConfig, nil, false)

	// TODO: find better place for this srid logic
	srid := outputSRID.GetOrDefault()
	if srid == d.UndefinedSRID || srid == d.WGS84SRID {
		srid = d.WGS84SRIDPostgis
	}

	query := fmt.Sprintf(`select %[1]s from "%[2]s" where "%[3]s"%[4]s = @fid%[4]s limit 1`,
		selectClause, table.Name, fidColumn, fidTypeCast)
	rows, err := pg.db.Query(queryCtx, query, pgx.NamedArgs{"fid": featureID, "outputSrid": srid})
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	features, _, err := common.MapRowsToFeatures(queryCtx, FromPgxRows(rows),
		pg.FidColumn, pg.ExternalFidColumn, table.GeometryColumnName,
		propConfig, table.Schema, mapPostGISGeometry, profile.MapRelationUsingProfile,
		common.FormatOpts{MaxDecimals: pg.MaxDecimals, ForceUTC: pg.ForceUTC})
	if err != nil {
		return nil, err
	}
	if len(features) != 1 {
		return nil, nil
	}

	return features[0], queryCtx.Err()
}

// Build specific features queries based on the given options.
func (pg *Postgres) makeFeaturesQuery(propConfig *config.FeatureProperties, table *common.Table,
	onlyFIDs bool, axisOrder d.AxisOrder, criteria ds.FeaturesCriteria) (query string, queryArgs pgx.NamedArgs, err error) {

	var selectClause string
	if onlyFIDs {
		selectClause = common.ColumnsToSQL([]string{pg.FidColumn, d.PrevFid, d.NextFid}, true)
	} else {
		selectClause = pg.SelectColumns(table, axisOrder, selectPostGISGeometry, propConfig, nil, true)
	}

	// TODO: find better place for this srid logic
	if criteria.InputSRID == d.UndefinedSRID || criteria.InputSRID == d.WGS84SRID {
		criteria.InputSRID = d.WGS84SRIDPostgis
	}
	if criteria.OutputSRID == d.UndefinedSRID || criteria.OutputSRID == d.WGS84SRID {
		criteria.OutputSRID = d.WGS84SRIDPostgis
	}

	return pg.makeQuery(table, selectClause, criteria)
}

func (pg *Postgres) makeQuery(table *common.Table, selectClause string, criteria ds.FeaturesCriteria) (string, map[string]any, error) {
	pfClause, pfNamedParams := common.PropertyFiltersToSQL(criteria.PropertyFilters, pgxNamedParamSymbol)
	temporalClause, temporalNamedParams := common.TemporalCriteriaToSQL(criteria.TemporalCriteria, pgxNamedParamSymbol)

	var bboxClause string
	var bboxNamedParams map[string]any
	if criteria.Bbox != nil {
		var err error
		bboxClause, bboxNamedParams, err = bboxToSQL(criteria.Bbox, criteria.InputSRID, table.GeometryColumnName)
		if err != nil {
			return "", nil, err
		}
	}

	query := fmt.Sprintf(`
with
    next as (select * from "%[1]s" where "%[2]s" >= @fid %[3]s %[4]s %[8]s order by %[2]s asc limit @limit + 1),
    prev as (select * from "%[1]s" where "%[2]s" < @fid %[3]s %[4]s %[8]s order by %[2]s desc limit @limit),
    nextprev as (select * from next union all select * from prev),
    nextprevfeat as (select *, lag("%[2]s", @limit) over (order by %[2]s) as %[6]s, lead("%[2]s", @limit) over (order by "%[2]s") as %[7]s from nextprev)
select %[5]s from nextprevfeat where "%[2]s" >= @fid %[3]s %[4]s limit @limit
`, table.Name, pg.FidColumn, temporalClause, pfClause, selectClause, d.PrevFid, d.NextFid, bboxClause)

	namedParams := map[string]any{
		"fid":        criteria.Cursor.FID,
		"limit":      criteria.Limit,
		"outputSrid": criteria.OutputSRID,
	}
	if criteria.Bbox != nil {
		maps.Copy(namedParams, bboxNamedParams)
	}
	maps.Copy(namedParams, pfNamedParams)
	maps.Copy(namedParams, temporalNamedParams)

	return query, namedParams, nil
}

func bboxToSQL(bbox *geom.Bounds, bboxSRID d.SRID, geomColumn string) (string, map[string]any, error) {
	var bboxFilter, bboxWkt string
	var bboxNamedParams map[string]any
	var err error
	if bbox != nil {
		bboxFilter = fmt.Sprintf(`and
			st_intersects(st_transform(%[1]s, @bboxSrid::int), st_geomfromtext(@bboxWkt::text, @bboxSrid::int))
		`, geomColumn)
		bboxWkt, err = wkt.Marshal(bbox.Polygon())
		if err != nil {
			return "", nil, err
		}
		bboxNamedParams = map[string]any{
			"bboxWkt":  bboxWkt,
			"bboxSrid": bboxSRID,
		}
	}

	return bboxFilter, bboxNamedParams, err
}

// mapPostGISGeometry Postgres/PostGIS specific way to read geometries into a geom.T.
// since we use 'pgx-geom' it's just a simple cast since conversion happens automatically.
func mapPostGISGeometry(columnValue any) (geom.T, error) {
	geometry, ok := columnValue.(geom.T)
	if !ok {
		return nil, errors.New("failed to convert column value to geometry")
	}

	return geometry, nil
}

// selectPostGISGeometry Postgres/PostGIS specific way to select geometry
// and take domain.AxisOrder into account.
func selectPostGISGeometry(axisOrder d.AxisOrder, table *common.Table) string {
	if axisOrder == d.AxisOrderYX {
		return fmt.Sprintf(", st_flipcoordinates(st_transform(\"%[1]s\", @outputSrid::int)) as \"%[1]s\"", table.GeometryColumnName)
	}

	return fmt.Sprintf(", st_transform(\"%[1]s\", @outputSrid::int) as \"%[1]s\"", table.GeometryColumnName)
}
