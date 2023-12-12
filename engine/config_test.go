package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
						},
					},
				},
			},
		},
	}

	expected := []string{"EPSG:3857", "EPSG:4326", "EPSG:4355"}
	assert.Equal(t, expected, oaf.ProjectionsForCollections())
}

func ptrTo[T any](val T) *T {
	return &val
}
