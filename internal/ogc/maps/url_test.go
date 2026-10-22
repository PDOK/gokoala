package maps

import (
	"net/url"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	crsEPSG4326URI = "http://www.opengis.net/def/crs/EPSG/0/4326"
	crsUnknownURI  = "http://www.opengis.net/def/crs/EPSG/0/999999"
	bboxCRS84      = "4.0,52.0,5.0,53.0"
	bboxRD         = "100000,400000,150000,450000"
	bboxLatLon     = "52.0,4.0,53.0,5.0"
)

type paramCase struct {
	name            string
	query           string
	wantCRS         domain.ContentCrs
	wantBboxSRID    domain.SRID
	wantBbox        []float64
	wantWidth       int
	wantHeight      int
	wantTransparent bool
	wantErr         bool
	knownDeviation  string
}

var paramCases = []paramCase{
	// all single parameters
	{name: "none", query: "", wantCRS: rd28992URI, wantTransparent: true},
	{name: "transparent true", query: "transparent=true", wantTransparent: true},
	{name: "transparent false", query: "transparent=false", wantTransparent: false},
	{name: "transparent invalid", query: "transparent=yes", wantErr: true},
	{name: "transparent empty", query: "transparent=", wantErr: true},
	{name: "transparent numeric", query: "transparent=1", wantErr: true},
	{name: "transparent uppercase true", query: "transparent=TRUE", wantTransparent: true},
	{name: "transparent uppercase false", query: "transparent=FALSE", wantTransparent: false},
	{name: "repeated transparent uses first value", query: "transparent=true&transparent=false", wantTransparent: true},
	{name: "width only", query: "width=800", wantWidth: 800, wantTransparent: true},
	{name: "height only", query: "height=600", wantHeight: 600, wantTransparent: true},
	{name: "crs only", query: "crs=" + rd28992URI, wantCRS: rd28992URI, wantTransparent: true},
	{name: "bbox only", query: "bbox=" + bboxCRS84, wantBboxSRID: domain.WGS84SRID, wantBbox: []float64{4, 52, 5, 53}, wantTransparent: true},
	{name: "bbox-crs only without bbox is ignored", query: "bbox-crs=" + domain.WGS84CrsURI, wantTransparent: true},
	{name: "invalid bbox-crs without bbox is ignored", query: "bbox-crs=foo", wantTransparent: true,
		knownDeviation: "D7: 20-058 /req/spatial-subsetting/bbox-crs F, bbox-crs is ignored when bbox is absent"},

	// combinations
	{name: "bbox + bbox-crs CRS84", query: "bbox=" + bboxCRS84 + "&bbox-crs=" + domain.WGS84CrsURI,
		wantBboxSRID: domain.WGS84SRID, wantBbox: []float64{4, 52, 5, 53}, wantTransparent: true},
	{name: "bbox + bbox-crs EPSG:4326", query: "bbox=" + bboxLatLon + "&bbox-crs=" + crsEPSG4326URI,
		wantBboxSRID: 4326, wantBbox: []float64{52, 4, 53, 5}, wantTransparent: true},
	{name: "crs CRS84", query: "crs=" + domain.WGS84CrsURI, wantCRS: domain.ContentCrs(domain.WGS84CrsURI), wantTransparent: true},
	{name: "bbox + bbox-crs EPSG:28992", query: "bbox=" + bboxRD + "&bbox-crs=" + rd28992URI,
		wantBboxSRID: 28992, wantBbox: []float64{100000, 400000, 150000, 450000}, wantTransparent: true},
	{name: "crs + bbox", query: "crs=" + rd28992URI + "&bbox=" + bboxCRS84,
		wantCRS: rd28992URI, wantBboxSRID: domain.WGS84SRID, wantBbox: []float64{4, 52, 5, 53}, wantTransparent: true},
	{name: "crs + bbox + bbox-crs", query: "crs=" + rd28992URI + "&bbox=" + bboxRD + "&bbox-crs=" + rd28992URI,
		wantCRS: rd28992URI, wantBboxSRID: 28992, wantBbox: []float64{100000, 400000, 150000, 450000}, wantTransparent: true},
	{name: "width + height", query: "width=800&height=600", wantWidth: 800, wantHeight: 600, wantTransparent: true},
	{name: "all five", query: "crs=" + rd28992URI + "&bbox=" + bboxRD + "&bbox-crs=" + rd28992URI + "&width=800&height=600",
		wantCRS: rd28992URI, wantBboxSRID: 28992, wantBbox: []float64{100000, 400000, 150000, 450000},
		wantWidth: 800, wantHeight: 600, wantTransparent: true},

	// errors
	{name: "invalid width non-numeric", query: "width=abc", wantErr: true},
	{name: "invalid width decimal", query: "width=1.5", wantErr: true},
	{name: "invalid width zero", query: "width=0", wantErr: true,
		knownDeviation: "D1: 20-058 /req/scaling/width-definition C, non-positive width must be rejected"},
	{name: "invalid width negative", query: "width=-5", wantErr: true,
		knownDeviation: "D1: 20-058 /req/scaling/width-definition C, non-positive width must be rejected"},
	{name: "invalid height non-numeric", query: "height=abc", wantErr: true},
	{name: "invalid height decimal", query: "height=1.5", wantErr: true},
	{name: "invalid height zero", query: "height=0", wantErr: true,
		knownDeviation: "D1: 20-058 /req/scaling/height-definition C, non-positive height must be rejected"},
	{name: "invalid height negative", query: "height=-5", wantErr: true,
		knownDeviation: "D1: 20-058 /req/scaling/height-definition C, non-positive height must be rejected"},
	{name: "invalid width above max", query: "width=5001", wantErr: true},
	{name: "invalid height above max", query: "height=5001", wantErr: true},
	{name: "invalid crs unknown", query: "crs=" + crsUnknownURI, wantErr: true},
	{name: "invalid bbox 3 values", query: "bbox=1,2,3", wantErr: true},
	{name: "invalid bbox 5 values", query: "bbox=1,2,3,4,5", wantErr: true},
	{name: "invalid bbox non-numeric", query: "bbox=a,b,c,d", wantErr: true},
	{name: "invalid bbox min greater than max", query: "bbox=4,53,5,52", wantErr: true},
	{name: "invalid bbox min x greater than max x", query: "bbox=179,52,-179,53", wantErr: true},
	{name: "invalid bbox zero area", query: "bbox=4,52,4,53", wantErr: true},
	{name: "invalid bbox-crs unknown with bbox", query: "bbox=" + bboxCRS84 + "&bbox-crs=" + crsUnknownURI, wantErr: true},
	{name: "invalid bbox-crs not a URI with bbox", query: "bbox=" + bboxCRS84 + "&bbox-crs=foo", wantErr: true},

	// errors but combined with valid params
	{name: "invalid width zero + valid others", query: "crs=" + rd28992URI + "&bbox=" + bboxCRS84 + "&height=600&width=0",
		wantErr: true, knownDeviation: "D1: 20-058 /req/scaling/width-definition C"},
	{name: "invalid width non-numeric + valid others", query: "crs=" + rd28992URI + "&bbox=" + bboxCRS84 + "&height=600&width=abc",
		wantErr: true},
	{name: "invalid width above max + valid others", query: "crs=" + rd28992URI + "&bbox=" + bboxCRS84 + "&height=600&width=9000",
		wantErr: true},
	{name: "invalid height above max + valid others", query: "crs=" + rd28992URI + "&bbox=" + bboxCRS84 + "&width=800&height=9000",
		wantErr: true},
	{name: "invalid crs + valid bbox and width", query: "crs=" + crsUnknownURI + "&bbox=" + bboxCRS84 + "&width=800",
		wantErr: true},
	{name: "invalid bbox + valid crs and size", query: "crs=" + rd28992URI + "&bbox=1,2,3&width=800&height=600",
		wantErr: true},
	{name: "invalid bbox-crs + valid crs, bbox and size", query: "crs=" + rd28992URI + "&bbox=" + bboxCRS84 + "&bbox-crs=foo&width=800&height=600",
		wantErr: true},

	{name: "invalid width and height both above max", query: "width=9000&height=9000", wantErr: true},
	{name: "invalid bbox and width non-numeric", query: "bbox=1,2,3&width=abc", wantErr: true},
	{name: "invalid crs and bbox-crs", query: "crs=" + crsUnknownURI + "&bbox=" + bboxCRS84 + "&bbox-crs=foo",
		wantErr: true},

	// edge cases
	{name: "width equal to max", query: "width=5000", wantWidth: 5000, wantTransparent: true,
		knownDeviation: "D2: 20-058 /req/scaling/width-definition D, only values above the maximum are rejected"},
	{name: "height equal to max", query: "height=5000", wantHeight: 5000, wantTransparent: true,
		knownDeviation: "D2: 20-058 /req/scaling/height-definition D, only values above the maximum are rejected"},
	{name: "width one", query: "width=1", wantWidth: 1, wantTransparent: true},
	{name: "width max minus one", query: "width=4999", wantWidth: 4999, wantTransparent: true},
	{name: "empty width", query: "width=", wantTransparent: true},
	{name: "empty height", query: "height=", wantTransparent: true},
	{name: "empty bbox", query: "bbox=", wantTransparent: true},
	{name: "empty crs", query: "crs=", wantCRS: rd28992URI, wantTransparent: true},
	{name: "empty bbox-crs", query: "bbox-crs=", wantTransparent: true},
	{name: "repeated width uses first value", query: "width=100&width=200", wantWidth: 100, wantTransparent: true},
	{name: "valid crs with invalid bbox", query: "crs=" + rd28992URI + "&bbox=1,2", wantErr: true},
	{name: "valid bbox with invalid width", query: "bbox=" + bboxCRS84 + "&width=x", wantErr: true},
}

