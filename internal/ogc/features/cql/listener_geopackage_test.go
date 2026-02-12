package cql

import (
	"testing"

	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	inputCQL := "prop1 = 10 AND prop2 < 5"

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
	assert.Equal(t, map[string]any{"cql_bcde": "'foo'", "cql_fghi": "'bar'", "cql_jklm": "'abc'"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}
