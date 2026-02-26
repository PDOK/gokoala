package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/PDOK/gokoala/internal/engine/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template/config files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestNewConfig(t *testing.T) {
	type args struct {
		configFile string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "read valid config file",
			args: args{
				configFile: "internal/engine/testdata/config_minimal.yaml",
			},
			wantErr: false,
		},
		{
			name: "read valid config file with all ogc apis",
			args: args{
				configFile: "examples/config_all.yaml",
			},
			wantErr: false,
		},
		{
			name: "fail on invalid config file with wrong version number",
			args: args{
				configFile: "internal/engine/testdata/config_invalid.yaml",
			},
			wantErr:    true,
			wantErrMsg: "validation for 'Version' failed on the 'semver' tag",
		},
		{
			name: "fail on invalid config file with wrong collection IDs",
			args: args{
				configFile: "internal/engine/testdata/config_invalid_collection_ids.yaml",
			},
			wantErr:    true,
			wantErrMsg: "Field validation for 'ID' failed on the 'lowercase_id' tag",
		},
		{
			name: "read config file with valid collection IDs",
			args: args{
				configFile: "internal/engine/testdata/config_valid_collection_ids.yaml",
			},
			wantErr: false,
		},
		{
			name: "fail on invalid config with unsupported tile projection",
			args: args{
				configFile: "internal/engine/testdata/config_invalid_tiles_projection.yaml",
			},
			wantErr:    true,
			wantErrMsg: "validation failed for srs 'EPSG:99999'; srs is not supported",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewConfig(tt.args.configFile)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAllCollections(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		expectedOrder []string
	}{
		{
			name: "should return all collections from different OGC APIs",
			config: &Config{
				OgcAPI: OgcAPI{
					GeoVolumes: &OgcAPI3dGeoVolumes{
						Collections: []GeoVolumesCollection{{ID: "volumes"}},
					},
					Tiles: &OgcAPITiles{
						Collections: []TilesCollection{{ID: "tiles"}},
					},
					Features: &OgcAPIFeatures{
						Collections: []FeaturesCollection{{ID: "features"}},
					},
				},
			},
			expectedOrder: []string{"features", "tiles", "volumes"}, // Alphabetical default
		},
		{
			name: "should respect literal order when OgcAPICollectionOrder is provided",
			config: &Config{
				OgcAPICollectionOrder: []string{"tiles", "volumes", "features"},
				OgcAPI: OgcAPI{
					GeoVolumes: &OgcAPI3dGeoVolumes{
						Collections: []GeoVolumesCollection{{ID: "volumes"}},
					},
					Tiles: &OgcAPITiles{
						Collections: []TilesCollection{{ID: "tiles"}},
					},
					Features: &OgcAPIFeatures{
						Collections: []FeaturesCollection{{ID: "features"}},
					},
				},
			},
			expectedOrder: []string{"tiles", "volumes", "features"},
		},
		{
			name: "should sort by title if available and no literal order",
			config: &Config{
				OgcAPI: OgcAPI{
					Features: &OgcAPIFeatures{
						Collections: []FeaturesCollection{
							{ID: "b", Metadata: &GeoSpatialCollectionMetadata{Title: new("Z Title")}},
							{ID: "a", Metadata: &GeoSpatialCollectionMetadata{Title: new("A Title")}},
						},
					},
				},
			},
			expectedOrder: []string{"a", "b"}, // "A Title" < "Z Title"
		},
		{
			name: "should handle empty config",
			config: &Config{
				OgcAPI: OgcAPI{},
			},
			expectedOrder: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collections := tt.config.AllCollections()
			actualIDs := Map(collections, func(c GeoSpatialCollection) string {
				return c.GetID()
			})
			assert.Equal(t, tt.expectedOrder, actualIDs)
		})
	}
}

