package features

import (
	"testing"

	"github.com/PDOK/gokoala/config"

	ds "github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/geopackage"
	"github.com/stretchr/testify/assert"
)

func TestCreatePropertyFiltersByCollection(t *testing.T) {
	eng, err := config.NewConfig("internal/ogc/features/testdata/config_features_bag.yaml")
	assert.NoError(t, err)
	oaf := eng.OgcAPI.Features

	eng2, err := config.NewConfig("internal/ogc/features/testdata/config_features_bag_invalid_filters.yaml")
	assert.NoError(t, err)
	oafWithInvalidPropertyFilter := eng2.OgcAPI.Features

	tests := []struct {
		name        string
		config      *config.OgcAPIFeatures
		datasources map[DatasourceKey]ds.Datasource
		wantResult  map[string][]OpenAPIPropertyFilter
		wantErr     bool
	}{
		{
			name:        "Empty input",
			config:      &config.OgcAPIFeatures{},
			datasources: nil,
			wantResult:  map[string][]OpenAPIPropertyFilter{},
			wantErr:     false,
		},
		{
			name:   "Valid property filters",
			config: oaf,
			datasources: map[DatasourceKey]ds.Datasource{
				DatasourceKey{collectionID: "foo"}: geopackage.NewGeoPackage(oaf.Collections, *oaf.Datasources.DefaultWGS84.GeoPackage),
			},
			wantResult: map[string][]OpenAPIPropertyFilter{"foo": {
				{Name: "straatnaam", Description: "Filter features by this property", DataType: "string"},
				{Name: "postcode", Description: "Filter features by this property", DataType: "string"},
			}},
			wantErr: false,
		},
		{
			name:   "Invalid property filter defined in config",
			config: oafWithInvalidPropertyFilter,
			datasources: map[DatasourceKey]ds.Datasource{
				DatasourceKey{collectionID: "foo"}: geopackage.NewGeoPackage(oaf.Collections, *oaf.Datasources.DefaultWGS84.GeoPackage),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := createPropertyFiltersByCollection(tt.config, tt.datasources)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPropertyFiltersByCollection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}