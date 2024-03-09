package engine

import (
	"net/url"
	"path/filepath"
	"testing"

	gokoalaconfig "github.com/PDOK/gokoala/config"

	"github.com/stretchr/testify/assert"
)

func Test_newOpenAPI(t *testing.T) {
	openAPIViaCliArgument, err := filepath.Abs("engine/testdata/ogcapi-tiles-1.bundled.json")
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
							TileServer: gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset"}},
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
							Collections: gokoalaconfig.GeoSpatialCollections{
								gokoalaconfig.GeoSpatialCollection{ID: "feature1"},
								gokoalaconfig.GeoSpatialCollection{ID: "feature2"},
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
			name: "Test render OpenAPI spec with OGC Tiles and extra spec provided through CLI for overwrite",
			args: args{
				config: &gokoalaconfig.Config{
					Version:  "2.3.0",
					Title:    "Test API",
					Abstract: "Test API description",
					BaseURL:  gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: gokoalaconfig.OgcAPI{
						Tiles: &gokoalaconfig.OgcAPITiles{
							TileServer: gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset"}},
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
			name: "Test render OpenAPI spec with ALL OGC APIs (common, tiles, styles, features, geovolumes)",
			args: args{
				config: &gokoalaconfig.Config{
					Version:  "2.3.0",
					Title:    "Test API",
					Abstract: "Test API description",
					BaseURL:  gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: gokoalaconfig.OgcAPI{
						GeoVolumes: &gokoalaconfig.OgcAPI3dGeoVolumes{
							TileServer: gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
							Collections: gokoalaconfig.GeoSpatialCollections{
								gokoalaconfig.GeoSpatialCollection{ID: "feature1"},
								gokoalaconfig.GeoSpatialCollection{ID: "feature2"},
							},
						},
						Tiles: &gokoalaconfig.OgcAPITiles{
							TileServer: gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset"}},
						},
						Styles: &gokoalaconfig.OgcAPIStyles{},
						Features: &gokoalaconfig.OgcAPIFeatures{
							Limit: gokoalaconfig.Limit{
								Default: 20,
								Max:     2000,
							},
							Collections: []gokoalaconfig.GeoSpatialCollection{
								{
									ID: "foobar",
									Features: &gokoalaconfig.CollectionEntryFeatures{
										Datasources: &gokoalaconfig.Datasources{
											DefaultWGS84: gokoalaconfig.Datasource{
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
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			openAPI := newOpenAPI(test.args.config, []string{test.args.openAPIFile}, nil)
			assert.NotNil(t, openAPI)

			// verify resulting OpenAPI spec contains expected strings (keywords, paths, etc)
			for _, expectedStr := range test.expectedStringsInOpenAPISpec {
				assert.Contains(t, string(openAPI.SpecJSON), expectedStr,
					"\"%s\" not found in spec: %s", expectedStr, string(openAPI.SpecJSON))
			}
		})
	}
}
