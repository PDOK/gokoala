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
	search "github.com/PDOK/gokoala/internal/search/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
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
	relationsConfig := pg.RelationsByCollectionID[collection]
	query, queryArgs, err := pg.makeFeaturesQuery(propConfig, relationsConfig, table, false, axisOrder, criteria)
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
	relationsConfig := pg.RelationsByCollectionID[collection]
	selectClause := pg.SelectColumns(table, axisOrder, selectPostGISGeometry, selectPostgresRelation,
		propConfig, relationsConfig, false)

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

func (pg *Postgres) SearchFeaturesAcrossCollections(ctx context.Context, criteria ds.FeaturesSearchCriteria,
	collections search.CollectionsWithParams) (*d.FeatureCollection, error) {

	queryCtx, cancel := context.WithTimeout(ctx, pg.QueryTimeout) // https://go.dev/doc/database/cancel-operations
	defer cancel()

	// TODO: find better place for this srid logic
	if criteria.InputSRID == d.UndefinedSRID || criteria.InputSRID == d.WGS84SRID {
		criteria.InputSRID = d.WGS84SRIDPostgis
	}
	if criteria.OutputSRID == d.UndefinedSRID || criteria.OutputSRID == d.WGS84SRID {
		criteria.OutputSRID = d.WGS84SRIDPostgis
	}

	bboxFilter, bboxQueryArgs, err := bboxToSQL(criteria.Bbox, criteria.InputSRID, "r.geom", "r.bbox")
	if err != nil {
		return nil, err
	}
	sql := makeSearchQuery(criteria.Settings.IndexName, criteria.OutputSRID, bboxFilter)
	wildcardQuery := criteria.SearchQuery.ToWildcardQuery()
	exactMatchQuery := criteria.SearchQuery.ToExactMatchQuery(criteria.Settings.SynonymsExactMatch)
	names, versions, relevance := collections.NamesAndVersionsAndRelevance()
	log.Printf("\nSEARCH QUERY (wildcard): %s\n", wildcardQuery)

	// Create query params
	namedParams := map[string]any{
		"lm":              criteria.Limit,
		"wildcardquery":   wildcardQuery,
		"exactmatchquery": exactMatchQuery,
		"names":           names,
		"versions":        versions,
		"relevance":       relevance,
		"rn":              criteria.Settings.RankNormalization,
		"emm":             criteria.Settings.ExactMatchMultiplier,
		"psm":             criteria.Settings.PrimarySuggestMultiplier,
		"rt":              criteria.Settings.RankThreshold,
		"prlm":            criteria.Settings.PreRankLimitMultiplier,
		"prwcc":           criteria.Settings.PreRankWordCountCutoff,
	}
	maps.Copy(namedParams, bboxQueryArgs)

	// Execute search query
	rows, err := pg.db.Query(queryCtx, sql, pgx.NamedArgs(namedParams))
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", sql, err)
	}
	defer rows.Close()

	// Turn rows into FeatureCollection
	return pg.mapRowsToFeatures(queryCtx, rows)
}

