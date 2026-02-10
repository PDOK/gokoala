package cql

import (
	"testing"

	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBooleanExpression(t *testing.T) {
	// given
	inputCQL := "prop1 = 10 AND prop2 < 5"
	expectedSQL := "(\"prop1\" = :cql_bcde AND \"prop2\" < :cql_fghi)"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewSqliteListener(&util.MockRandomizer{}))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "10", "cql_fghi": "5"}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestMultipleBooleanExpressions(t *testing.T) {
	// given
	inputCQL := "(prop1 = 10 OR prop1 = 20) AND NOT (prop2 = 'X')"
	expectedSQL := "((\"prop1\" = :cql_bcde OR \"prop1\" = :cql_fghi) AND NOT (\"prop2\" = :cql_jklm))"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewSqliteListener(&util.MockRandomizer{}))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"cql_bcde": "10", "cql_fghi": "20", "cql_jklm": "'X'"}, params) // TODO: check spec for single quote handling
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestMultipleBooleanExpressions2(t *testing.T) {
	// given
	inputCQL := "(prop1 = foo AND prop2 = bar) OR prop3 = abc"
	expectedSQL := "((\"prop1\" = \"foo\" AND \"prop2\" = \"bar\") OR \"prop3\" = \"abc\")"

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewSqliteListener(&util.MockRandomizer{}))

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]any{}, params)
	assert.Equal(t, expectedSQL, actualSQL)
}
