package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"strings"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twpayne/go-geom"
	pgxgeom "github.com/twpayne/pgx-geom"
)

const (
	selectAll = "*"

	// https://github.com/jackc/pgx/issues/387#issuecomment-1107666716
	pgxNamedParamSymbol = '@'
)

type Postgres struct {
	db *pgxpool.Pool

	schemaName        string
	fidColumn         string
	externalFidColumn string
	queryTimeout      time.Duration

	featureTableByCollectionID    map[string]*featureTable
	propertyFiltersByCollectionID map[string]datasources.PropertyFiltersWithAllowedValues
	propertiesByCollectionID      map[string]*config.FeatureProperties
}

func NewPostgres(collections config.GeoSpatialCollections, pgConfig config.Postgres, transformOnTheFly bool) (*Postgres, error) {
	if !transformOnTheFly {
		return nil, errors.New("ahead-of-time transformed features are currently not " +
			"supported for postgresql, reprojection/transformation is always applied")
	}

	pgxConfig, err := pgxpool.ParseConfig(pgConfig.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// set connection to read-only for safety since we (should) never write to Postgres.
	pgxConfig.ConnConfig.RuntimeParams["default_transaction_read_only"] = "on"
	// add support for Go <-> PostGIS conversions
	pgxConfig.AfterConnect = pgxgeom.Register

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
		db:                       db,
		schemaName:               pgConfig.Schema,
		fidColumn:                pgConfig.Fid,
		externalFidColumn:        pgConfig.ExternalFid,
		queryTimeout:             pgConfig.QueryTimeout.Duration,
		propertiesByCollectionID: collections.FeaturePropertiesByID(),
	}

	pg.featureTableByCollectionID, pg.propertyFiltersByCollectionID = readMetadata(
		db, collections, pg.fidColumn, pg.externalFidColumn, pg.schemaName)

	return pg, nil
}

func (pg Postgres) Close() {
	pg.db.Close()
}

