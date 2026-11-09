package engine

import (
	"encoding/json"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	gokoalaconfig "github.com/PDOK/gokoala/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newOpenAPI(t *testing.T) {
	openAPIViaCliArgument, err := filepath.Abs("internal/engine/testdata/ogcapi-tiles-1.modified.json")
	if err != nil {
		t.Fatalf("can't locate testdata %v", err)
	}

	type args struct {
		config      *gokoalaconfig.Config
		openAPIFile string
	}
	tests := []struct {
		name                         string
		args                         args
		expectedStringsInOpenAPISpec []string
	}{
		{
			name: "Test render OpenAPI spec with MINIMAL config",
			args: args{
				config: &gokoalaconfig.Config{
					Version:  "2.3.0",
					Title:    "Test API",
					Abstract: "Test API description",
					BaseURL:  gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: gokoalaconfig.OgcAPI{
						GeoVolumes: nil,
						Tiles:      nil,
						Styles:     nil,
					},
				},
			},
			expectedStringsInOpenAPISpec: []string{
				"Landing page",
				"/conformance",
				"/api",
			},
		},
		{
			name: "Test render OpenAPI spec with OGC Tiles config",
			args: args{
				config: &gokoalaconfig.Config{
					Version:  "2.3.0",
					Title:    "Test API",
					Abstract: "Test API description",
					BaseURL:  gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: gokoalaconfig.OgcAPI{
						Tiles: &gokoalaconfig.OgcAPITiles{
							DatasetTiles: &gokoalaconfig.Tiles{
								TileServer: gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset"}},
							},
						},
					},
				},
			},
			expectedStringsInOpenAPISpec: []string{
				"Landing page",
				"/conformance",
				"/api",
				"Vector Tiles",
				"/tiles/{tileMatrixSetId}",
			},
		},
		{
			name: "Test render OpenAPI spec with OGC Styles config",
			args: args{
				config: &gokoalaconfig.Config{
					Version:  "2.3.0",
					Title:    "Test API",
					Abstract: "Test API description",
					BaseURL:  gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: gokoalaconfig.OgcAPI{
						Styles: &gokoalaconfig.OgcAPIStyles{},
					},
				},
			},
			expectedStringsInOpenAPISpec: []string{
				"Landing page",
				"/conformance",
				"/api",
				"/styles",
				"/styles/{styleId}",
				"/resources",
			},
		},
		{
			name: "Test render OpenAPI spec with OGC GeoVolumes config",
			args: args{
				config: &gokoalaconfig.Config{
					Version:  "2.3.0",
					Title:    "Test API",
					Abstract: "Test API description",
					BaseURL:  gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: gokoalaconfig.OgcAPI{
						GeoVolumes: &gokoalaconfig.OgcAPI3dGeoVolumes{
							TileServer: gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
							Collections: gokoalaconfig.GeoVolumesCollections{
								gokoalaconfig.GeoVolumesCollection{ID: "feature1"},
								gokoalaconfig.GeoVolumesCollection{ID: "feature2"},
							},
						},
					},
				},
			},
			expectedStringsInOpenAPISpec: []string{
				"Landing page",
				"/conformance",
				"/api",
				"Collections",
				"/collections",
			},
		},
		{
			name: "Test render OpenAPI spec with OGC Maps config",
			args: args{
				config: &gokoalaconfig.Config{
					Version:  "2.3.0",
					Title:    "Test API",
					Abstract: "Test API description",
					BaseURL:  gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: gokoalaconfig.OgcAPI{
						Maps: &gokoalaconfig.OgcAPIMaps{
							WMSServer: &gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "wms.foobar.example", Path: "/wms"}},
							MaxWidth:  6000, MaxHeight: 3000,
							DefaultMap: gokoalaconfig.DefaultMap{
								CRS:   "EPSG:28992",
								Width: 1024, Height: 768,
								Bbox: []string{"10000", "300000", "280000", "620000"},
							},
							Collections: []gokoalaconfig.MapsCollection{
								{
									ID: "map1",
									Layers: []gokoalaconfig.MapLayer{
										{ID: "layer1", Name: "layer1", Styles: []gokoalaconfig.MapStyle{{ID: "style1", WMSStyle: "default"}}},
									},
									SupportedSrs: []gokoalaconfig.MapsSupportedSrs{{Srs: "EPSG:28992"}},
								},
								{
									ID:           "map2",
									Layers:       []gokoalaconfig.MapLayer{{ID: "layer2", Name: "layer2"}},
									SupportedSrs: []gokoalaconfig.MapsSupportedSrs{{Srs: "EPSG:28992"}},
								},
							},
						},
					},
				},
			},
			expectedStringsInOpenAPISpec: []string{
				"/collections/map1/map",
				"/collections/map2/map",
				"/collections/map1/styles/{styleId}/map",
				"map1.collection.getMap",
				"map2.collection.getMap",
				"image/png",
				"image/jpeg",
				"f-map",
			},
		},
		{
			name: "Test render OpenAPI spec with OGC Tiles and extra spec provided through CLI for overwrite",
			args: args{
				config: &gokoalaconfig.Config{
					Version:  "2.3.0",
					Title:    "Test API",
					Abstract: "Test API description",
					BaseURL:  gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: gokoalaconfig.OgcAPI{
						Tiles: &gokoalaconfig.OgcAPITiles{
							DatasetTiles: &gokoalaconfig.Tiles{
								TileServer: gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset"}},
							},
						},
					},
				},
				openAPIFile: openAPIViaCliArgument,
			},
			expectedStringsInOpenAPISpec: []string{
				"Test API",
				"/conformance",
				"/api",
				"Vector Tiles",
				"/tiles/{tileMatrixSetId}",
				"Map Tiles",                             // extra from given spec through CLI
				"/collections/{collectionId}/map/tiles", // extra from given spec through CLI
			},
		},
		{
			name: "Test render OpenAPI spec with ALL OGC APIs (common, tiles, styles, features, geovolumes, maps)",
			args: args{
				config: &gokoalaconfig.Config{
					Version:  "2.3.0",
					Title:    "Test API",
					Abstract: "Test API description",
					BaseURL:  gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: gokoalaconfig.OgcAPI{
						GeoVolumes: &gokoalaconfig.OgcAPI3dGeoVolumes{
							TileServer: gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
							Collections: gokoalaconfig.GeoVolumesCollections{
								gokoalaconfig.GeoVolumesCollection{ID: "feature1"},
								gokoalaconfig.GeoVolumesCollection{ID: "feature2"},
							},
						},
						Tiles: &gokoalaconfig.OgcAPITiles{
							DatasetTiles: &gokoalaconfig.Tiles{
								TileServer: gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset"}},
							},
						},
						Styles: &gokoalaconfig.OgcAPIStyles{},
						Features: &gokoalaconfig.OgcAPIFeatures{
							Limit: gokoalaconfig.Limit{
								Default: 20,
								Max:     2000,
							},
							Collections: gokoalaconfig.FeaturesCollections{
								{
									ID: "foobar",
									Datasources: &gokoalaconfig.Datasources{
										DefaultWGS84: &gokoalaconfig.Datasource{
											GeoPackage: &gokoalaconfig.GeoPackage{
												Local: &gokoalaconfig.GeoPackageLocal{
													File: "./examples/resources/addresses-crs84.gpkg",
												},
											},
										},
									},
								},
							},
						},
						Maps: &gokoalaconfig.OgcAPIMaps{
							WMSServer: &gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "wms.foobar.example", Path: "/wms"}},
							MaxWidth:  4000, MaxHeight: 4000,
							DefaultMap: gokoalaconfig.DefaultMap{
								CRS:   "EPSG:28992",
								Width: 1024, Height: 768,
								Bbox: []string{"10000", "300000", "280000", "620000"},
							},
							Collections: []gokoalaconfig.MapsCollection{
								{
									ID:           "map1",
									Layers:       []gokoalaconfig.MapLayer{{ID: "layer1", Name: "layer1"}},
									SupportedSrs: []gokoalaconfig.MapsSupportedSrs{{Srs: "EPSG:28992"}},
								},
							},
						},
					},
				},
			},
			expectedStringsInOpenAPISpec: []string{
				"Landing page",
				"/conformance",
				"/api",
				"Vector Tiles",
				"Features",
				"date and time when the collection was last updated",
				"/collections/map1/map",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			openAPI := newOpenAPI(test.args.config, []string{test.args.openAPIFile}, nil)
			assert.NotNil(t, openAPI)

			if maps := test.args.config.OgcAPI.Maps; maps != nil {
				var spec struct {
					Info map[string]json.RawMessage `json:"info"`
				}
				require.NoError(t, json.Unmarshal(openAPI.SpecJSON, &spec))
				var limits struct {
					Maps map[string]int64 `json:"maps"`
				}
				require.NoError(t, json.Unmarshal(spec.Info["x-ogc-limits"], &limits))
				require.Equal(t, map[string]int64{
					"maxWidth": int64(maps.MaxWidth), "maxHeight": int64(maps.MaxHeight), "maxPixels": maps.MaxPixels(),
				}, limits.Maps)
				for name, maximum := range map[string]int{"width": maps.MaxWidth, "height": maps.MaxHeight} {
					schema := openAPI.spec.Components.Parameters[name].Value.Schema.Value
					require.NotNil(t, schema.Max)
					require.InDelta(t, float64(maximum), *schema.Max, 1e-9)
					require.NotNil(t, schema.Min)
					require.InDelta(t, float64(1), *schema.Min, 1e-9)
				}
				for _, item := range openAPI.spec.Paths.Map() {
					if item.Get == nil || !strings.HasSuffix(item.Get.OperationID, ".getMap") {
						continue
					}
					response := item.Get.Responses.Status(413)
					require.NotNil(t, response)
					require.NotNil(t, response.Value.Content["application/problem+json"])
				}
			}

			// verify resulting OpenAPI spec contains expected strings (keywords, paths, etc)
			for _, expectedStr := range test.expectedStringsInOpenAPISpec {
				assert.Contains(t, string(openAPI.SpecJSON), expectedStr,
					"\"%s\" not found in spec: %s", expectedStr, string(openAPI.SpecJSON))
			}
		})
	}
}
