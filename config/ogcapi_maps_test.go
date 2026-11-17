package config

import (
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
			name: "same style on multiple layers is returned once",
			collection: MapsCollection{ID: "c", Layers: []MapLayer{
				{ID: "l1", Styles: []MapStyle{{ID: "a", Title: "A"}}},
				{ID: "l2", Styles: []MapStyle{{ID: "a"}, {ID: "b"}}},
			}},
			want: []MapStyle{{ID: "a", Title: "A"}, {ID: "b"}},
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

	require.Error(t, v.Struct(MapStyle{ID: "a", WMSStyle: "a"}))
	require.NoError(t, v.Struct(MapStyle{ID: "a", WMSStyle: "a", Title: "A"}))
}
