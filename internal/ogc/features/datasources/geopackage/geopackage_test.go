package geopackage

import (
	"context"
	neturl "net/url"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/common"
	"github.com/stretchr/testify/require"

	"github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/stretchr/testify/assert"
)

var pwd string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	pwd = path.Dir(filename)
}

func newTestGeoPackage(file string) geoPackageBackend {
	loadDriver()

	return newLocalGeoPackage(&config.GeoPackageLocal{
		GeoPackageCommon: config.GeoPackageCommon{
			DatasourceCommon: config.DatasourceCommon{
				Fid:          "feature_id",
				QueryTimeout: config.Duration{Duration: 15 * time.Second},
			},
			MaxBBoxSizeToUseWithRTree: 30000,
			InMemoryCacheSize:         -2000,
		},
		File: pwd + file,
	})
}

func TestNewGeoPackage(t *testing.T) {
	type args struct {
		config     config.GeoPackage
		collection []config.FeaturesCollection
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
							DatasourceCommon: config.DatasourceCommon{
								Fid: "feature_id",
							},
							InMemoryCacheSize: -2000,
						},
						File: pwd + "/testdata/bag.gpkg",
					},
				},
				collection: []config.FeaturesCollection{
					{
						ID: "ligplaatsen",
					},
				},
			},
			wantNrOfFeatureTablesInGpkg: 1, // 3 in geopackage, but we define only 1 collection
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGeoPackage(tt.args.collection, tt.args.config, false, 0, false)
			require.NoError(t, err)
			assert.Lenf(t, g.TableByCollectionID, tt.wantNrOfFeatureTablesInGpkg, "NewGeoPackage(%v)", tt.args.config)
		})
	}
}

