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
	_, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, []string{}, 0))

	// then
	require.ErrorContains(t, err, "syntax error at column 7: mismatched input '=' expecting ")
	require.ErrorContains(t, err, "syntax error at column 23: no viable alternative at input 'prop2!'")
	assert.Empty(t, params)
}

func TestFailOnNonQueryablePropertyQuery(t *testing.T) {
	// given
	queryables := []string{"prop1"}
	inputCQL := "prop1 = 30 AND prop2 > 77"

	// when
	_, _, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	assert.ErrorContains(t, err, "property 'prop2' cannot be used in CQL filter, is not a queryable property")
}

func TestBooleanQueryWithNumbers(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "prop1 = 10 AND prop2 < 5"
	expectedSQL := "(\"prop1\" = :cql_bcde AND \"prop2\" < :cql_fghi)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": int64(10), "cql_fghi": int64(5)}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestMultipleBooleanQueries(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "(prop1 = 10 OR prop1 = 20) AND NOT (prop2 = 'X')"
	expectedSQL := "((\"prop1\" = :cql_bcde OR \"prop1\" = :cql_fghi) AND NOT (\"prop2\" = :cql_jklm))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": int64(10), "cql_fghi": int64(20), "cql_jklm": "'X'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestBooleanLiterals(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "(prop1 = true OR prop2 = 20)"
	expectedSQL := "(\"prop1\" = 1 OR \"prop2\" = :cql_bcde)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": int64(20)}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestMultipleBooleanQueriesWithStrings(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2", "prop3"}
	inputCQL := "(prop1 = 'foo' AND prop2 = 'bar') OR prop3 = 'abc'"
	expectedSQL := "((\"prop1\" = :cql_bcde AND \"prop2\" = :cql_fghi) OR \"prop3\" = :cql_jklm)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": "'foo'", "cql_fghi": "'bar'", "cql_jklm": "'abc'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestLikeOperator(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2", "prop3"}
	inputCQL := "prop1 LIKE 'foo%' AND prop2 LIKE 'bar_' OR prop3 LIKE '%abc'"
	expectedSQL := "((\"prop1\" LIKE :cql_bcde AND \"prop2\" LIKE :cql_fghi) OR \"prop3\" LIKE :cql_jklm)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": "'foo%'", "cql_fghi": "'bar_'", "cql_jklm": "'%abc'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestNotLikeOperator(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2", "prop3"}
	inputCQL := "prop1 NOT LIKE 'foo%' AND prop2 LIKE 'bar_' OR prop3 LIKE '%abc'"
	expectedSQL := "((\"prop1\" NOT LIKE :cql_bcde AND \"prop2\" LIKE :cql_fghi) OR \"prop3\" LIKE :cql_jklm)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": "'foo%'", "cql_fghi": "'bar_'", "cql_jklm": "'%abc'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestLikeOperatorFailOnMissingWildcard(t *testing.T) {
	// given
	queryables := []string{"prop1"}
	inputCQL := "prop1 LIKE 'foo'"

	// when
	_, _, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	assert.ErrorContains(t, err, "LIKE pattern is missing wildcard symbol. "+
		"Either percentage '%' to match multiple characters or underscore '_' to match a "+
		"single character can be used as a wildcard symbol. For example: LIKE 'foo%'.")
}

