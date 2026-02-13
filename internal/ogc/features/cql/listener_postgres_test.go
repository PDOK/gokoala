package cql

import (
	"testing"

	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoop(t *testing.T) {
	// given
	inputCQL := "prop1 = 10 AND prop2 < 5"
	expectedSQL := ""

	// when
	actualSQL, params, err := ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, []string{}))

	// then
	require.NoError(t, err)
	assert.Empty(t, params)
	assert.Equal(t, expectedSQL, actualSQL)
}
