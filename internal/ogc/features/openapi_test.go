package features

import (
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/stretchr/testify/require"

	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/geopackage"
	"github.com/stretchr/testify/assert"
)

func TestCreatePropertyFiltersByCollection(t *testing.T) {
	eng, err := config.NewConfig("internal/ogc/features/testdata/geopackage/config_features_bag.yaml")
	require.NoError(t, err)
	oaf := eng.OgcAPI.Features

	eng2, err := config.NewConfig("internal/ogc/features/testdata/geopackage/config_features_bag_invalid_filters.yaml")
	require.NoError(t, err)
	oafWithInvalidPropertyFilter := eng2.OgcAPI.Features

	gpkg, err := geopackage.NewGeoPackage(oaf.Collections, *oaf.Datasources.DefaultWGS84.GeoPackage, false, 0, false)
	require.NoError(t, err)

	tests := []struct {
		name        string
		config      *config.OgcAPIFeatures
		datasources map[DatasourceKey]ds.Datasource
		pf          map[string]ds.QueryablesWithAllowedValues
		wantResult  map[string][]OpenAPIPropertyFilter
		wantErr     bool
	}{
		{
			name:        "Empty input",
			config:      &config.OgcAPIFeatures{},
			datasources: nil,
			pf:          map[string]ds.QueryablesWithAllowedValues{"boo": map[string]ds.QueryableWithAllowedValues{}},
			wantResult:  map[string][]OpenAPIPropertyFilter{},
			wantErr:     false,
		},
		{
			name:   "Valid property filters",
			config: oaf,
			datasources: map[DatasourceKey]ds.Datasource{
				{collectionID: "foo"}: gpkg,
			},
			// keep this in line with the filters in "internal/ogc/features/testdata/geopackage/config_features_bag.yaml"
			pf: map[string]ds.QueryablesWithAllowedValues{
				"foo": map[string]ds.QueryableWithAllowedValues{
					"straatnaam": {
						Queryable:     config.Queryable{Name: "straatnaam", Description: "Filter features by this property"},
						AllowedValues: nil,
					},
					"postcode": {
						Queryable:     config.Queryable{Name: "postcode", Description: "Filter features by this property"},
						AllowedValues: []string{"1234AB", "5678XY"},
					},
				},
			},
			wantResult: map[string][]OpenAPIPropertyFilter{"foo": {
				{Name: "postcode", Description: "Filter features by this property", DataType: "string", AllowedValues: []string{"1234AB", "5678XY"}},
				{Name: "straatnaam", Description: "Filter features by this property", DataType: "string"},
			}},
			wantErr: false,
		},
		{
			name:   "Invalid property filter defined in config",
			config: oafWithInvalidPropertyFilter,
			datasources: map[DatasourceKey]ds.Datasource{
				{collectionID: "foo"}: gpkg,
			},
			// keep this in line with the filters in "internal/ogc/features/testdata/geopackage/config_features_bag_invalid_filters.yaml"
			pf: map[string]ds.QueryablesWithAllowedValues{
				"foo": map[string]ds.QueryableWithAllowedValues{
					"straatnaam": {
						Queryable:     config.Queryable{Name: "straatnaam", Description: "Filter features by this property"},
						AllowedValues: nil,
					},
					"invalid_this_does_not_exist_in_gpkg": {
						Queryable:     config.Queryable{Name: "invalid_this_does_not_exist_in_gpkg", Description: "Filter features by this property"},
						AllowedValues: nil,
					},
					"postcode": {
						Queryable:     config.Queryable{Name: "postcode", Description: "Filter features by this property"},
						AllowedValues: []string{"1234AB", "5678XY"},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := createPropertyFiltersByCollection(tt.datasources, tt.pf)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPropertyFiltersByCollection() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}