func TestBetweenOperator(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "prop1 BETWEEN 4 AND 6 AND prop2 = 'bar'"
	expectedSQL := "(\"prop1\" BETWEEN :cql_bcde AND :cql_fghi AND \"prop2\" = :cql_jklm)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": int64(4), "cql_fghi": int64(6), "cql_jklm": "'bar'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestNotBetweenOperator(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "prop1 NOT BETWEEN 4 AND 6 AND prop2 = 'bar'"
	expectedSQL := "(\"prop1\" NOT BETWEEN :cql_bcde AND :cql_fghi AND \"prop2\" = :cql_jklm)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": int64(4), "cql_fghi": int64(6), "cql_jklm": "'bar'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestInListOperator(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "prop1 IN ('foo', 'bar', 'baz') AND prop2 = 'baz'"
	expectedSQL := "(\"prop1\" IN (:cql_bcde, :cql_fghi, :cql_jklm) AND \"prop2\" = :cql_nopq)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": "'foo'", "cql_fghi": "'bar'", "cql_jklm": "'baz'", "cql_nopq": "'baz'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestNotInListOperator(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "prop1 NOT IN ('foo', 'bar', 'baz') AND prop2 = 'baz'"
	expectedSQL := "(\"prop1\" NOT IN (:cql_bcde, :cql_fghi, :cql_jklm) AND \"prop2\" = :cql_nopq)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": "'foo'", "cql_fghi": "'bar'", "cql_jklm": "'baz'", "cql_nopq": "'baz'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestIsNullOperator(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "prop1 IS NULL AND prop2 = 'baz'"
	expectedSQL := "(\"prop1\" IS NULL AND \"prop2\" = :cql_bcde)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": "'baz'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestIsNotNullOperator(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "prop1 IS NOT NULL AND prop2 = 'baz'"
	expectedSQL := "(\"prop1\" IS NOT NULL AND \"prop2\" = :cql_bcde)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": "'baz'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestFailOnInvalidInListQuery(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "prop1 IN ('foo', 'bar' 'baz')"

	// when
	_, _, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

	// then
	assert.ErrorContains(t, err, "syntax error at column 23: extraneous input ''baz'' expecting {')', ','}")
}