func TestGeoPackage_GetFeatures(t *testing.T) {
	type fields struct {
		backend          geoPackageBackend
		fidColumn        string
		featureTableByID map[string]*common.Table
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
		wantGeom   bool
	}{
		{
			name: "get first page of features",
			fields: fields{
				backend:          newTestGeoPackage("/testdata/bag.gpkg"),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*common.Table{"ligplaatsen": {Name: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     60 * time.Second,
			},
			args: args{
				ctx:        t.Context(),
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
						Properties: domain.NewFeaturePropertiesWithData(false, map[string]any{
							"straatnaam": "Van Diemenkade",
							"nummer_id":  "0363200000454013",
						}),
					},
					{
						Properties: domain.NewFeaturePropertiesWithData(false, map[string]any{
							"straatnaam": "Realengracht",
							"nummer_id":  "0363200000398886",
						}),
					},
				},
			},
			wantCursor: domain.Cursors{
				Prev: "|",
				Next: "Dv4|", // 3838
			},
			wantGeom: true,
			wantErr:  false,
		},
		{
			name: "get second page of features",
			fields: fields{
				backend:          newTestGeoPackage("/testdata/bag.gpkg"),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*common.Table{"ligplaatsen": {Name: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        t.Context(),
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
						Properties: domain.NewFeaturePropertiesWithData(false, map[string]any{
							"straatnaam": "Realengracht",
							"nummer_id":  "0363200000398887",
						}),
					},
					{

						Properties: domain.NewFeaturePropertiesWithData(false, map[string]any{
							"straatnaam": "Realengracht",
							"nummer_id":  "0363200000398888",
						}),
					},
					{
						Properties: domain.NewFeaturePropertiesWithData(false, map[string]any{
							"straatnaam": "Realengracht",
							"nummer_id":  "0363200000398889",
						}),
					},
				},
			},
			wantCursor: domain.Cursors{
				Prev: "|",
				Next: "DwE|",
			},
			wantGeom: true,
			wantErr:  false,
		},
		{
			name: "get first page of features with reference date",
			fields: fields{
				backend:          newTestGeoPackage("/testdata/bag-temporal-wgs84.gpkg"),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*common.Table{"ligplaatsen": {Name: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     60 * time.Second,
			},
			args: args{
				ctx:        t.Context(),
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
						Properties: domain.NewFeaturePropertiesWithData(false, map[string]any{
							"straatnaam": "Van Diemenkade",
							"nummer_id":  "0363200000454013",
						}),
					},
					{
						Properties: domain.NewFeaturePropertiesWithData(false, map[string]any{
							"straatnaam": "Realengracht",
							"nummer_id":  "0363200000398886",
						}),
					},
				},
			},
			wantCursor: domain.Cursors{
				Prev: "|",
				Next: "Dv4|", // 3838
			},
			wantGeom: true,
			wantErr:  false,
		},
		{
			name: "fail on non existing collection",
			fields: fields{
				backend:          newTestGeoPackage("/testdata/bag.gpkg"),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*common.Table{"ligplaatsen": {Name: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        t.Context(),
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
		{
			name: "get features with empty geometry",
			fields: fields{
				backend:          newTestGeoPackage("/testdata/null-empty-geoms.gpkg"),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*common.Table{"ligplaatsen": {Name: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     60 * time.Second,
			},
			args: args{
				ctx:        t.Context(),
				collection: "ligplaatsen",
				queryParams: datasources.FeaturesCriteria{
					Cursor: domain.DecodedCursor{FID: 0, FiltersChecksum: []byte{}},
					Limit:  1,
				},
			},
			wantFC: &domain.FeatureCollection{
				NumberReturned: 1,
				Features: []*domain.Feature{
					{
						Properties: domain.NewFeaturePropertiesWithData(false, map[string]any{
							"straatnaam": "Van Diemenkade",
							"nummer_id":  "0363200000454013",
						}),
					},
				},
			},
			wantCursor: domain.Cursors{
				Prev: "|",
				Next: "GSQ|",
			},
			wantGeom: false, // should be null
			wantErr:  false,
		},
		{
			name: "get features with null geometry",
			fields: fields{
				backend:          newTestGeoPackage("/testdata/null-empty-geoms.gpkg"),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*common.Table{"ligplaatsen": {Name: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     60 * time.Second,
			},
			args: args{
				ctx:        t.Context(),
				collection: "ligplaatsen",
				queryParams: datasources.FeaturesCriteria{
					Cursor: domain.DecodedCursor{FID: 6436, FiltersChecksum: []byte{}},
					Limit:  1,
				},
			},
			wantFC: &domain.FeatureCollection{
				NumberReturned: 1,
				Features: []*domain.Feature{
					{
						Properties: domain.NewFeaturePropertiesWithData(false, map[string]any{
							"straatnaam": "Bokkinghangen",
							"nummer_id":  "0363200012163629",
						}),
					},
				},
			},
			wantCursor: domain.Cursors{
				Prev: "DdY|",
				Next: "|",
			},
			wantGeom: false, // should be null
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GeoPackage{
				backend: tt.fields.backend,
				DatasourceCommon: common.DatasourceCommon{
					FidColumn:           tt.fields.fidColumn,
					TableByCollectionID: tt.fields.featureTableByID,
					QueryTimeout:        tt.fields.queryTimeout,
				},
			}
			g.preparedStmtCache = NewCache()
			url, _ := neturl.Parse("http://example.com")
			s, err := domain.NewSchema([]domain.Field{}, tt.fields.fidColumn, "")
			require.NoError(t, err)
			p := domain.NewProfile(domain.RelAsLink, *url, *s)
			fc, cursor, err := g.GetFeatures(tt.args.ctx, tt.args.collection, tt.args.queryParams, domain.AxisOrderXY, p)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("GetFeatures, error %v, wantErr %v", err, tt.wantErr)
				}

				return
			}
			assert.Equal(t, tt.wantFC.NumberReturned, fc.NumberReturned)
			assert.Equal(t, len(tt.wantFC.Features), fc.NumberReturned)
			for i, wantedFeature := range tt.wantFC.Features {
				assert.Equal(t, wantedFeature.Properties.Value("straatnaam"), fc.Features[i].Properties.Value("straatnaam"))
				assert.Equal(t, wantedFeature.Properties.Value("nummer_id"), fc.Features[i].Properties.Value("nummer_id"))
				if !tt.wantGeom {
					assert.Nil(t, fc.Features[i].Geometry)
				}
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
		featureTableByID map[string]*common.Table
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
				backend:          newTestGeoPackage("/testdata/bag.gpkg"),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*common.Table{"ligplaatsen": {Name: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        t.Context(),
				collection: "ligplaatsen",
				featureID:  3837,
			},
			want: &domain.Feature{
				ID:    "0",
				Links: nil,
				Properties: domain.NewFeaturePropertiesWithData(false, map[string]any{
					"straatnaam": "Realengracht",
					"nummer_id":  "0363200000398886",
				}),
			},
			wantErr: false,
		},
		{
			name: "get non existing feature",
			fields: fields{
				backend:          newTestGeoPackage("/testdata/bag.gpkg"),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*common.Table{"ligplaatsen": {Name: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        t.Context(),
				collection: "ligplaatsen",
				featureID:  999991111111111111,
			},
			want:    nil,
			wantErr: false, // not an error situation
		},
		{
			name: "fail on non existing collection",
			fields: fields{
				backend:          newTestGeoPackage("/testdata/bag.gpkg"),
				fidColumn:        "feature_id",
				featureTableByID: map[string]*common.Table{"ligplaatsen": {Name: "ligplaatsen", GeometryColumnName: "geom"}},
				queryTimeout:     5 * time.Second,
			},
			args: args{
				ctx:        t.Context(),
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
				backend: tt.fields.backend,
				DatasourceCommon: common.DatasourceCommon{
					FidColumn:           tt.fields.fidColumn,
					TableByCollectionID: tt.fields.featureTableByID,
					QueryTimeout:        tt.fields.queryTimeout,
				},
			}
			url, _ := neturl.Parse("http://example.com")
			s, err := domain.NewSchema([]domain.Field{}, tt.fields.fidColumn, "")
			require.NoError(t, err)
			p := domain.NewProfile(domain.RelAsLink, *url, *s)
			got, err := g.GetFeature(tt.args.ctx, tt.args.collection, tt.args.featureID, 0, domain.AxisOrderXY, p)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("GetFeature, error %v, wantErr %v", err, tt.wantErr)
				}

				return
			}
			if tt.want != nil {
				assert.Equal(t, tt.want.Properties.Value("straatnaam"), got.Properties.Value("straatnaam"))
				assert.Equal(t, tt.want.Properties.Value("nummer_id"), got.Properties.Value("nummer_id"))
			}
		})
	}
}

func TestGeoPackage_Warmup(t *testing.T) {
	t.Run("warmup", func(t *testing.T) {
		g := &GeoPackage{
			backend: newTestGeoPackage("/testdata/bag.gpkg"),
			DatasourceCommon: common.DatasourceCommon{
				FidColumn:           "feature_id",
				TableByCollectionID: map[string]*common.Table{"ligplaatsen": {Name: "ligplaatsen", GeometryColumnName: "geom"}},
				QueryTimeout:        5 * time.Second,
			},
		}
		collections :=
			[]config.FeaturesCollection{
				{
					ID: "ligplaatsen",
				},
			}
		err := warmUpFeatureTables(collections, g.TableByCollectionID, g.backend.getDB())
		require.NoError(t, err)
	})
}
