package cql

import (
	"database/sql"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pwd string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	pwd = path.Dir(filename)
}

var once sync.Once

func loadExtensions() {
	once.Do(func() {
		spatialite := path.Join(os.Getenv("SPATIALITE_LIBRARY_PATH"), "mod_spatialite")
		driver := &sqlite3.SQLiteDriver{Extensions: []string{spatialite}}
		sql.Register("sqlite_spatialite", driver)
	})
}

func TestInvalidBooleanQuery(t *testing.T) {
	// given
	inputCQL := "prop1 ==== 1 AND prop2 !!= 5"

	// when
	_, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, []domain.Field{}, 0))

	// then
	require.ErrorContains(t, err, "syntax error at column 7: mismatched input '=' expecting ")
	require.ErrorContains(t, err, "syntax error at column 23: no viable alternative at input 'prop2!'")
}

func TestFailOnNonQueryablePropertyQuery(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}}
	inputCQL := "prop1 = 30 AND prop2 > 77"

	// when
	_, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	assert.ErrorContains(t, err, "property 'prop2' cannot be used in CQL filter, is not a queryable property")
}

func TestPreventSQLInjectionAttack(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}}
	inputCQL := "prop1 > 5 OR 1 = 1"
	expectedSQL := "(\"prop1\" > :cql_bcde OR :cql_fghi = :cql_jklm)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": int64(5), "cql_fghi": int64(1), "cql_jklm": int64(1)}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestPreventSQLInjectionAttackAdvanced(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "prop5 = 'Square';DROP TABLE cql"

	// when
	_, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	assert.ErrorContains(t, err, "syntax error at column 16")
}

func TestBooleanQueryWithNumbers(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 = 10 AND prop2 < 5"
	expectedSQL := "(\"prop1\" = :cql_bcde AND \"prop2\" < :cql_fghi)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": int64(10), "cql_fghi": int64(5)}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestAllSimpleComparisionOperators(t *testing.T) {
	// given
	operators := []string{"=", "<", ">", "<=", ">=", "<>"} // note '!=' is not valid CQL, but '<>' is used by CQL.
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}

	for _, operator := range operators {
		t.Run(operator, func(t *testing.T) {
			// when
			expectedSQL := "\"prop1\" " + operator + " :cql_bcde"
			actual, err := ParseToSQL("prop1 "+operator+" 10", NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

			// then
			require.NoError(t, err)
			assertValidSQLiteQuery(t, actual)
			assert.Equal(t, map[string]any{"cql_bcde": int64(10)}, actual.Params)
			assert.Equal(t, expectedSQL, actual.SQL)
		})
	}
}

func TestMultipleBooleanQueries(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "(prop1 = 10 OR prop1 = 20) AND NOT (prop2 = 'X')"
	expectedSQL := "((\"prop1\" = :cql_bcde OR \"prop1\" = :cql_fghi) AND NOT (\"prop2\" = :cql_jklm))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": int64(10), "cql_fghi": int64(20), "cql_jklm": "'X'"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestBooleanTrueLiteral(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "(prop1 = true AND prop2 = 20)"
	expectedSQL := "(\"prop1\" = 1 AND \"prop2\" = :cql_bcde)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": int64(20)}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestBooleanFalseLiteral(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "(prop1 = false AND prop2 = 20)"
	expectedSQL := "(\"prop1\" = 0 AND \"prop2\" = :cql_bcde)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": int64(20)}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestMultipleBooleanQueriesWithStrings(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}, {Name: "prop3"}}
	inputCQL := "(prop1 = 'foo' AND prop2 = 'bar') OR prop3 = 'abc'"
	expectedSQL := "((\"prop1\" = :cql_bcde AND \"prop2\" = :cql_fghi) OR \"prop3\" = :cql_jklm)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": "'foo'", "cql_fghi": "'bar'", "cql_jklm": "'abc'"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestLikeOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}, {Name: "prop3"}}
	inputCQL := "prop1 LIKE 'foo%' AND prop2 LIKE 'bar_' OR prop3 LIKE '%abc'"
	expectedSQL := "((\"prop1\" LIKE :cql_bcde AND \"prop2\" LIKE :cql_fghi) OR \"prop3\" LIKE :cql_jklm)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": "'foo%'", "cql_fghi": "'bar_'", "cql_jklm": "'%abc'"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestNotLikeOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}, {Name: "prop3"}}
	inputCQL := "prop1 NOT LIKE 'foo%' AND prop2 LIKE 'bar_' OR prop3 LIKE '%abc'"
	expectedSQL := "((\"prop1\" NOT LIKE :cql_bcde AND \"prop2\" LIKE :cql_fghi) OR \"prop3\" LIKE :cql_jklm)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": "'foo%'", "cql_fghi": "'bar_'", "cql_jklm": "'%abc'"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestLikeOperatorFailOnMissingWildcard(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}}
	inputCQL := "prop1 LIKE 'foo'"

	// when
	_, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	assert.ErrorContains(t, err, "LIKE pattern is missing wildcard symbol. "+
		"Either percentage '%' to match multiple characters or underscore '_' to match a "+
		"single character can be used as a wildcard symbol. For example: LIKE 'foo%'.")
}

func TestBetweenOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 BETWEEN 4 AND 6 AND prop2 = 'bar'"
	expectedSQL := "(\"prop1\" BETWEEN :cql_bcde AND :cql_fghi AND \"prop2\" = :cql_jklm)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": int64(4), "cql_fghi": int64(6), "cql_jklm": "'bar'"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestNotBetweenOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 NOT BETWEEN 4 AND 6 AND prop2 = 'bar'"
	expectedSQL := "(\"prop1\" NOT BETWEEN :cql_bcde AND :cql_fghi AND \"prop2\" = :cql_jklm)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": int64(4), "cql_fghi": int64(6), "cql_jklm": "'bar'"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestInListOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 IN ('foo', 'bar', 'baz') AND prop2 = 'baz'"
	expectedSQL := "(\"prop1\" IN (:cql_bcde, :cql_fghi, :cql_jklm) AND \"prop2\" = :cql_nopq)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": "'foo'", "cql_fghi": "'bar'", "cql_jklm": "'baz'", "cql_nopq": "'baz'"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestNotInListOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 NOT IN ('foo', 'bar', 'baz') AND prop2 = 'baz'"
	expectedSQL := "(\"prop1\" NOT IN (:cql_bcde, :cql_fghi, :cql_jklm) AND \"prop2\" = :cql_nopq)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": "'foo'", "cql_fghi": "'bar'", "cql_jklm": "'baz'", "cql_nopq": "'baz'"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestIsNullOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 IS NULL AND prop2 = 'baz'"
	expectedSQL := "(\"prop1\" IS NULL AND \"prop2\" = :cql_bcde)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": "'baz'"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestIsNotNullOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 IS NOT NULL AND prop2 = 'baz'"
	expectedSQL := "(\"prop1\" IS NOT NULL AND \"prop2\" = :cql_bcde)"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actual)
	assert.Equal(t, map[string]any{"cql_bcde": "'baz'"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestFailOnInvalidInListQuery(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 IN ('foo', 'bar' 'baz')"

	// when
	_, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	assert.ErrorContains(t, err, "syntax error at column 23: extraneous input ''baz'' expecting {')', ','}")
}

func TestSpatialQueryWithPoint(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, POINT(4.897 52.377))"
	expectedSQL := "ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 4326))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "POINT(4.897 52.377)"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestSpatialQueryWithPoint3D(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, POINTZ(4.897 52.377 10.0))"
	expectedSQL := "ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 4326))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "POINTZ(4.897 52.377 10.0)"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestSpatialQueryWithLinestring(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, LINESTRING(0.0 0.0, 1.0 1.0, 2.0 0.0))"
	expectedSQL := "ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 4326))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "LINESTRING(0.0 0.0, 1.0 1.0, 2.0 0.0)"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestSpatialQueryWithPolygon(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, POLYGON((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0)))"
	expectedSQL := "ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 28992))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 28992))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "POLYGON((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0))"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestSpatialQueryWithPolygonWithHole(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, POLYGON((0.0 0.0, 10.0 0.0, 10.0 10.0, 0.0 10.0, 0.0 0.0),(2.0 2.0, 8.0 2.0, 8.0 8.0, 2.0 8.0, 2.0 2.0)))"
	expectedSQL := "ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 4326))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "POLYGON((0.0 0.0, 10.0 0.0, 10.0 10.0, 0.0 10.0, 0.0 0.0), (2.0 2.0, 8.0 2.0, 8.0 8.0, 2.0 8.0, 2.0 2.0))"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestSpatialQueryWithMultiPoint(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, MULTIPOINT(0.0 0.0, 1.0 1.0))"
	expectedSQL := "ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 4326))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "MULTIPOINT(0.0 0.0, 1.0 1.0)"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestSpatialQueryWithMultiLinestring(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, MULTILINESTRING((0.0 0.0, 1.0 1.0),(2.0 2.0, 3.0 3.0)))"
	expectedSQL := "ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 4326))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "MULTILINESTRING((0.0 0.0, 1.0 1.0), (2.0 2.0, 3.0 3.0))"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestSpatialQueryWithMultiPolygon(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, MULTIPOLYGON(((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0)),((2.0 2.0, 3.0 2.0, 3.0 3.0, 2.0 3.0, 2.0 2.0))))"
	expectedSQL := "ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 4326))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "MULTIPOLYGON(((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0)), ((2.0 2.0, 3.0 2.0, 3.0 3.0, 2.0 3.0, 2.0 2.0)))"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestSpatialQueryWithGeometryCollection(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, GEOMETRYCOLLECTION(POINT(0.0 0.0),LINESTRING(0.0 0.0, 1.0 1.0)))"
	expectedSQL := "ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 4326))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "GEOMETRYCOLLECTION(POINT(0.0 0.0), LINESTRING(0.0 0.0, 1.0 1.0))"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestSpatialQueryWithBbox(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, BBOX(10.0, 20.1, 30.0, 40.0))"
	expectedSQL := "ST_Intersects(CastAutomagic(\"geom\"), BuildMbr(:cql_bcde, :cql_fghi, :cql_jklm, :cql_nopq, 4326))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": 10.0, "cql_fghi": 20.1, "cql_jklm": 30.0, "cql_nopq": 40.0}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestSpatialQueryFailsOnInvalidBbox(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, BBOX(10.0 20.1 30.0 40.0))"

	// when
	_, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	assert.ErrorContains(t, err, "missing east bound coordinate (maxx) in bounding box")
}

