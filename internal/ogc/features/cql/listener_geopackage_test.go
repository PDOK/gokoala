package cql

import (
	"database/sql"
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

func TestInvalidBooleanExpression(t *testing.T) {
	// given
	inputCQL := "prop1 ==== 1 AND prop2 !!= 5"

	// when
	_, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, []string{}))

	// then
	require.ErrorContains(t, err, "syntax error at column 7: mismatched input '=' expecting ")
	require.ErrorContains(t, err, "syntax error at column 23: no viable alternative at input 'prop2!'")
	assert.Empty(t, params)
}

func TestFailOnNonQueryablePropertyExpression(t *testing.T) {
	// given
	queryables := []string{"prop1"}
	inputCQL := "prop1 = 30 AND prop2 > 77"

	// when
	_, _, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

	// then
	assert.ErrorContains(t, err, "property 'prop2' cannot be used in CQL filter, is not a queryable property")
}

func TestBooleanExpressionWithNumbers(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "prop1 = 10 AND prop2 < 5"
	expectedSQL := "(\"prop1\" = :cql_bcde AND \"prop2\" < :cql_fghi)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": "10", "cql_fghi": "5"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestMultipleBooleanExpressions(t *testing.T) {
	// given
	queryables := []string{"prop1", "prop2"}
	inputCQL := "(prop1 = 10 OR prop1 = 20) AND NOT (prop2 = 'X')"
	expectedSQL := "((\"prop1\" = :cql_bcde OR \"prop1\" = :cql_fghi) AND NOT (\"prop2\" = :cql_jklm))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

	// then
	require.NoError(t, err)
	assertValidSQLiteQuery(t, actualSQL, params)
	assert.Equal(t, map[string]any{"cql_bcde": "10", "cql_fghi": "20", "cql_jklm": "'X'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestMultipleBooleanExpressionsWithStrings(t *testing.T) {
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

// TODO: enable once implementation is further completed!!
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

			inputCQL := strings.TrimSpace(string(example))
			require.NotEmpty(t, inputCQL)

			queryables := []string{"*"} // allow all
			actualSQL, params, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables))

			require.NoError(t, err)
			assert.NotEmpty(t, actualSQL)
			assertValidSQLiteQuery(t, actualSQL, params)
		})
	}
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
}
