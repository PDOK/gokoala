package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapsCollection_GetStyles(t *testing.T) {
	tests := []struct {
		name       string
		collection MapsCollection
		want       []MapStyle
	}{
		{
			name:       "no layers",
			collection: MapsCollection{ID: "c"},
			want:       nil,
		},
		{
			name: "layer without styles",
			collection: MapsCollection{ID: "c", Layers: []MapLayer{
				{ID: "l1"},
			}},
			want: nil,
		},
		{
			name: "only styles shared by all layers are returned",
			collection: MapsCollection{ID: "c", Layers: []MapLayer{
				{ID: "l1", Styles: []MapStyle{{ID: "a", Title: "A"}, {ID: "c"}}},
				{ID: "l2", Styles: []MapStyle{{ID: "a"}, {ID: "b"}, {ID: "c"}}},
			}},
			want: []MapStyle{{ID: "a", Title: "A"}, {ID: "c"}},
		},
		{
			name: "no shared styles",
			collection: MapsCollection{ID: "c", Layers: []MapLayer{
				{ID: "l1", Styles: []MapStyle{{ID: "a"}}},
				{ID: "l2", Styles: []MapStyle{{ID: "b"}}},
			}},
			want: nil,
		},
		{
			name: "layer without styles means no shared styles",
			collection: MapsCollection{ID: "c", Layers: []MapLayer{
				{ID: "l1", Styles: []MapStyle{{ID: "a"}}},
				{ID: "l2"},
			}},
			want: nil,
		},
		{
			name: "order of first appearance",
			collection: MapsCollection{ID: "c", Layers: []MapLayer{
				{ID: "l1", Styles: []MapStyle{{ID: "z"}, {ID: "a"}}},
			}},
			want: []MapStyle{{ID: "z"}, {ID: "a"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.GetStyles())
		})
	}
}

func TestConfig_MapsStyleID(t *testing.T) {
	tests := []struct {
		name              string
		serviceIdentifier string
		want              string
	}{
		{name: "lowercased", serviceIdentifier: "BestuurlijkeGebieden", want: "bestuurlijkegebieden_province_standaard"},
		{name: "already lowercase", serviceIdentifier: "bg", want: "bg_province_standaard"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{ServiceIdentifier: tt.serviceIdentifier}
			assert.Equal(t, tt.want, c.MapsStyleID("province", "standaard"))
		})
	}
}

func TestMapStyle_TitleRequired(t *testing.T) {
	v := validator.New()
	require.NoError(t, v.RegisterValidation(lowercaseID, LowercaseID))

	require.Error(t, v.Struct(MapStyle{ID: "a", StyleName: "a", LegendPath: "legend.png"}))
	require.NoError(t, v.Struct(MapStyle{ID: "a", StyleName: "a", LegendPath: "legend.png", Title: "A"}))
}

func TestValidateMapLegendPaths(t *testing.T) {
	dir := t.TempDir()
	legend := filepath.Join(dir, "legend.png")
	require.NoError(t, os.WriteFile(legend, nil, 0o600))

	for _, tt := range []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "empty", wantErr: `legendPath for maps collection "collection", layer "layer", style "style":`},
		{name: "existing file", path: legend},
		{name: "missing file", path: filepath.Join(dir, "missing.png"), wantErr: "missing.png"},
		{name: "directory", path: dir, wantErr: "is not a regular file"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			collections := MapsCollections{{
				ID: "collection",
				Layers: []MapLayer{{
					ID:     "layer",
					Styles: []MapStyle{{ID: "style", LegendPath: tt.path}},
				}},
			}}

			err := validateMapLegendPaths(collections)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestValidateMapCollections_AmbiguousStyleIDs(t *testing.T) {
	coll := func(id, styleID string) MapsCollection {
		return MapsCollection{
			ID:           id,
			SupportedSrs: []MapsSupportedSrs{{Srs: "EPSG:28992"}},
			Layers:       []MapLayer{{ID: "l", Styles: []MapStyle{{ID: styleID}}}},
		}
	}

	require.NoError(t, validateMapCollections([]MapsCollection{coll("a", "x"), coll("b", "x")}, "EPSG:28992"))

	err := validateMapCollections([]MapsCollection{coll("a_b", "c"), coll("a", "b_c")}, "EPSG:28992")
	require.ErrorContains(t, err, "ambiguous global style ID 'a_b_c': 'a_b/c' clashes with 'a/b_c'")
}