func TestSpatialQueryFailsOnInvalidBboxWithText(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, BBOX(10.0, 'bla'))"

	// when
	_, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	assert.ErrorContains(t, err, "'bla' is not a valid numeric type")
}

func TestSpatialQueryFailsOnNonGeometryProperty(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geom, BBOX(10.0, 20.1, 30.0, 40.0))" // should be geometry instead of geom

	// when
	_, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	assert.ErrorContains(t, err, "spatial filtering is only supported on property 'geometry'")
}

func TestSpatialQueryFailsOnUndefinedGeometry(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geometry", IsPrimaryGeometry: false}} // not marked as primary geometry
	inputCQL := "S_INTERSECTS(geometry, BBOX(10.0, 20.1, 30.0, 40.0))"

	// when
	_, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	assert.ErrorContains(t, err, "spatial filtering is not supported for this collection since there is no geometry field defined")
}

func TestSpatialQueryUsesRtree(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, BBOX(10.0, 20.1, 30.0, 40.0))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	normalizedRtreeClause := strings.Join(strings.Fields(strings.Map(removeNewlinesAndTabs, actual.RtreeSQL)), " ")
	assert.Equal(t, "AND EXISTS (SELECT 1 FROM rtree_%[1]s_%[2]s r WHERE r.minx <= MbrMaxX(BuildMbr(:cql_bcde, :cql_fghi, :cql_jklm, :cql_nopq, 4326)) AND r.maxx >= MbrMinX(BuildMbr(:cql_bcde, :cql_fghi, :cql_jklm, :cql_nopq, 4326)) AND r.miny <= MbrMaxY(BuildMbr(:cql_bcde, :cql_fghi, :cql_jklm, :cql_nopq, 4326)) AND r.maxy >= MbrMinY(BuildMbr(:cql_bcde, :cql_fghi, :cql_jklm, :cql_nopq, 4326)))", normalizedRtreeClause)
}

func TestSpatialQueryWithGeometryAndBooleanFilter(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "prop1 = 'foo' AND S_INTERSECTS(geometry, POINT(4.897 52.377))"
	expectedSQL := "(\"prop1\" = :cql_bcde AND ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_fghi, 4326)))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "'foo'", "cql_fghi": "POINT(4.897 52.377)"}, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

func TestSpatialQueryWithAllSpatialFunctions(t *testing.T) {
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}

	for cqlFunc, sqlFunc := range spatialFunctions {
		t.Run(cqlFunc, func(t *testing.T) {
			inputCQL := cqlFunc + "(geometry, POINT(4.897 52.377))"
			expectedSQL := sqlFunc + "(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 4326))"

			actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

			require.NoError(t, err)
			assert.Equal(t, map[string]any{"cql_bcde": "POINT(4.897 52.377)"}, actual.Params)
			assert.Equal(t, expectedSQL, actual.SQL)
		})
	}
}