//nolint:funlen
func makeSearchQuery(index string, srid d.SRID, bboxFilter string) string {
	// language=postgresql
	return fmt.Sprintf(
		`WITH query_wildcard AS (
		SELECT to_tsquery('custom_dict', @wildcardquery) query
	),
	query_exact AS (
		SELECT to_tsquery('custom_dict', @exactmatchquery) query
	),
	results AS NOT MATERIALIZED ( -- the results query is called twice, materializing it results in a non optimal query plan for one of the calls
		SELECT
			r.display_name,
			r.feature_id,
			r.external_fid,
			r.collection_id,
			r.collection_version,
			r.geometry_type,
			r.bbox,
			r.geometry,
			r.suggest,
			r.ts
		FROM
			%[1]s r
		WHERE
			r.ts @@ (SELECT query FROM query_wildcard) AND (r.collection_id, r.collection_version) IN (
				-- make a virtual table by creating tuples from the provided arrays.
				SELECT * FROM unnest(@names::text[], @versions::int[])
			)
		%[3]s -- bounding box intersect filter
	),
	results_count AS (
	    SELECT
	    	COUNT(*) c
	    FROM (
	        SELECT
				r.feature_id
	        FROM
				results r
	        LIMIT @rt
	    ) as rc
	)
	SELECT
	    rn.display_name,
		rn.feature_id,
		rn.external_fid,
		rn.collection_id,
		rn.collection_version,
		rn.geometry_type,
		st_transform(rn.bbox, %[2]d)::geometry AS bbox,
		st_transform(rn.geometry, %[2]d)::geometry AS geometry,
		rn.rank,
		ts_headline('custom_dict', rn.suggest, (SELECT query FROM query_wildcard)) AS highlighted_text
	FROM (
		SELECT
			r.*,
			ROW_NUMBER() OVER (
				PARTITION BY
					r.display_name,
					r.collection_id,
					r.collection_version,
					r.feature_id,
					r.external_fid
				ORDER BY -- use same "order by" clause everywhere
					r.rank DESC,
					r.display_name COLLATE "custom_numeric" ASC
			) AS row_number
		FROM (
			SELECT
				u.*,
				CASE WHEN u.display_name = u.suggest THEN (
					ts_rank_cd(u.ts, (SELECT query FROM query_exact), @rn) * @emm * @psm + ts_rank_cd(u.ts, (SELECT query FROM query_wildcard), @rn)
				) * rel.relevance
				ELSE (
					ts_rank_cd(u.ts, (SELECT query FROM query_exact), @rn) * @emm + ts_rank_cd(u.ts, (SELECT query FROM query_wildcard), @rn)
				) * rel.relevance
				END AS rank
			FROM (
				-- a UNION ALL is used, because a CASE in the ORDER BY clause causes a sequence scan instead of an index scan
				-- because of 1 = 1 in the WHERE clauses below the results are only added if WHEN is true, otherwise the results are ignored
				(
					SELECT
						*
					FROM
						results r
					WHERE
						-- less then rank threshold results don't need to be pre-ranked, they can be ranked based on score
						CASE WHEN (SELECT c from results_count) < @rt THEN 1 = 1 END
				) UNION ALL (
					SELECT
						*
					FROM
						results r
					WHERE
						-- pre-rank more then rank threshold results by ordering on suggest length and display_name
						CASE WHEN (SELECT c from results_count) = @rt THEN 1 = 1 END AND
						array_length(string_to_array(r.suggest, ' '), 1) <= @prwcc
					ORDER BY
						array_length(string_to_array(r.suggest, ' '), 1) ASC,
						r.display_name COLLATE "custom_numeric" ASC
					LIMIT (@lm::int * @prlm::int) -- return limited pre-ranked results for ranking based on scor
				)
			) u
			LEFT JOIN
				(SELECT * FROM unnest(@names::text[], @relevance::float[]) rel(collection_id,relevance)) rel
			ON
				rel.collection_id = u.collection_id
		) r
	) rn
	WHERE rn.row_number = 1
	ORDER BY -- use same "order by" clause everywhere
	    rn.rank DESC,
	    rn.display_name COLLATE "custom_numeric" ASC
	LIMIT (@lm::int)`, index, srid, bboxFilter) // don't add user input here, use $X params for user input!
}

// TODO move to mapper
func (pg *Postgres) mapRowsToFeatures(queryCtx context.Context, rows pgx.Rows) (*d.FeatureCollection, error) {
	fc := d.FeatureCollection{Features: make([]*d.Feature, 0)}
	for rows.Next() {
		var displayName, highlightedText, featureID, collectionID, collectionVersion, geomType string
		var rank float64
		var bbox, geometry geom.T
		var externalFid *string

		if err := rows.Scan(&displayName, &featureID, &externalFid, &collectionID, &collectionVersion, &geomType,
			&bbox, &geometry, &rank, &highlightedText); err != nil {
			return nil, err
		}
		geojsonBbox, err := encodeBBox(bbox)
		if err != nil {
			return nil, err
		}
		geojsonGeom, err := geojson.Encode(geometry, geojson.EncodeGeometryWithMaxDecimalDigits(pg.MaxDecimals))
		if err != nil {
			return nil, err
		}
		fc.Features = append(fc.Features, &d.Feature{
			ID:       getFeatureID(externalFid, featureID),
			Geometry: geojsonGeom,
			Bbox:     geojsonBbox,
			Properties: d.NewFeaturePropertiesWithData(false, map[string]any{
				search.PropCollectionID:      collectionID,
				search.PropCollectionVersion: collectionVersion,
				search.PropGeomType:          geomType,
				search.PropDisplayName:       displayName,
				search.PropHighlight:         highlightedText,
				search.PropScore:             rank,
			}),
		})
		fc.NumberReturned = len(fc.Features)
	}
	return &fc, queryCtx.Err()
}

