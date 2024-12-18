package config

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
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
			name: "fail on invalid config file",
			args: args{
				configFile: "internal/engine/testdata/config_invalid.yaml",
			},
			wantErr:    true,
			wantErrMsg: "validation for 'Version' failed on the 'semver' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewConfig(tt.args.configFile)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
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
			assert.NoError(t, err)
			var actual []string
			if tt.args.expectedTitles {
				actual = Map(config.AllCollections(), func(item GeoSpatialCollection) string { return *item.Metadata.Title })
			} else {
				actual = Map(config.AllCollections(), func(item GeoSpatialCollection) string { return item.ID })
			}
			assert.ElementsMatch(t, actual, tt.args.expectedOrder)
		})
	}
}

func TestGeoSpatialCollections_Unique(t *testing.T) {
	type args struct {
		configFile            string
		nrOfCollections       int
		nrOfUniqueCollections int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "should filter duplicate collections when calling Unique()",
			args: args{
				configFile:            "internal/engine/testdata/config_collections_unique.yaml",
				nrOfCollections:       3,
				nrOfUniqueCollections: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(tt.args.configFile)
			assert.NoError(t, err)
			assert.Len(t, config.AllCollections(), tt.args.nrOfCollections)
			assert.Len(t, config.AllCollections().Unique(), tt.args.nrOfUniqueCollections)
		})
	}
}

func TestGeoSpatialCollections_ContainsID(t *testing.T) {
	tests := []struct {
		name string
		g    GeoSpatialCollections
		id   string
		want bool
	}{
		{
			name: "ID is present",
			g: []GeoSpatialCollection{
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
			g: []GeoSpatialCollection{
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
			assert.Equalf(t, tt.want, tt.g.ContainsID(tt.id), "ContainsID(%v)", tt.id)
		})
	}
}

func TestProjectionsForCollections(t *testing.T) {
	oaf := OgcAPIFeatures{
		Datasources: &Datasources{
			DefaultWGS84: Datasource{},
			Additional: []AdditionalDatasource{
				{Srs: "EPSG:4355"},
			},
		},
		Collections: GeoSpatialCollections{
			GeoSpatialCollection{
				ID: "coll1",
				Features: &CollectionEntryFeatures{
					Datasources: &Datasources{
						DefaultWGS84: Datasource{},
						Additional: []AdditionalDatasource{
							{Srs: "EPSG:4326"},
							{Srs: "EPSG:3857"},
							{Srs: "EPSG:3857"},
						},
					},
				},
			},
		},
	}

	expected := []string{"EPSG:3857", "EPSG:4326", "EPSG:4355"}
	assert.Equal(t, expected, oaf.ProjectionsForCollections())
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
					Path: ptrTo("/tmp"),
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
				assert.Fail(t, "error = %v, wantErr %v", err, tt.wantErr)
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
			coll: GeoSpatialCollection{
				ID: "test i",
				Metadata: &GeoSpatialCollectionMetadata{
					Description: ptrTo("test d"),
				},
				GeoVolumes: &CollectionEntry3dGeoVolumes{
					TileServerPath: ptrTo("test p"),
				},
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

type TestEmbeddedGeoSpatialCollection struct {
	C GeoSpatialCollection `json:"C"`
}

func TestGeoSpatialCollection_Unmarshalling_JSON(t *testing.T) {
	tests := []struct {
		marshalled string
		want       *GeoSpatialCollection
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			// language=json
			marshalled: `{"id": "test i", "metadata": {"description": "test d"}, "tileServerPath":  "test p"}`,
			want: &GeoSpatialCollection{
				ID: "test i",
				Metadata: &GeoSpatialCollectionMetadata{
					Description: ptrTo("test d"),
				},
				GeoVolumes: &CollectionEntry3dGeoVolumes{
					TileServerPath: ptrTo("test p"),
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			unmarshalled := &GeoSpatialCollection{}
			err := json.Unmarshal([]byte(tt.marshalled), unmarshalled)
			if !tt.wantErr(t, err, errors.New("json.Unmarshal")) {
				return
			}
			assert.Equalf(t, tt.want, unmarshalled, "json.Unmarshal")

			// non-pointer
			unmarshalledEmbedded := &TestEmbeddedGeoSpatialCollection{}
			err = json.Unmarshal([]byte(`{"C": `+tt.marshalled+`}`), unmarshalledEmbedded)
			if !tt.wantErr(t, err, errors.New("json.Unmarshal")) {
				return
			}
			assert.EqualValuesf(t, &TestEmbeddedGeoSpatialCollection{C: *tt.want}, unmarshalledEmbedded, "json.Unmarshal")
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
				Collections: GeoSpatialCollections{
					GeoSpatialCollection{
						Tiles: &CollectionEntryTiles{
							GeoDataTiles: Tiles{
								Types: []TilesType{"raster", "vector"},
							},
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
				Collections: GeoSpatialCollections{
					GeoSpatialCollection{
						Tiles: &CollectionEntryTiles{
							GeoDataTiles: Tiles{
								Types: []TilesType{"raster", "vector"},
							},
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
				Collections: []GeoSpatialCollection{},
			},
			srs:      "EPSG:4326",
			expected: true,
		},
		{
			name: "SRS found in a Collection Tiles",
			ogcAPITiles: OgcAPITiles{
				DatasetTiles: nil,
				Collections: []GeoSpatialCollection{
					{
						Tiles: &CollectionEntryTiles{
							GeoDataTiles: Tiles{
								SupportedSrs: []SupportedSrs{
									{Srs: "EPSG:28992"},
								},
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
				Collections: []GeoSpatialCollection{},
			},
			srs:      "EPSG:9999",
			expected: false,
		},
		{
			name: "Empty Top-level tiles and Collections",
			ogcAPITiles: OgcAPITiles{
				DatasetTiles: nil,
				Collections:  []GeoSpatialCollection{},
			},
			srs:      "EPSG:4326",
			expected: false,
		},
		{
			name: "Handle nil",
			ogcAPITiles: OgcAPITiles{
				DatasetTiles: nil,
				Collections: []GeoSpatialCollection{
					{
						Tiles: nil,
					},
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

func ptrTo[T any](val T) *T {
	return &val
}

func Map[T, V any](collection []T, fn func(T) V) []V {
	result := make([]V, len(collection))
	for i, t := range collection {
		result[i] = fn(t)
	}
	return result
}
