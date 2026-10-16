package maps

import (
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/stretchr/testify/require"
)

var mapsConfig = &config.OgcAPIMaps{
	WMSServer: &config.URL{
		URL: &url.URL{
			Scheme: "https",
			Host:   "example.com",
			Path:   "/wms",
		},
	},
	DefaultMap: config.DefaultMap{
		CRS:    "EPSG:28992",
		Width:  800,
		Height: 600,
		Bbox:   []string{"10000", "300000", "280000", "620000"},
	},
	Collections: []config.MapsCollection{
		{
			ID:         "collection1",
			DefaultCRS: "EPSG:28992",
			Bbox:       []string{"10000", "300000", "280000", "620000"},
			Layers: []config.MapLayer{
				{
					ID:   "layer1",
					Name: "layer1",
					Styles: []config.MapStyle{
						{ID: "default", WMSStyle: "default"},
						{ID: "terrain", WMSStyle: "terrain"},
					},
				},
			},
		},
	},
}

func TestBuildGetMapRequest(t *testing.T) {
	baselineMapRequest := MapRequest{
		mapType:     CollectionMap,
		collection:  "collection1",
		style:       "default",
		crs:         "EPSG:4326",
		bboxCRS:     4326,
		bbox:        &[4]float64{52.05, 5.05, 52.14, 5.18},
		width:       1200,
		height:      900,
		transparent: true,
		format:      "image/png",
	}
	const baselineWMSRequest = "SERVICE=WMS" +
		"&VERSION=1.3.0" +
		"&REQUEST=GetMap" +
		"&BBOX=52.050000,5.050000,52.140000,5.180000" +
		"&CRS=EPSG:4326" +
		"&LAYERS=layer1" +
		"&STYLES=default" +
		"&TRANSPARENT=true" +
		"&FORMAT=image/png" +
		"&HEIGHT=900" +
		"&WIDTH=1200"

	tests := []struct {
		name          string
		change        func(*MapRequest)
		wantParams    url.Values
		wantError     string
		bboxTolerance float64 // Allow small rounding differences after reprojection 
	}{
		{
			name: "baseline test",
		},
		{
			name:       "terrain style",
			change:     func(request *MapRequest) { request.style = "terrain" },
			wantParams: url.Values{"STYLES": {"terrain"}},
		},
		{
			name:       "empty style uses WMS default",
			change:     func(request *MapRequest) { request.style = "" },
			wantParams: url.Values{"STYLES": {""}},
		},
		{
			name:       "transparency false",
			change:     func(request *MapRequest) { request.transparent = false },
			wantParams: url.Values{"TRANSPARENT": {"false"}},
		},
		{
			name:       "JPEG format",
			change:     func(request *MapRequest) { request.format = "image/jpeg" },
			wantParams: url.Values{"FORMAT": {"image/jpeg"}},
		},
		{
			name: "Utrecht RD bbox with matching CRS",
			change: func(request *MapRequest) {
				request.crs = "EPSG:28992"
				request.bboxCRS = 28992
				request.bbox = &[4]float64{131869, 451318, 140816, 461366}
			},
			wantParams: url.Values{
				"CRS":  {"EPSG:28992"},
				"BBOX": {"131869.000000,451318.000000,140816.000000,461366.000000"},
			},
		},
		{
			name: "dataset map is not supported",
			change: func(request *MapRequest) {
				request.mapType = DatasetMap
			},
			wantError: "dataset maps are not supported",
		},
		{
			name: "Utrecht EPSG4326 bbox transformed to RD",
			change: func(request *MapRequest) {
				request.crs = "EPSG:28992"
			},
			wantParams: url.Values{
				"CRS": {"EPSG:28992"}, "BBOX": {"131869.474,451352.254,140815.394,461332.147"},
			},
			bboxTolerance: 0.01,
		},
		{
			name: "Utrecht RD bbox transformed to EPSG4326",
			change: func(request *MapRequest) {
				request.bboxCRS = 28992
				request.bbox = &[4]float64{131869.474, 451352.254, 140815.394, 461332.147}
			},
			wantParams:    url.Values{"BBOX": {"52.05,5.05,52.14,5.18"}},
			bboxTolerance: 0.000001,
		},
		{
			name: "Utrecht default CRS84 bbox transformed to RD",
			change: func(request *MapRequest) {
				request.crs = "EPSG:28992"
				request.bboxCRS = 0
				request.bbox = &[4]float64{5.05, 52.05, 5.18, 52.14}
			},
			wantParams: url.Values{
				"CRS": {"EPSG:28992"}, "BBOX": {"131869.474,451352.254,140815.394,461332.147"},
			},
			bboxTolerance: 0.01,
		},
		{
			name:      "missing bbox",
			change:    func(request *MapRequest) { request.bbox = nil },
			wantError: "bbox is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := baselineMapRequest
			if tt.change != nil {
				tt.change(&request)
			}
			var originalBBox [4]float64
			if request.bbox != nil {
				originalBBox = *request.bbox
			}

			getMapRequest, err := buildGetMapRequest(&request, mapsConfig)
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				require.Nil(t, getMapRequest)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, getMapRequest)

			expected, err := url.ParseQuery(baselineWMSRequest)
			require.NoError(t, err)
			for key, values := range tt.wantParams {
				expected[key] = values
			}

			actual := getMapRequest.ToQueryParameters()
			t.Logf("WMS request: %s", actual.Encode())
			if tt.bboxTolerance > 0 {
				require.Len(t, actual["BBOX"], 1)
				requireBBoxWithinTolerance(t, expected.Get("BBOX"), actual.Get("BBOX"), tt.bboxTolerance)
				expected.Del("BBOX")
				actual.Del("BBOX")
			}
			require.Equal(t, expected, actual)
			require.Equal(t, originalBBox, *request.bbox)
		})
	}
}

func requireBBoxWithinTolerance(t *testing.T, expected, actual string, tolerance float64) {
	t.Helper()
	expectedValues := strings.Split(expected, ",")
	actualValues := strings.Split(actual, ",")
	require.Len(t, expectedValues, 4)
	require.Len(t, actualValues, 4)
	for index := range expectedValues {
		expectedCoordinate, err := strconv.ParseFloat(expectedValues[index], 64)
		require.NoError(t, err)
		actualCoordinate, err := strconv.ParseFloat(actualValues[index], 64)
		require.NoError(t, err)
		require.InDelta(t, expectedCoordinate, actualCoordinate, tolerance, "BBOX coordinate %d", index)
	}
}
