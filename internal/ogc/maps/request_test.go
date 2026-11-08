package maps

import (
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-geom"
)

func TestMapRequestResolutionPreservesRequested(t *testing.T) {
	for _, tt := range []struct {
		name      string
		requested MapViewParameters
		want      MapViewParameters
		defaults  *config.DefaultMap
	}{
		{
			name: "all defaults",
			want: MapViewParameters{crs: rd28992URI, bboxSRID: 28992,
				bbox:  geom.NewBounds(geom.XY).Set(10000, 300000, 280000, 620000),
				width: 800, height: 600, mmPerPixel: 0.28},
		},
		{
			name: "explicit values and distorted view unchanged",
			requested: MapViewParameters{crs: rd28992URI, bboxSRID: 28992,
				bbox: geom.NewBounds(geom.XY).Set(0, 0, 200, 100), width: 100, height: 100, mmPerPixel: 0.254},
			want: MapViewParameters{crs: rd28992URI, bboxSRID: 28992,
				bbox: geom.NewBounds(geom.XY).Set(0, 0, 200, 100), width: 100, height: 100, mmPerPixel: 0.254},
		},
		{
			name: "CRS84 input to EPSG4326 native axis order",
			requested: MapViewParameters{crs: crsEPSG4326URI,
				bbox: geom.NewBounds(geom.XY).Set(4, 50, 6, 51), width: 123},
			want: MapViewParameters{crs: crsEPSG4326URI, bboxSRID: 4326,
				bbox: geom.NewBounds(geom.XY).Set(50, 4, 51, 6), width: 123, height: 62, mmPerPixel: 0.28},
		},
		{
			name: "EPSG4326 input to CRS84",
			requested: MapViewParameters{crs: crs84URI, bboxSRID: 4326,
				bbox: geom.NewBounds(geom.XY).Set(50, 4, 51, 6), height: 345},
			want: MapViewParameters{crs: crs84URI, bboxSRID: domain.WGS84SRID,
				bbox: geom.NewBounds(geom.XY).Set(4, 50, 6, 51), width: 690, height: 345, mmPerPixel: 0.28},
		},
		{
			name: "default bbox reprojected to requested CRS",
			defaults: &config.DefaultMap{CRS: "EPSG:4326", Bbox: []string{"50", "4", "51", "6"},
				Width: 800, Height: 600, MMPerPixel: "0.28"},
			requested: MapViewParameters{crs: crs84URI},
			want: MapViewParameters{crs: crs84URI, bboxSRID: domain.WGS84SRID,
				bbox: geom.NewBounds(geom.XY).Set(4, 50, 6, 51), width: 800, height: 600, mmPerPixel: 0.28},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defaults := mapsConfig.DefaultMap
			if tt.defaults != nil {
				defaults = *tt.defaults
			}
			request := MapRequest{requested: tt.requested}
			var originalBBox *geom.Bounds
			if tt.requested.bbox != nil {
				b := tt.requested.bbox
				originalBBox = geom.NewBounds(geom.XY).Set(b.Min(0), b.Min(1), b.Max(0), b.Max(1))
			}
			require.NoError(t, request.setDefaults(defaults))
			require.Equal(t, tt.requested, request.requested)
			require.Equal(t, originalBBox, request.requested.bbox)
			require.Equal(t, tt.want, request.resolved)
			if originalBBox != nil {
				require.NotSame(t, request.requested.bbox, request.resolved.bbox)
			}
			require.NoError(t, request.setDefaults(defaults))
			require.Equal(t, tt.want, request.resolved, "repeated resolution must start from requested values")
			request.requested.mmPerPixel = 0.127
			require.NoError(t, request.setDefaults(defaults))
			require.InDelta(t, 0.127, request.resolved.mmPerPixel, 1e-9)
		})
	}
}
