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
	_, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, []string{}))

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
	_, _, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

	// then
	assert.ErrorContains(t, err, "property 'prop2' cannot be used in CQL filter, is not a queryable property")
}

func TestBooleanQueryWithNumbers(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "prop1 = 10 AND prop2 < 5"
	expectedSQL := "(\"prop1\" = :cql_bcde AND \"prop2\" < :cql_fghi)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	_, _, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
	_, _, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

	// then
	assert.ErrorContains(t, err, "syntax error at column 23: extraneous input ''baz'' expecting {')', ','}")
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
			actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

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