func (pg Postgres) GetFeatureIDs(_ context.Context, _ string, _ datasources.FeaturesCriteria) ([]int64, domain.Cursors, error) {
	log.Println("Postgres support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return []int64{}, domain.Cursors{}, nil
}

func (pg Postgres) GetFeaturesByID(_ context.Context, _ string, _ []int64, _ domain.AxisOrder, _ domain.Profile) (*domain.FeatureCollection, error) {
	log.Println("Postgres support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return &domain.FeatureCollection{}, nil
}

func (pg Postgres) GetFeatures(ctx context.Context, collection string, criteria datasources.FeaturesCriteria,
	axisOrder domain.AxisOrder, profile domain.Profile) (*domain.FeatureCollection, domain.Cursors, error) {

	table, err := pg.getFeatureTable(collection)
	if err != nil {
		return nil, domain.Cursors{}, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, pg.queryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	propConfig := pg.propertiesByCollectionID[collection]
	query, queryArgs, err := pg.makeFeaturesQuery(queryCtx, propConfig, table, false, axisOrder, criteria)
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("failed to create query '%s' error: %w", query, err)
	}

	rows, err := pg.db.Query(queryCtx, query, queryArgs)
	if err != nil {
		return nil, domain.Cursors{}, fmt.Errorf("failed to execute query '%s' error: %w", query, err)
	}
	defer rows.Close()

	var prevNext *domain.PrevNextFID
	fc := domain.FeatureCollection{}
	fc.Features, prevNext, err = domain.MapRowsToFeatures(queryCtx, FromPgxRows(rows), pg.fidColumn, pg.externalFidColumn, table.GeometryColumnName,
		propConfig, table.Schema, mapPostgisGeometry, profile.MapRelationUsingProfile)
	if err != nil {
		return nil, domain.Cursors{}, err
	}
	if prevNext == nil {
		return nil, domain.Cursors{}, nil
	}
	fc.NumberReturned = len(fc.Features)
	return &fc, domain.NewCursors(*prevNext, criteria.Cursor.FiltersChecksum), queryCtx.Err()
}

func (pg Postgres) GetFeature(_ context.Context, _ string, _ any, _ domain.AxisOrder, _ domain.Profile) (*domain.Feature, error) {
	log.Println("Postgres support is not implemented yet, this just serves to demonstrate that we can support multiple types of datasources")
	return nil, nil
}

func (pg Postgres) GetSchema(collection string) (*domain.Schema, error) {
	table, err := pg.getFeatureTable(collection)
	if err != nil {
		return nil, err
	}
	return table.Schema, nil
}

func (pg Postgres) GetPropertyFiltersWithAllowedValues(collection string) datasources.PropertyFiltersWithAllowedValues {
	return pg.propertyFiltersByCollectionID[collection]
}

// Build specific features queries based on the given options.
func (pg Postgres) makeFeaturesQuery(_ context.Context, propConfig *config.FeatureProperties, table *featureTable,
	onlyFIDs bool, axisOrder domain.AxisOrder, criteria datasources.FeaturesCriteria) (query string, queryArgs pgx.NamedArgs, err error) {

	var selectClause string
	if onlyFIDs {
		selectClause = columnsToSQL([]string{pg.fidColumn, domain.PrevFid, domain.NextFid})
	} else {
		selectClause = pg.selectColumns(table, axisOrder, propConfig, true)
	}

	// make query
	if criteria.Bbox != nil {
		// TODO
		//query, queryArgs, err = pg.makeBboxQuery(table, selectClause, criteria)
		//if err != nil {
		//	return
		//}
	} else {
		query, queryArgs = pg.makeDefaultQuery(table, selectClause, criteria)
	}
	return
}

func (pg Postgres) makeDefaultQuery(table *featureTable, selectClause string, criteria datasources.FeaturesCriteria) (string, map[string]any) {
	pfClause, pfNamedParams := propertyFiltersToSQL(criteria.PropertyFilters, pgxNamedParamSymbol)
	temporalClause, temporalNamedParams := temporalCriteriaToSQL(criteria.TemporalCriteria, pgxNamedParamSymbol)

	defaultQuery := fmt.Sprintf(`
with
    next as (select * from "%[1]s" where "%[2]s" >= @fid %[3]s %[4]s order by %[2]s asc limit @limit + 1),
    prev as (select * from "%[1]s" where "%[2]s" < @fid %[3]s %[4]s order by %[2]s desc limit @limit),
    nextprev as (select * from next union all select * from prev),
    nextprevfeat as (select *, lag("%[2]s", @limit) over (order by %[2]s) as %[6]s, lead("%[2]s", @limit) over (order by "%[2]s") as %[7]s from nextprev)
select %[5]s from nextprevfeat where "%[2]s" >= @fid %[3]s %[4]s limit @limit
`, table.TableName, pg.fidColumn, temporalClause, pfClause, selectClause, domain.PrevFid, domain.NextFid) // don't add user input here, use named params for user input!

	namedParams := map[string]any{
		"fid":   criteria.Cursor.FID,
		"limit": criteria.Limit,
	}
	maps.Copy(namedParams, pfNamedParams)
	maps.Copy(namedParams, temporalNamedParams)
	return defaultQuery, namedParams
}

// selectColumns build select clause
func (pg Postgres) selectColumns(table *featureTable, axisOrder domain.AxisOrder,
	propConfig *config.FeatureProperties, includePrevNext bool) string {

	columns := make(map[string]struct{}) // map (actually a set) to prevent accidental duplicate columns
	switch {
	case propConfig != nil:
		// select columns in a specific order
		for _, prop := range propConfig.Properties {
			if prop != table.GeometryColumnName {
				columns[prop] = struct{}{}
			}
		}
	case table.Schema != nil:
		// select all columns according to the table schema
		for _, field := range table.Schema.Fields {
			if field.Name != table.GeometryColumnName {
				columns[field.Name] = struct{}{}
			}
		}
	default:
		log.Println("Warning: table doesn't have a schema. Can't select columns by name, selecting all")
		return selectAll
	}

	columns[pg.fidColumn] = struct{}{}
	if includePrevNext {
		columns[domain.PrevFid] = struct{}{}
		columns[domain.NextFid] = struct{}{}
	}

	result := columnsToSQL(util.Keys(columns))

	// Add the geometry column. GeoPackage geometries are stored in WKB format and WKB is always XY.
	// So swap coordinates when needed. This requires casting to a SpatiaLite geometry first, executing
	// the swap and then casting back to a GeoPackage geometry.
	if axisOrder == domain.AxisOrderYX {
		result += fmt.Sprintf(", st_flipcoordinates(\"%[1]s\") as \"%[1]s\"", table.GeometryColumnName)
	} else {
		result += fmt.Sprintf(", \"%s\"", table.GeometryColumnName)
	}

	if propConfig != nil && !propConfig.PropertiesExcludeUnknown {
		result += ", " + selectAll
	}
	return result
}

func (pg Postgres) getFeatureTable(collection string) (*featureTable, error) {
	table, ok := pg.featureTableByCollectionID[collection]
	if !ok {
		return nil, fmt.Errorf("can't query collection '%s' since it doesn't exist in "+
			"postgresql, available in postgresql: %v", collection, util.Keys(pg.featureTableByCollectionID))
	}
	return table, nil
}

// mapPostgisGeometry Postgres/PostGIS specific way to read geometries.
// since we use 'pgx-geom' it's just a simple cast since conversion happens automatically.
func mapPostgisGeometry(columnValue any) (geom.T, error) {
	geometry, ok := columnValue.(geom.T)
	if !ok {
		return nil, errors.New("failed to convert column value to geometry")
	}
	return geometry, nil
}

func propertyFiltersToSQL(pf map[string]string, symbol byte) (sql string, namedParams map[string]any) {
	namedParams = make(map[string]any)
	if len(pf) > 0 {
		position := 0
		for k, v := range pf {
			position++
			namedParam := fmt.Sprintf("pf%d", position)
			// column name in double quotes in case it is a reserved keyword
			// also: we don't currently support LIKE since wildcard searches don't use the index
			sql += fmt.Sprintf(" and \"%s\" = %b%s", k, symbol, namedParam)
			namedParams[namedParam] = v
		}
	}
	return sql, namedParams
}

func temporalCriteriaToSQL(temporalCriteria datasources.TemporalCriteria, symbol byte) (sql string, namedParams map[string]any) {
	namedParams = make(map[string]any)
	if !temporalCriteria.ReferenceDate.IsZero() {
		namedParams["referenceDate"] = temporalCriteria.ReferenceDate
		startDate := temporalCriteria.StartDateProperty
		endDate := temporalCriteria.EndDateProperty
		sql = fmt.Sprintf(" and \"%[1]s\" <= %[3]breferenceDate and (\"%[2]s\" >= %[3]breferenceDate or \"%[2]s\" is null)",
			startDate, endDate, symbol)
	}
	return sql, namedParams
}

func columnsToSQL(columns []string) string {
	return fmt.Sprintf("\"%s\"", strings.Join(columns, `", "`))
}
