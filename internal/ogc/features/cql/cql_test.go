package cql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseToSQL(t *testing.T) {
	type args struct {
		cql  string
		tree string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "simple",
			args: args{
				cql:  "amount <= 100",
				tree: "(cqlFilter (booleanExpression (booleanTerm (booleanFactor (booleanPrimary (predicate (comparisonPredicate (binaryComparisonPredicate (scalarExpression (propertyName amount)) <= (scalarExpression (numericLiteral 100))))))))) <EOF>)",
			},
		},
		{
			name: "spatial",
			args: args{
				cql:  "S_INTERSECTS(geometry,POINT(36.319836 32.288087))",
				tree: "(cqlFilter (booleanExpression (booleanTerm (booleanFactor (booleanPrimary (predicate (spatialPredicate S_INTERSECTS ( (geomExpression (propertyName geometry)) , (geomExpression (spatialInstance (geometryLiteral (point POINT ( (coordinate (xCoord 36.319836) (yCoord 32.288087)) ))))) ))))))) <EOF>)",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ParseToSQL(tt.args.cql)
			assert.Equal(t, tt.args.tree, actual)
		})
	}
}