func TestGeoSpatialCollections_Ordering(t *testing.T) {
	type args struct {
		configFile     string
		expectedOrder  []string
		expectedTitles bool // ids are default, when true titles are used
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "should return collections in default order (alphabetic)",
			args: args{
				configFile:    "internal/engine/testdata/config_collections_order_alphabetic.yaml",
				expectedOrder: []string{"a", "b", "c", "z", "z"},
			},
		},
		{
			name: "should return collections in default order (alphabetic) - by title",
			args: args{
				configFile:    "internal/engine/testdata/config_collections_order_alphabetic_titles.yaml",
				expectedOrder: []string{"b", "c", "z", "z", "a"},
			},
		},
		{
			name: "should return collections in default order (alphabetic) - extra test",
			args: args{
				configFile:    "internal/engine/testdata/config_collections_unique.yaml",
				expectedOrder: []string{"bar_collection", "foo_collection", "foo_collection"},
			},
		},
		{
			name: "should return collections in explicit / literal order",
			args: args{
				configFile:    "internal/engine/testdata/config_collections_order_literal.yaml",
				expectedOrder: []string{"z", "z", "c", "a", "b"},
			},
		},
		{
			name: "should not error when no collections",
			args: args{
				configFile:    "internal/engine/testdata/config_minimal.yaml",
				expectedOrder: []string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(tt.args.configFile)
			require.NoError(t, err)
			var actual []string
			if tt.args.expectedTitles {
				actual = Map(config.AllCollections(), func(item GeoSpatialCollection) string { return *item.GetMetadata().Title })
			} else {
				actual = Map(config.AllCollections(), func(item GeoSpatialCollection) string { return item.GetID() })
			}
			assert.ElementsMatch(t, actual, tt.args.expectedOrder)
		})
	}
}

func TestGeoSpatialCollections_Unique(t *testing.T) {
	config, err := NewConfig("internal/engine/testdata/config_collections_unique.yaml")
	require.NoError(t, err)

	assert.Len(t, config.AllCollections(), 3)
	assert.Len(t, config.AllCollections().Unique(), 2)
}

func TestGeoSpatialCollections_Unique_WithMetadata(t *testing.T) {
	config, err := NewConfig("internal/engine/testdata/config_collections_unique_with_metadata.yaml")
	require.NoError(t, err)

	assert.Len(t, config.AllCollections(), 3)

	unique := config.AllCollections().Unique()
	assert.Len(t, unique, 2)

	uniqueColl := unique[0]
	assert.Equal(t, "Foo Collection", *uniqueColl.GetMetadata().Title)
	assert.Equal(t, "https://example.com/awesome.zip", uniqueColl.GetLinks().Downloads[0].AssetURL.String())
}

func TestGeoSpatialCollections_Unique_WithMetadataAndMoreLinks(t *testing.T) {
	config, err := NewConfig("internal/engine/testdata/config_collections_unique_with_links.yaml")
	require.NoError(t, err)

	assert.Len(t, config.AllCollections(), 3)

	unique := config.AllCollections().Unique()
	assert.Len(t, unique, 2)

	uniqueColl := unique[0]
	assert.Equal(t, "Foo Collection", *uniqueColl.GetMetadata().Title)
	assert.Equal(t, "https://example.com/awesome.gpkg", uniqueColl.GetLinks().Downloads[0].AssetURL.String())
	assert.Equal(t, "https://example.com/awesome.zip", uniqueColl.GetLinks().Downloads[1].AssetURL.String())
}

func TestGeoSpatialCollections_ContainsID(t *testing.T) {
	tests := []struct {
		name string
		g    []FeaturesCollection
		id   string
		want bool
	}{
		{
			name: "ID is present",
			g: []FeaturesCollection{
				{
					ID: "3",
				},
				{
					ID: "1",
				},
				{
					ID: "2",
				},
			},
			id:   "1",
			want: true,
		},
		{
			name: "ID is not present",
			g: []FeaturesCollection{
				{
					ID: "3",
				},
				{
					ID: "1",
				},
				{
					ID: "2",
				},
			},
			id:   "55",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collections := GeoSpatialCollections(types.ToInterfaceSlice[FeaturesCollection, GeoSpatialCollection](tt.g))
			assert.Equalf(t, tt.want, collections.ContainsID(tt.id), "ContainsID(%v)", tt.id)
		})
	}
}

