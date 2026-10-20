package maps

import (
	"net/url"
	"testing"

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

// For now only the default CRS of the collection is supported, see getSupportedCrs.
var supportedCrs = []domain.ContentCrs{rd28992URI}

type paramCase struct {
	name           string
	query          string
	wantCRS        domain.ContentCrs
	wantBboxSRID   domain.SRID
	wantBbox       []float64
	wantWidth      int
	wantHeight     int
	wantErr        bool
	knownDeviation string
}

var paramCases = []paramCase{
	// all single parameters
	{name: "none", query: "", wantCRS: rd28992URI},
	{name: "width only", query: "width=800", wantWidth: 800},
	{name: "height only", query: "height=600", wantHeight: 600},
	{name: "crs only", query: "crs=" + rd28992URI, wantCRS: rd28992URI},
	{name: "bbox only", query: "bbox=" + bboxCRS84, wantBboxSRID: domain.WGS84SRID, wantBbox: []float64{4, 52, 5, 53}},
	{name: "bbox-crs only without bbox is ignored", query: "bbox-crs=" + domain.WGS84CrsURI},
	{name: "invalid bbox-crs without bbox is ignored", query: "bbox-crs=foo",
		knownDeviation: "D7: 20-058 /req/spatial-subsetting/bbox-crs F, bbox-crs is ignored when bbox is absent"},

	// combinations
	{name: "unsupported bbox + bbox-crs CRS84", query: "bbox=" + bboxCRS84 + "&bbox-crs=" + domain.WGS84CrsURI, wantErr: true},
	{name: "unsupported bbox + bbox-crs EPSG:4326", query: "bbox=" + bboxLatLon + "&bbox-crs=" + crsEPSG4326URI, wantErr: true},
	{name: "unsupported crs CRS84", query: "crs=" + domain.WGS84CrsURI, wantErr: true},
	{name: "bbox + bbox-crs EPSG:28992", query: "bbox=" + bboxRD + "&bbox-crs=" + rd28992URI,
		wantBboxSRID: 28992, wantBbox: []float64{100000, 400000, 150000, 450000}},
	{name: "crs + bbox", query: "crs=" + rd28992URI + "&bbox=" + bboxCRS84,
		wantCRS: rd28992URI, wantBboxSRID: domain.WGS84SRID, wantBbox: []float64{4, 52, 5, 53}},
	{name: "crs + bbox + bbox-crs", query: "crs=" + rd28992URI + "&bbox=" + bboxRD + "&bbox-crs=" + rd28992URI,
		wantCRS: rd28992URI, wantBboxSRID: 28992, wantBbox: []float64{100000, 400000, 150000, 450000}},
	{name: "width + height", query: "width=800&height=600", wantWidth: 800, wantHeight: 600},
	{name: "all five", query: "crs=" + rd28992URI + "&bbox=" + bboxRD + "&bbox-crs=" + rd28992URI + "&width=800&height=600",
		wantCRS: rd28992URI, wantBboxSRID: 28992, wantBbox: []float64{100000, 400000, 150000, 450000},
		wantWidth: 800, wantHeight: 600},

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
	{name: "width equal to max", query: "width=5000", wantWidth: 5000,
		knownDeviation: "D2: 20-058 /req/scaling/width-definition D, only values above the maximum are rejected"},
	{name: "height equal to max", query: "height=5000", wantHeight: 5000,
		knownDeviation: "D2: 20-058 /req/scaling/height-definition D, only values above the maximum are rejected"},
	{name: "width one", query: "width=1", wantWidth: 1},
	{name: "width max minus one", query: "width=4999", wantWidth: 4999},
	{name: "empty width", query: "width="},
	{name: "empty height", query: "height="},
	{name: "empty bbox", query: "bbox="},
	{name: "empty crs", query: "crs=", wantCRS: rd28992URI},
	{name: "empty bbox-crs", query: "bbox-crs="},
	{name: "repeated width uses first value", query: "width=100&width=200", wantWidth: 100},
	{name: "valid crs with invalid bbox", query: "crs=" + rd28992URI + "&bbox=1,2", wantErr: true},
	{name: "valid bbox with invalid width", query: "bbox=" + bboxCRS84 + "&width=x", wantErr: true},
}

func TestMapParamsParse(t *testing.T) {
	seen := make(map[string]struct{}, len(paramCases))

	for _, tc := range paramCases {
		_, dup := seen[tc.name]
		require.False(t, dup, "duplicate case name: %s", tc.name)
		seen[tc.name] = struct{}{}

		t.Run(tc.name, func(t *testing.T) {
			params, err := url.ParseQuery(tc.query)
			require.NoError(t, err)
			m := mapsURL{params: params}

			crs, bboxSRID, bbox, width, height, err := m.parse(rd28992URI, supportedCrs)

			if tc.wantErr {
				require.Error(t, err, tc.knownDeviation)

				return
			}

			require.NoError(t, err, tc.knownDeviation)
			if tc.wantCRS != "" {
				assert.Equal(t, tc.wantCRS, crs, tc.knownDeviation)
			}
			assert.Equal(t, tc.wantBboxSRID.GetOrDefault(), bboxSRID.GetOrDefault(), tc.knownDeviation)
			if tc.wantBbox == nil {
				assert.Nil(t, bbox, tc.knownDeviation)
			} else {
				require.NotNil(t, bbox, tc.knownDeviation)
				assert.Equal(t, tc.wantBbox, []float64{bbox.Min(0), bbox.Min(1), bbox.Max(0), bbox.Max(1)}, tc.knownDeviation)
			}
			assert.Equal(t, tc.wantWidth, width, tc.knownDeviation)
			assert.Equal(t, tc.wantHeight, height, tc.knownDeviation)
		})
	}
}