func TestMapParamsParse(t *testing.T) {
	supportedCrs, err := getSupportedCrs(config.MapsCollection{SupportedSrs: testSupportedSrs})
	require.NoError(t, err)
	seen := make(map[string]struct{}, len(paramCases))

	for _, tc := range paramCases {
		_, dup := seen[tc.name]
		require.False(t, dup, "duplicate case name: %s", tc.name)
		seen[tc.name] = struct{}{}

		t.Run(tc.name, func(t *testing.T) {
			params, err := url.ParseQuery(tc.query)
			require.NoError(t, err)
			m := mapsURL{params: params}

			request, err := m.parse(rd28992URI, supportedCrs)

			if tc.wantErr {
				require.Error(t, err, tc.knownDeviation)

				return
			}

			require.NoError(t, err, tc.knownDeviation)
			if tc.wantCRS != "" {
				assert.Equal(t, tc.wantCRS, request.crs, tc.knownDeviation)
			}
			assert.Equal(t, tc.wantBboxSRID.GetOrDefault(), request.bboxSRID.GetOrDefault(), tc.knownDeviation)
			if tc.wantBbox == nil {
				assert.Nil(t, request.bbox, tc.knownDeviation)
			} else {
				require.NotNil(t, request.bbox, tc.knownDeviation)
				assert.Equal(t, tc.wantBbox, []float64{request.bbox.Min(0), request.bbox.Min(1), request.bbox.Max(0), request.bbox.Max(1)}, tc.knownDeviation)
			}
			assert.Equal(t, tc.wantWidth, request.width, tc.knownDeviation)
			assert.Equal(t, tc.wantHeight, request.height, tc.knownDeviation)
			assert.Equal(t, tc.wantTransparent, request.transparent)
		})
	}
}