func TestCollectionsSRS(t *testing.T) {
	oaf := OgcAPIFeatures{
		Datasources: &Datasources{
			DefaultWGS84: &Datasource{},
			Additional: []AdditionalDatasource{
				{Srs: "EPSG:4355"},
			},
		},
		Collections: []FeaturesCollection{
			{
				ID: "coll1",
				Datasources: &Datasources{
					DefaultWGS84: &Datasource{},
					Additional: []AdditionalDatasource{
						{Srs: "EPSG:4326"},
						{Srs: "EPSG:3857"},
						{Srs: "EPSG:3857"},
					},
				},
			},
		},
	}

	expected := []string{"EPSG:3857", "EPSG:4326", "EPSG:4355"}
	assert.Equal(t, expected, oaf.CollectionsSRS())
}

func TestCacheDir(t *testing.T) {
	tests := []struct {
		name    string
		gc      GeoPackageCloud
		wantErr bool
	}{
		{
			name: "With explicit cache path provided",
			gc: GeoPackageCloud{
				File: "test.gpkg",
				Cache: GeoPackageCloudCache{
					Path: new("/tmp"),
				},
			},
			wantErr: false,
		},
		{
			name: "Without explicit cache path provided",
			gc: GeoPackageCloud{
				File: "test.gpkg",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.gc.CacheDir()
			if (err != nil) != tt.wantErr {
				assert.Fail(t, fmt.Sprintf("error = %v, wantErr %v", err, tt.wantErr))
			}
			if tt.gc.Cache.Path == nil {
				assert.DirExists(t, got)
			} else {
				assert.Contains(t, got, *tt.gc.Cache.Path+"/test-")
			}
		})
	}
}

func TestGeoSpatialCollection_Marshalling_JSON(t *testing.T) {
	tests := []struct {
		coll    GeoSpatialCollection
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			coll: &GeoVolumesCollection{
				ID: "test i",
				Metadata: &GeoSpatialCollectionMetadata{
					Description: new("test d"),
				},
				TileServerPath: new("test p"),
			},
			// language=json
			want:    `{"id": "test i", "metadata": {"description": "test d"}, "tileServerPath":  "test p"}`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			marshalled, err := json.Marshal(tt.coll)
			if !tt.wantErr(t, err, errors.New("json.Marshal")) {
				return
			}
			assert.JSONEqf(t, tt.want, string(marshalled), "json.Marshal")
		})
	}
}

