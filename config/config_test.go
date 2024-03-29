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
				configFile: "engine/testdata/config_minimal.yaml",
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
				configFile: "engine/testdata/config_invalid.yaml",
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

func TestGeoSpatialCollections_Unique(t *testing.T) {
	tests := []struct {
		name string
		g    GeoSpatialCollections
		want []GeoSpatialCollection
	}{
		{
			name: "empty input",
			g:    nil,
			want: []GeoSpatialCollection{},
		},
		{
			name: "no dups, sorted by id",
			g: []GeoSpatialCollection{
				{
					ID: "3",
				},
				{
					ID: "1",
				},
				{
					ID: "1",
				},
				{
					ID: "2",
				},
			},
			want: []GeoSpatialCollection{
				{
					ID: "1",
				},
				{
					ID: "2",
				},
				{
					ID: "3",
				},
			},
		},
		{
			name: "no dups, sorted by title",
			g: []GeoSpatialCollection{
				{
					ID: "3",
					Metadata: &GeoSpatialCollectionMetadata{
						Title:         ptrTo("a"),
						LastUpdatedBy: "",
					},
				},
				{
					ID: "1",
					Metadata: &GeoSpatialCollectionMetadata{
						Title:         ptrTo("c"),
						LastUpdatedBy: "",
					},
				},
				{
					ID: "3",
					Metadata: &GeoSpatialCollectionMetadata{
						Title:         ptrTo("a"),
						LastUpdatedBy: "",
					},
				},
				{
					ID: "2",
					Metadata: &GeoSpatialCollectionMetadata{
						Title:         ptrTo("b"),
						LastUpdatedBy: "",
					},
				},
			},
			want: []GeoSpatialCollection{
				{
					ID: "3",
					Metadata: &GeoSpatialCollectionMetadata{
						Title:         ptrTo("a"),
						LastUpdatedBy: "",
					},
				},
				{
					ID: "2",
					Metadata: &GeoSpatialCollectionMetadata{
						Title:         ptrTo("b"),
						LastUpdatedBy: "",
					},
				},
				{
					ID: "1",
					Metadata: &GeoSpatialCollectionMetadata{
						Title:         ptrTo("c"),
						LastUpdatedBy: "",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.g.Unique(), "Unique()")
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
					Title: ptrTo("test t"),
				},
				GeoVolumes: &CollectionEntry3dGeoVolumes{
					TileServerPath: ptrTo("test p"),
				},
			},
			// language=json
			want:    `{"id": "test i", "metadata": {"title": "test t"}, "tileServerPath":  "test p"}`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			marshalled, err := json.Marshal(tt.coll)
			if !tt.wantErr(t, err, errors.New("json.Marshal")) {
				return
			}
			assert.Equalf(t, tt.want, string(marshalled), "json.Marshal")
		})
	}
}

func ptrTo[T any](val T) *T {
	return &val
}
