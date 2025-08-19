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
				expectedOrder: []string{"A", "B", "C", "Z", "Z"},
			},
		},
		{
			name: "should return collections in default order (alphabetic) - by title",
			args: args{
				configFile:    "internal/engine/testdata/config_collections_order_alphabetic_titles.yaml",
				expectedOrder: []string{"B", "C", "Z", "Z", "A"},
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