func TestOgcAPITiles_HasType(t *testing.T) {
	tests := []struct {
		name      string
		tiles     OgcAPITiles
		tilesType TilesType
		expected  bool
	}{
		{
			name: "Has type in DatasetTiles",
			tiles: OgcAPITiles{
				DatasetTiles: &Tiles{
					Types: []TilesType{"raster", "vector"},
				},
			},
			tilesType: "raster",
			expected:  true,
		},
		{
			name: "Does not have type in DatasetTiles",
			tiles: OgcAPITiles{
				DatasetTiles: &Tiles{
					Types: []TilesType{"raster", "vector"},
				},
			},
			tilesType: "some-other-type",
			expected:  false,
		},
		{
			name: "Has type in Collections",
			tiles: OgcAPITiles{
				Collections: []TilesCollection{
					{
						GeoDataTiles: Tiles{
							Types: []TilesType{"raster", "vector"},
						},
					},
				},
			},
			tilesType: "raster",
			expected:  true,
		},
		{
			name: "Does not have type in Collections",
			tiles: OgcAPITiles{
				Collections: []TilesCollection{
					{
						GeoDataTiles: Tiles{
							Types: []TilesType{"raster", "vector"},
						},
					},
				},
			},
			tilesType: "some-other-type",
			expected:  false,
		},
		{
			name:      "No DatasetTiles and Collections",
			tiles:     OgcAPITiles{},
			tilesType: "raster",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tiles.HasType(tt.tilesType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOgcAPITiles_HasProjection(t *testing.T) {
	tests := []struct {
		name        string
		ogcAPITiles OgcAPITiles
		srs         string
		expected    bool
	}{
		{
			name: "SRS found in Top-level Tiles",
			ogcAPITiles: OgcAPITiles{
				DatasetTiles: &Tiles{
					SupportedSrs: []SupportedSrs{
						{Srs: "EPSG:4326"},
						{Srs: "EPSG:3857"},
					},
				},
				Collections: []TilesCollection{},
			},
			srs:      "EPSG:4326",
			expected: true,
		},
		{
			name: "SRS found in a Collection Tiles",
			ogcAPITiles: OgcAPITiles{
				DatasetTiles: nil,
				Collections: []TilesCollection{
					{
						GeoDataTiles: Tiles{
							SupportedSrs: []SupportedSrs{
								{Srs: "EPSG:28992"},
							},
						},
					},
				},
			},
			srs:      "EPSG:28992",
			expected: true,
		},
		{
			name: "SRS not found",
			ogcAPITiles: OgcAPITiles{
				DatasetTiles: &Tiles{
					SupportedSrs: []SupportedSrs{
						{Srs: "EPSG:4326"},
					},
				},
				Collections: []TilesCollection{},
			},
			srs:      "EPSG:9999",
			expected: false,
		},
		{
			name: "Empty Top-level tiles and Collections",
			ogcAPITiles: OgcAPITiles{
				DatasetTiles: nil,
				Collections:  []TilesCollection{},
			},
			srs:      "EPSG:4326",
			expected: false,
		},
		{
			name: "Handle nil",
			ogcAPITiles: OgcAPITiles{
				DatasetTiles: nil,
				Collections: []TilesCollection{
					{},
				},
			},
			srs:      "EPSG:4326",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ogcAPITiles.HasProjection(tt.srs))
		})
	}
}

func TestFeaturesCollection_MarshalUnmarshal_JSON(t *testing.T) {
	in := FeaturesCollection{
		ID: "buildings",
		FeatureProperties: &FeatureProperties{
			Properties:                []string{"id", "name"},
			PropertiesExcludeUnknown:  true,
			PropertiesInSpecificOrder: true,
		},
		Filters: FeatureFilters{
			Properties: []PropertyFilter{
				{Name: "status"},
			},
		},
	}

	b, err := json.Marshal(in)
	require.NoError(t, err)

	// language=json
	assert.JSONEq(t, `{
		"id": "buildings",
		"filters": { "properties": [ { "name": "status" } ] },
		"properties": ["id", "name"],
		"propertiesExcludeUnknown": true,
		"propertiesInSpecificOrder": true
	}`, string(b))

	var out FeaturesCollection
	require.NoError(t, json.Unmarshal(b, &out))

	require.NotNil(t, out.FeatureProperties, "embedded FeatureProperties should be allocated when fields are present")
	assert.Equal(t, in.ID, out.ID)
	assert.Equal(t, in.Properties, out.Properties)
	assert.Equal(t, in.PropertiesExcludeUnknown, out.PropertiesExcludeUnknown)
	assert.Equal(t, in.PropertiesInSpecificOrder, out.PropertiesInSpecificOrder)

	require.Len(t, out.Filters.Properties, 1)
	assert.Equal(t, "status", out.Filters.Properties[0].Name)
}

func TestFeaturesCollection_MarshalUnmarshal_YAML(t *testing.T) {
	in := FeaturesCollection{
		ID: "roads",
		FeatureProperties: &FeatureProperties{
			Properties:                []string{"id", "type"},
			PropertiesExcludeUnknown:  true,
			PropertiesInSpecificOrder: true,
		},
	}

	yamlBytes, err := yaml.Marshal(in)
	require.NoError(t, err)

	yamlText := string(yamlBytes)
	assert.Contains(t, yamlText, "id: roads")
	assert.Contains(t, yamlText, "properties:\n    - id\n    - type")
	assert.Contains(t, yamlText, "propertiesExcludeUnknown: true")
	assert.Contains(t, yamlText, "propertiesInSpecificOrder: true")

	var out FeaturesCollection
	require.NoError(t, yaml.Unmarshal(yamlBytes, &out))

	require.NotNil(t, out.FeatureProperties, "embedded FeatureProperties should be allocated when fields are present")
	assert.Equal(t, in.ID, out.ID)
	assert.Equal(t, in.Properties, out.Properties)
	assert.Equal(t, in.PropertiesExcludeUnknown, out.PropertiesExcludeUnknown)
	assert.Equal(t, in.PropertiesInSpecificOrder, out.PropertiesInSpecificOrder)
}

func Map[T, V any](collection []T, fn func(T) V) []V {
	result := make([]V, len(collection))
	for i, t := range collection {
		result[i] = fn(t)
	}

	return result
}
