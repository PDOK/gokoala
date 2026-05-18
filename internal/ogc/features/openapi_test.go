package features

import (
	"testing"

	"github.com/stretchr/testify/assert"

	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

func TestCreatePropertyFiltersByCollection(t *testing.T) {
	tests := []struct {
		name                   string
		queryablesByCollection map[string]ds.Queryables
		wantResult             map[string][]OpenAPIPropertyFilter
	}{
		{
			name:                   "Empty input",
			queryablesByCollection: map[string]ds.Queryables{"boo": map[string]ds.QueryableWithAllowedValues{}},
			wantResult:             map[string][]OpenAPIPropertyFilter{},
		},
		{
			name: "Valid property filters",
			queryablesByCollection: map[string]ds.Queryables{
				"foo": map[string]ds.QueryableWithAllowedValues{
					"straatnaam": {
						Field:         domain.Field{Name: "straatnaam", Type: "text", Description: "Filter features by this property"},
						AllowedValues: nil,
					},
					"postcode": {
						Field:         domain.Field{Name: "postcode", Type: "text", Description: "Filter features by this property"},
						AllowedValues: []string{"1234AB", "5678XY"},
					},
				},
			},
			wantResult: map[string][]OpenAPIPropertyFilter{"foo": {
				{Name: "postcode", Description: "Filter features by this property", DataType: "string", AllowedValues: []string{"1234AB", "5678XY"}},
				{Name: "straatnaam", Description: "Filter features by this property", DataType: "string"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := toOpenAPIFilters(tt.queryablesByCollection)
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}
