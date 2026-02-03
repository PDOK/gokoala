package cql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBooleanExpression(t *testing.T) {
	inputCQL := "prop1 = 10 AND prop2 < 5"
	expectedSQL := "(\"prop1\" = 10 AND \"prop2\" < 5)"
	actualSQL := ParseToSQL(inputCQL, NewSqliteListener())
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestMultipleBooleanExpressions(t *testing.T) {
	inputCQL := "(prop1 = 10 OR prop1 = 20) AND NOT (prop2 = 'X')"
	expectedSQL := "((\"prop1\" = 10 OR \"prop1\" = 20) AND NOT (\"prop2\" = 'X'))" // TODO: check spec for single quote handling
	actualSQL := ParseToSQL(inputCQL, NewSqliteListener())
	assert.Equal(t, expectedSQL, actualSQL)
}

func TestMultipleBooleanExpressions2(t *testing.T) {
	inputCQL := "(prop1 = foo AND prop2 = bar) OR prop3 = abc"
	expectedSQL := "((\"prop1\" = \"foo\" AND \"prop2\" = \"bar\") OR \"prop3\" = \"abc\")"
	actualSQL := ParseToSQL(inputCQL, NewSqliteListener())
	assert.Equal(t, expectedSQL, actualSQL)
}