// TODO move to mapper
func getFeatureID(externalFid *string, featureID string) string {
	if externalFid != nil {
		return *externalFid
	}
	return featureID
}

// TODO move to mapper
// adapted from https://github.com/twpayne/go-geom/blob/b22fd061f1531a51582333b5bd45710a455c4978/encoding/geojson/geojson.go#L525
// encodeBBox encodes b as a GeoJson Bounding Box.
func encodeBBox(bbox geom.T) (*[]float64, error) {
	if bbox == nil {
		return nil, nil
	}
	b := bbox.Bounds()
	switch l := b.Layout(); l {
	case geom.XY, geom.XYM:
		return &[]float64{b.Min(0), b.Min(1), b.Max(0), b.Max(1)}, nil
	case geom.XYZ, geom.XYZM, geom.NoLayout:
		return nil, fmt.Errorf("unsupported type: %d", rune(l))
	default:
		return nil, fmt.Errorf("unsupported type: %d", rune(l))
	}
}

// Build specific features queries based on the given options.
func (pg *Postgres) makeFeaturesQuery(propConfig *config.FeatureProperties, relationsConfig []config.Relation, table *common.Table,
	onlyFIDs bool, axisOrder d.AxisOrder, criteria ds.FeaturesCriteria) (query string, queryArgs pgx.NamedArgs, err error) {

	var selectClause string
	if onlyFIDs {
		selectClause = common.ColumnsToSQL([]string{pg.FidColumn, d.PrevFid, d.NextFid}, true)
	} else {
		selectClause = pg.SelectColumns(table, axisOrder, selectPostGISGeometry, selectPostgresRelation,
			propConfig, relationsConfig, true)
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
		bboxClause, bboxNamedParams, err = bboxToSQL(criteria.Bbox, criteria.InputSRID, table.GeometryColumnName, "")
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

func bboxToSQL(bbox *geom.Bounds, bboxSRID d.SRID, geomColumn string, bboxColumn string) (string, map[string]any, error) {
	var bboxFilter, bboxWkt string
	var bboxNamedParams map[string]any
	var err error
	if bbox != nil {
		if bboxColumn == "" {
			bboxFilter = fmt.Sprintf(`and
				st_intersects(st_transform(%[1]s, @bboxSrid::int), st_geomfromtext(@bboxWkt::text, @bboxSrid::int))
			`, geomColumn)
		} else {
			bboxFilter = fmt.Sprintf(`and
				(st_intersects(st_transform(%[1]s, @bboxSrid::int), st_geomfromtext(@bboxWkt::text, @bboxSrid::int)) or
				st_intersects(st_transform(%[2]s, @bboxSrid::int), st_geomfromtext(@bboxWkt::text, @bboxSrid::int)))
			`, geomColumn, bboxColumn)
		}
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

// selectPostgresRelation Assemble Postgres specific query to select related features using a many-to-many table e.g.:
//
//	select string_agg(other.external_fid, ',')
//	from building_apartment junction join apartment other on other.id = junction.apartment_id
//	where junction.building_id = building.id
func selectPostgresRelation(relation config.Relation, relationName string, targetFID string, sourceTableAlias string) string {
	return fmt.Sprintf(`(
				select string_agg(other.%[1]s::text, ',')
				from %[2]s junction join %[4]s other on other.%[5]s::text = junction.%[6]s::text
				where junction.%[7]s::text = %[9]s.%[8]s::text
			) as %[3]s`, targetFID, relation.Junction.Name,
		relationName, relation.RelatedCollection,
		relation.Columns.Target, relation.Junction.Columns.Target,
		relation.Junction.Columns.Source, relation.Columns.Source,
		sourceTableAlias)
}
