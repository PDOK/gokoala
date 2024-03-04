package geopackage

import (
	"context"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/PDOK/gokoala/config"

	"github.com/PDOK/gokoala/ogc/features/datasources"
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
	loadDriver()
	return newLocalGeoPackage(&config.GeoPackageLocal{
		GeoPackageCommon: config.GeoPackageCommon{
			Fid:                       "feature_id",
			QueryTimeout:              15 * time.Second,
			MaxBBoxSizeToUseWithRTree: 30000,
		},
		File: pwd + "/testdata/bag.gpkg",
	})
}

func newTemporalAddressesGeoPackage() geoPackageBackend {
	loadDriver()
	return newLocalGeoPackage(&config.GeoPackageLocal{
		GeoPackageCommon: config.GeoPackageCommon{
			Fid:                       "feature_id",
			QueryTimeout:              15 * time.Second,
			MaxBBoxSizeToUseWithRTree: 30000,
		},
		File: pwd + "/testdata/bag-temporal.gpkg",
	})
}

func TestNewGeoPackage(t *testing.T) {
	type args struct {
		config     config.GeoPackage
		collection config.GeoSpatialCollections
	}
	tests := []struct {
		name                        string
		args                        args
		wantNrOfFeatureTablesInGpkg int
	}{
		{
			name: "open local geopackage",
			args: args{
				config: config.GeoPackage{
					Local: &config.GeoPackageLocal{
						GeoPackageCommon: config.GeoPackageCommon{
							Fid: "feature_id",
						},
						File: pwd + "/testdata/bag.gpkg",
					},
				},
				collection: []config.GeoSpatialCollection{
					{
						ID:       "ligplaatsen",
						Features: &config.CollectionEntryFeatures{},
					},
				},
			},
			wantNrOfFeatureTablesInGpkg: 1, // 3 in geopackage, but we define only 1 collection
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantNrOfFeatureTablesInGpkg, len(NewGeoPackage(tt.args.collection, tt.args.config).featureTableByCollectionID), "NewGeoPackage(%v)", tt.args.config)
		})
	}
}

func TestGeoPackage_GetFeatures(t *testing.T) {
	type fields struct {
		backend          geoPackageBackend
		fidColumn        string
		featureTableByID map[string]*featureTable
		queryTimeout     time.Duration
	}
	type args struct {
		ctx         context.Context
		collection  string
		queryParams datasources.FeaturesCriteria
	}
	refDate, _ := time.Parse(time.RFC3339, "2023-12-31T00:00:00Z")
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFC     *domain.FeatureCollection
		wantCursor domain.Cursors
		wantErr    bool
	}{
		{
			name: "get first page of features",
			fields: fields{
				backend:          newAddressesGeoPackage(),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*featureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     60 * time.Second,
			},
			args: args{
				ctx:        context.Background(),
				collection: "ligplaatsen",
				queryParams: datasources.FeaturesCriteria{
					Cursor: domain.DecodedCursor{FID: 0, FiltersChecksum: []byte{}},
					Limit:  2,
				},
			},
			wantFC: &domain.FeatureCollection{
				NumberReturned: 2,
				Features: []*domain.Feature{
					{
						Feature: geojson.Feature{
							Properties: map[string]any{
								"straatnaam": "Van Diemenkade",
								"nummer_id":  "0363200000454013",
							},
						},
					},
					{
						Feature: geojson.Feature{
							Properties: map[string]any{
								"straatnaam": "Realengracht",
								"nummer_id":  "0363200000398886",
							},
						},
					},
				},
			},
			wantCursor: domain.Cursors{
				Prev: "|",
				Next: "Dv4|", // 3838
			},
			wantErr: false,
		},
		{
			name: "get second page of features",
			fields: fields{
				backend:          newAddressesGeoPackage(),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*featureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        context.Background(),
				collection: "ligplaatsen",
				queryParams: datasources.FeaturesCriteria{
					Cursor: domain.DecodedCursor{
						FID:             3838, // see next cursor from test above
						FiltersChecksum: []byte{},
					},
					Limit: 3,
				},
			},
			wantFC: &domain.FeatureCollection{
				NumberReturned: 3,
				Features: []*domain.Feature{
					{
						Feature: geojson.Feature{
							Properties: map[string]any{
								"straatnaam": "Realengracht",
								"nummer_id":  "0363200000398887",
							},
						},
					},
					{
						Feature: geojson.Feature{
							Properties: map[string]any{
								"straatnaam": "Realengracht",
								"nummer_id":  "0363200000398888",
							},
						},
					},
					{
						Feature: geojson.Feature{
							Properties: map[string]any{
								"straatnaam": "Realengracht",
								"nummer_id":  "0363200000398889",
							},
						},
					},
				},
			},
			wantCursor: domain.Cursors{
				Prev: "|",
				Next: "DwE|",
			},
			wantErr: false,
		},
		{
			name: "get first page of features with reference date",
			fields: fields{
				backend:          newTemporalAddressesGeoPackage(),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*featureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     60 * time.Second,
			},
			args: args{
				ctx:        context.Background(),
				collection: "ligplaatsen",
				queryParams: datasources.FeaturesCriteria{
					Cursor: domain.DecodedCursor{FID: 0, FiltersChecksum: []byte{}},
					Limit:  2,
					TemporalCriteria: datasources.TemporalCriteria{
						ReferenceDate:     refDate,
						StartDateProperty: "datum_strt",
						EndDateProperty:   "datum_eind",
					},
				},
			},
			wantFC: &domain.FeatureCollection{
				NumberReturned: 2,
				Features: []*domain.Feature{
					{
						Feature: geojson.Feature{
							Properties: map[string]any{
								"straatnaam": "Van Diemenkade",
								"nummer_id":  "0363200000454013",
							},
						},
					},
					{
						Feature: geojson.Feature{
							Properties: map[string]any{
								"straatnaam": "Realengracht",
								"nummer_id":  "0363200000398886",
							},
						},
					},
				},
			},
			wantCursor: domain.Cursors{
				Prev: "|",
				Next: "Dv4|", // 3838
			},
			wantErr: false,
		},
		{
			name: "fail on non existing collection",
			fields: fields{
				backend:          newAddressesGeoPackage(),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*featureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        context.Background(),
				collection: "vakantiehuizen", // not in gpkg
				queryParams: datasources.FeaturesCriteria{
					Cursor: domain.DecodedCursor{FID: 0, FiltersChecksum: []byte{}},
					Limit:  10,
				},
			},
			wantFC:     nil,
			wantCursor: domain.Cursors{},
			wantErr:    true, // should fail
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GeoPackage{
				backend:                    tt.fields.backend,
				fidColumn:                  tt.fields.fidColumn,
				featureTableByCollectionID: tt.fields.featureTableByID,
				queryTimeout:               tt.fields.queryTimeout,
			}
			g.preparedStmtCache = NewCache()

			fc, cursor, err := g.GetFeatures(tt.args.ctx, tt.args.collection, tt.args.queryParams)
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
		featureTableByID map[string]*featureTable
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
				featureTableByID: map[string]*featureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
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
					Properties: map[string]any{
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
				featureTableByID: map[string]*featureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
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
				featureTableByID: map[string]*featureTable{"ligplaatsen": {TableName: "ligplaatsen", GeometryColumnName: "geom"}},
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
				backend:                    tt.fields.backend,
				fidColumn:                  tt.fields.fidColumn,
				featureTableByCollectionID: tt.fields.featureTableByID,
				queryTimeout:               tt.fields.queryTimeout,
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