func TestSpatialQueryForAllWellKnownTexts(t *testing.T) {
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	wkts := []string{
		"POINT(10 20)",
		"POINT (10 20)",
		"POINT Z(10 20 5)",
		"POINT Z (10 20 5)",
		"LINESTRING(0 0, 10 10)",
		"LINESTRING (0 0, 10 10)",
		"LINESTRING Z(0 0 0, 10 10 5)",
		"LINESTRING Z (0 0 0, 10 10 5)",
		"POLYGON((0 0, 10 0, 10 10, 0 10, 0 0))",
		"POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
		"POLYGON((0 0, 10 0, 10 10, 0 10, 0 0), (3 3, 7 3, 7 7, 3 7, 3 3))",
		"POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0), (3 3, 7 3, 7 7, 3 7, 3 3))",
		"POLYGON Z((0 0 0, 10 0 1, 10 10 2, 0 10 1, 0 0 0))",
		"POLYGON Z ((0 0 0, 10 0 1, 10 10 2, 0 10 1, 0 0 0))",
		"MULTIPOINT(10 20, 30 40)",
		"MULTIPOINT (10 20, 30 40)",
		"MULTIPOINT((10 20), (30 40))",
		"MULTIPOINT ((10 20), (30 40))",
		"MULTIPOINT Z (10 20 5, 30 40 7)",
		"MULTIPOINT Z(10 20 5, 30 40 7)",
		"MULTIPOINT Z ((10 20 5), (30 40 7))",
		"MULTIPOINT Z((10 20 5), (30 40 7))",
		"MULTILINESTRING((0 0, 10 10), (20 20, 30 10))",
		"MULTILINESTRING ((0 0, 10 10), (20 20, 30 10))",
		"MULTILINESTRING Z((0 0 0, 10 10 5), (20 20 1, 30 10 2))",
		"MULTILINESTRING Z ((0 0 0, 10 10 5), (20 20 1, 30 10 2))",
		"MULTIPOLYGON(((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 20, 30 20, 30 30, 20 30, 20 20)))",
		"MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 20, 30 20, 30 30, 20 30, 20 20)))",
		"MULTIPOLYGON(((0 0, 10 0, 10 10, 0 10, 0 0), (3 3, 7 3, 7 7, 3 7, 3 3)))",
		"MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0), (3 3, 7 3, 7 7, 3 7, 3 3)))",
		"MULTIPOLYGON Z(((0 0 0, 10 0 1, 10 10 2, 0 10 1, 0 0 0)))",
		"MULTIPOLYGON Z (((0 0 0, 10 0 1, 10 10 2, 0 10 1, 0 0 0)))",
		"GEOMETRYCOLLECTION(POINT(10 20), LINESTRING(0 0, 10 10), POLYGON((0 0, 10 0, 10 10, 0 10, 0 0)))",
		"GEOMETRYCOLLECTION (POINT (10 20), LINESTRING (0 0, 10 10), POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)))",
		"GEOMETRYCOLLECTION Z(POINT(10 20 5), LINESTRING(0 0 0, 10 10 5))",
		"GEOMETRYCOLLECTION Z (POINT (10 20 5), LINESTRING (0 0 0, 10 10 5))",
	}

	for _, wkt := range wkts {
		t.Run(wkt, func(t *testing.T) {
			inputCQL := "S_INTERSECTS(geometry, " + wkt + ")"
			expectedSQL := "ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 4326))"

			actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

			require.NoError(t, err)
			assert.Equal(t, map[string]any{"cql_bcde": wkt}, actual.Params)
			assert.Equal(t, expectedSQL, actual.SQL)
		})
	}
}

// Test CQL examples provided by OGC.
// See https://github.com/opengeospatial/ogcapi-features/tree/64ac2d892b877b711a4570336cb9d42e2afb4ef8/cql2/standard/schema/examples/text
func TestCQLExamplesProvidedByOGC(t *testing.T) {
	ogcExamples := path.Join(pwd, "testdata", "ogc")

	entries, err := os.ReadDir(ogcExamples)
	require.NoError(t, err)

	for _, entry := range entries {
		t.Skip("DISABLED FOR NOW, enable once implementation is further completed") // TODO: enable.

		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			example, err := os.ReadFile(path.Join(ogcExamples, entry.Name()))
			require.NoError(t, err)

			inputCQL := strings.Map(removeNewlinesAndTabs, strings.TrimSpace(string(example)))
			require.NotEmpty(t, inputCQL)
			log.Printf("Parsing CQL: %s", inputCQL)

			queryables := []domain.Field{{Name: "*"}} // allow all
			actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

			require.NoError(t, err)
			require.NotNil(t, actual)
			assert.NotEmpty(t, actual.SQL)
			assertValidSQLiteQuery(t, actual)
		})
	}
}

func removeNewlinesAndTabs(r rune) rune {
	if r == '\n' || r == '\r' || r == '\t' {
		return -1
	}
	return r
}

func assertValidSQLiteQuery(t *testing.T, result *SQLResult) {
	t.Helper()

	loadExtensions()

	require.NotNil(t, result)

	dbPath := pwd + "/testdata/cql.gpkg"
	db, err := sqlx.Open("sqlite_spatialite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	query := "select * from cql where " + result.SQL
	rows, err := db.NamedQuery(query, result.Params)
	require.NoError(t, err)

	defer rows.Close()
	for rows.Next() {
		_ = rows.Scan()
	}
	require.NoError(t, rows.Err())
}
