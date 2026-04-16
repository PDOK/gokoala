package cql

import (
	"testing"

	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoop(t *testing.T) {
	// given
	inputCQL := "prop1 = 10 AND prop2 < 5"
	expectedSQL := ""

	// when
	actual, err := ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, []domain.Field{}, 0))

	// then
	require.NoError(t, err)
	require.NotNil(t, actual)
	assert.Empty(t, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}