func TestSpatialQueryWithPoint(t *testing.T) {
	// given
	queryables := []string{"geom"}
	inputCQL := "S_INTERSECTS(geom, POINT(4.897 52.377))"
	expectedSQL := "ST_Intersects(\"geom\", ST_GeomFromText('POINT(4.897 52.377)', 4326))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Empty(t, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestSpatialQueryWithPoint3D(t *testing.T) {
	// given
	queryables := []string{"geom"}
	inputCQL := "S_INTERSECTS(geom, POINT(4.897 52.377 10.0))"
	expectedSQL := "ST_Intersects(\"geom\", ST_GeomFromText('POINT(4.897 52.377 10.0)', 4326))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Empty(t, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestSpatialQueryWithLinestring(t *testing.T) {
	// given
	queryables := []string{"geom"}
	inputCQL := "S_INTERSECTS(geom, LINESTRING(0.0 0.0, 1.0 1.0, 2.0 0.0))"
	expectedSQL := "ST_Intersects(\"geom\", ST_GeomFromText('LINESTRING(0.0 0.0, 1.0 1.0, 2.0 0.0)', 4326))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Empty(t, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestSpatialQueryWithPolygon(t *testing.T) {
	// given
	queryables := []string{"geom"}
	inputCQL := "S_INTERSECTS(geom, POLYGON((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0)))"
	expectedSQL := "ST_Intersects(\"geom\", ST_GeomFromText('POLYGON((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0))', 28992))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 28992))

	// then
	require.NoError(t, err)
	assert.Empty(t, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestSpatialQueryWithPolygonWithHole(t *testing.T) {
	// given
	queryables := []string{"geom"}
	inputCQL := "S_INTERSECTS(geom, POLYGON((0.0 0.0, 10.0 0.0, 10.0 10.0, 0.0 10.0, 0.0 0.0),(2.0 2.0, 8.0 2.0, 8.0 8.0, 2.0 8.0, 2.0 2.0)))"
	expectedSQL := "ST_Intersects(\"geom\", ST_GeomFromText('POLYGON((0.0 0.0, 10.0 0.0, 10.0 10.0, 0.0 10.0, 0.0 0.0), (2.0 2.0, 8.0 2.0, 8.0 8.0, 2.0 8.0, 2.0 2.0))', 4326))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Empty(t, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestSpatialQueryWithMultiPoint(t *testing.T) {
	// given
	queryables := []string{"geom"}
	inputCQL := "S_INTERSECTS(geom, MULTIPOINT(0.0 0.0, 1.0 1.0))"
	expectedSQL := "ST_Intersects(\"geom\", ST_GeomFromText('MULTIPOINT(0.0 0.0, 1.0 1.0)', 4326))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Empty(t, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestSpatialQueryWithMultiLinestring(t *testing.T) {
	// given
	queryables := []string{"geom"}
	inputCQL := "S_INTERSECTS(geom, MULTILINESTRING((0.0 0.0, 1.0 1.0),(2.0 2.0, 3.0 3.0)))"
	expectedSQL := "ST_Intersects(\"geom\", ST_GeomFromText('MULTILINESTRING((0.0 0.0, 1.0 1.0), (2.0 2.0, 3.0 3.0))', 4326))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Empty(t, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestSpatialQueryWithMultiPolygon(t *testing.T) {
	// given
	queryables := []string{"geom"}
	inputCQL := "S_INTERSECTS(geom, MULTIPOLYGON(((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0)),((2.0 2.0, 3.0 2.0, 3.0 3.0, 2.0 3.0, 2.0 2.0))))"
	expectedSQL := "ST_Intersects(\"geom\", ST_GeomFromText('MULTIPOLYGON(((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0)), ((2.0 2.0, 3.0 2.0, 3.0 3.0, 2.0 3.0, 2.0 2.0)))', 4326))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Empty(t, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestSpatialQueryWithGeometryCollection(t *testing.T) {
	// given
	queryables := []string{"geom"}
	inputCQL := "S_INTERSECTS(geom, GEOMETRYCOLLECTION(POINT(0.0 0.0),LINESTRING(0.0 0.0, 1.0 1.0)))"
	expectedSQL := "ST_Intersects(\"geom\", ST_GeomFromText('GEOMETRYCOLLECTION(POINT(0.0 0.0), LINESTRING(0.0 0.0, 1.0 1.0))', 4326))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Empty(t, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestSpatialQueryWithBbox(t *testing.T) {
	// given
	queryables := []string{"geom"}
	inputCQL := "S_INTERSECTS(geom, BBOX(10.0, 20.0, 30.0, 40.0))"
	expectedSQL := "ST_Intersects(\"geom\", BuildMbr(10.0, 20.0, 30.0, 40.0, 4326))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Empty(t, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestSpatialQueryWithGeometryAndBooleanFilter(t *testing.T) {
	// given
	queryables := []string{"geom", "prop1"}
	inputCQL := "prop1 = 'foo' AND S_INTERSECTS(geom, POINT(4.897 52.377))"
	expectedSQL := "(\"prop1\" = :cql_bcde AND ST_Intersects(\"geom\", ST_GeomFromText('POINT(4.897 52.377)', 4326)))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "'foo'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestSpatialQueryWithAllSpatialFunctions(t *testing.T) {
	queryables := []string{"geom"}

	for cqlFunc, sqlFunc := range spatialFunctions {
		t.Run(cqlFunc, func(t *testing.T) {
			inputCQL := cqlFunc + "(geom, POINT(4.897 52.377))"
			expectedSQL := sqlFunc + "(\"geom\", ST_GeomFromText('POINT(4.897 52.377)', 4326))"

			actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

			require.NoError(t, err)
			assert.Empty(t, params)
			assert.Equal(t, expectedSQL, actualSQL)
		})
	}
}

func TestSpatialQueryForAllWellKnownTexts(t *testing.T) {
	queryables := []string{"geom"}
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
			inputCQL := "S_INTERSECTS(geom, " + wkt + ")"
			expectedSQL := "ST_Intersects(\"geom\", ST_GeomFromText('" + wkt + "', 4326))"

			actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326))

			require.NoError(t, err)
			assert.Empty(t, params)
			assert.Equal(t, expectedSQL, actualSQL)
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

			queryables := []string{"*"} // allow all
			actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0))

			require.NoError(t, err)
			assert.NotEmpty(t, actualSQL)
			assertValidSQLiteQuery(t, actualSQL, params)
		})
	}
}

func removeNewlinesAndTabs(r rune) rune {
	if r == '\n' || r == '\r' || r == '\t' {
		return -1
	}
	return r
}

func assertValidSQLiteQuery(t *testing.T, filter string, params map[string]any) {
	t.Helper()

	loadExtensions()

	dbPath := pwd + "/testdata/cql.gpkg"
	db, err := sqlx.Open("sqlite_spatialite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	query := "select * from cql where " + filter
	rows, err := db.NamedQuery(query, params)
	require.NoError(t, err)

	defer rows.Close()
	for rows.Next() {
		_ = rows.Scan()
	}
	require.NoError(t, rows.Err())
}
