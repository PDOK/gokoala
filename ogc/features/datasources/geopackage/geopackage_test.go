package geopackage

import (
	"context"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/features/domain"
	"github.com/go-spatial/geom/encoding/geojson"
	"github.com/stretchr/testify/assert"
)

var pwd string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	pwd = path.Dir(filename)
}

func newAddressesGeoPackage() geoPackageBackend {
	return newLocalGeoPackage(&engine.GeoPackageLocal{
		GeoPackageCommon: engine.GeoPackageCommon{
			Fid: "feature_id",
		},
		File: pwd + "/testdata/addresses.gpkg",
	})
}

func TestNewGeoPackage(t *testing.T) {
	type args struct {
		config engine.GeoPackage
	}
	tests := []struct {
		name                        string
		args                        args
		wantNrOfFeatureTablesInGpkg int
	}{
		{
			name: "open local geopackage",
			args: args{
				engine.GeoPackage{
					Local: &engine.GeoPackageLocal{
						File: pwd + "/testdata/addresses.gpkg",
					},
				},
			},
			wantNrOfFeatureTablesInGpkg: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantNrOfFeatureTablesInGpkg, len(NewGeoPackage(tt.args.config).featureTableByID), "NewGeoPackage(%v)", tt.args.config)
		})
	}
}

func TestGeoPackage_GetFeatures(t *testing.T) {
	type fields struct {
		backend          geoPackageBackend
		fidColumn        string
		featureTableByID map[string]*gpkgFeatureTable
		queryTimeout     time.Duration
	}
	type args struct {
		ctx        context.Context
		collection string
		cursor     int64
		limit      int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFC     *domain.FeatureCollection
		wantCursor domain.Cursor
		wantErr    bool
	}{
		{
			name: "get first page of features",
			fields: fields{
				backend:          newAddressesGeoPackage(),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*gpkgFeatureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        context.Background(),
				collection: "ligplaatsen",
				cursor:     0,
				limit:      2,
			},
			wantFC: &domain.FeatureCollection{
				NumberReturned: 2,
				Features: []*domain.Feature{
					{
						Feature: geojson.Feature{
							Properties: map[string]interface{}{
								"straatnaam": "Van Diemenkade",
								"nummer_id":  "0363200000454013",
							},
						},
					},
					{
						Feature: geojson.Feature{
							Properties: map[string]interface{}{
								"straatnaam": "Realengracht",
								"nummer_id":  "0363200000398886",
							},
						},
					},
				},
			},
			wantCursor: domain.Cursor{
				Prev: "spDyEwb4",
				Next: "trrEb5db", // 3837
			},
			wantErr: false,
		},
		{
			name: "get second page of features",
			fields: fields{
				backend:          newAddressesGeoPackage(),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*gpkgFeatureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        context.Background(),
				collection: "ligplaatsen",
				cursor:     3837, // see next cursor from test above
				limit:      3,
			},
			wantFC: &domain.FeatureCollection{
				NumberReturned: 3,
				Features: []*domain.Feature{
					{
						Feature: geojson.Feature{
							Properties: map[string]interface{}{
								"straatnaam": "Realengracht",
								"nummer_id":  "0363200000398887",
							},
						},
					},
					{
						Feature: geojson.Feature{
							Properties: map[string]interface{}{
								"straatnaam": "Realengracht",
								"nummer_id":  "0363200000398888",
							},
						},
					},
					{
						Feature: geojson.Feature{
							Properties: map[string]interface{}{
								"straatnaam": "Realengracht",
								"nummer_id":  "0363200000398889",
							},
						},
					},
				},
			},
			wantCursor: domain.Cursor{
				Prev: "LZZS4c3w",
				Next: "CNNniQpu",
			},
			wantErr: false,
		},
		{
			name: "fail on non existing collection",
			fields: fields{
				backend:          newAddressesGeoPackage(),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*gpkgFeatureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        context.Background(),
				collection: "vakantiehuizen", // not in gpkg
				cursor:     0,
				limit:      10,
			},
			wantFC:     nil,
			wantCursor: domain.Cursor{},
			wantErr:    true, // should fail
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GeoPackage{
				backend:          tt.fields.backend,
				fidColumn:        tt.fields.fidColumn,
				featureTableByID: tt.fields.featureTableByID,
				queryTimeout:     tt.fields.queryTimeout,
			}
			fc, cursor, err := g.GetFeatures(tt.args.ctx, tt.args.collection, tt.args.cursor, tt.args.limit)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("GetFeatures, error %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.Equal(t, tt.wantFC.NumberReturned, fc.NumberReturned)
			assert.Equal(t, len(tt.wantFC.Features), fc.NumberReturned)
			for i, wantedFeature := range tt.wantFC.Features {
				assert.Equal(t, wantedFeature.Properties["straatnaam"], fc.Features[i].Properties["straatnaam"])
				assert.Equal(t, wantedFeature.Properties["nummer_id"], fc.Features[i].Properties["nummer_id"])
			}
			assert.Equal(t, tt.wantCursor.Prev, cursor.Prev)
			assert.Equal(t, tt.wantCursor.Next, cursor.Next)
		})
	}
}

func TestGeoPackage_GetFeature(t *testing.T) {
	type fields struct {
		backend          geoPackageBackend
		fidColumn        string
		featureTableByID map[string]*gpkgFeatureTable
		queryTimeout     time.Duration
	}
	type args struct {
		ctx        context.Context
		collection string
		featureID  int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *domain.Feature
		wantErr bool
	}{
		{
			name: "get feature",
			fields: fields{
				backend:          newAddressesGeoPackage(),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*gpkgFeatureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        context.Background(),
				collection: "ligplaatsen",
				featureID:  3837,
			},
			want: &domain.Feature{
				ID:    0,
				Links: nil,
				Feature: geojson.Feature{
					Properties: map[string]interface{}{
						"straatnaam": "Realengracht",
						"nummer_id":  "0363200000398886",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "get non existing feature",
			fields: fields{
				backend:          newAddressesGeoPackage(),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*gpkgFeatureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        context.Background(),
				collection: "ligplaatsen",
				featureID:  999991111111111111,
			},
			want:    nil,
			wantErr: false, // not an error situation
		},
		{
			name: "fail on non existing collection",
			fields: fields{
				backend:          newAddressesGeoPackage(),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*gpkgFeatureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        context.Background(),
				collection: "vakantieparken", // not in gpkg
				featureID:  3837,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GeoPackage{
				backend:          tt.fields.backend,
				fidColumn:        tt.fields.fidColumn,
				featureTableByID: tt.fields.featureTableByID,
				queryTimeout:     tt.fields.queryTimeout,
			}
			got, err := g.GetFeature(tt.args.ctx, tt.args.collection, tt.args.featureID)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("GetFeature, error %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if tt.want != nil {
				assert.Equal(t, tt.want.Properties["straatnaam"], got.Properties["straatnaam"])
				assert.Equal(t, tt.want.Properties["nummer_id"], got.Properties["nummer_id"])
			}
		})
	}
}
