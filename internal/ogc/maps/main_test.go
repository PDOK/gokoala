package maps

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

const (
	crs84URI       = domain.WGS84CrsURI
	rd28992URI     = "http://www.opengis.net/def/crs/EPSG/0/28992"
	testCollection = "foo"
	testStyle      = "default"
)

// RFC 3339 section 5.6 time instant, extended with the reduced precision
// formats allowed by OGC API Maps: yyyy, yyyy-mm, yyyy-mm-dd, yyyy-mm-ddThhZ
// (or offset) and yyyy-mm-ddThh:mmZ (or offset).
var (
	testSupportedSrs = []config.MapsSupportedSrs{
		{Srs: "EPSG:28992"},
		{Srs: "EPSG:25831"},
		{Srs: "EPSG:25832"},
		{Srs: "EPSG:3034"},
		{Srs: "EPSG:3035"},
		{Srs: "EPSG:3857"},
		{Srs: "EPSG:4258"},
		{Srs: "EPSG:4326"},
	}
	timeInstant = `(\d{4}(-\d{2}(-\d{2}(T\d{2}(:\d{2}(:\d{2}(\.\d+)?)?)?(Z|[+-]\d{2}(:\d{2})?))?)?)?)`
	datetimeRe  = regexp.MustCompile(`^` + timeInstant + `(/` + timeInstant + `)?$`)
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

// A.1 Conformance class "Core": https://www.opengis.net/spec/ogcapi-maps-1/1.0/conf/core

// Abstract test A.1: /conf/core/map-op
//
// Given: a geospatial data resource conforming to the Maps API Standard, with an API path including …/map
// When:  performing a GET operation on the /map resource with a supported media type in the Accept header
// Then:  assert that the implementation supports retrieving map resources at one or more …/map URL.
func TestCoreMapOperation(t *testing.T) {
	router := newTestRouter(t, newMockWMS(t).URL, nil)

	tests := []struct {
		name       string
		path       string
		accept     string
		wantStatus int
	}{
		{
			name:       "map of collection, Accept image/png",
			path:       "/collections/" + testCollection + "/map",
			accept:     "image/png",
			wantStatus: http.StatusOK,
		},
		{
			name:       "map of collection, Accept with multiple supported types",
			path:       "/collections/" + testCollection + "/map",
			accept:     "image/png,image/jpeg",
			wantStatus: http.StatusOK,
		},
		{
			name:       "map of collection in a specific style, Accept image/png",
			path:       "/collections/" + testCollection + "/styles/" + testStyle + "/map",
			accept:     "image/png",
			wantStatus: http.StatusOK,
		},
		{
			name:       "map of unknown collection",
			path:       "/collections/does-not-exist/map",
			accept:     "image/png",
			wantStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := doGet(router, tt.path, tt.accept)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

// Abstract test A.2: /conf/core/map-response
//
// Given: a map resource that was successfully retrieved for Abstract test A.1
// When:  retrieving that resource for the map operation
// Then:  assert the status code, CRS, Content-Crs, Content-Bbox, Content-Datetime headers and body, see sub-tests.
func TestCoreMapResponse(t *testing.T) {
	router := newTestRouter(t, newMockWMS(t).URL, nil)
	mapPath := "/collections/" + testCollection + "/map"

	t.Run("response has HTTP status code 200", func(t *testing.T) {
		rr := doGet(router, mapPath, "image/png")

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("map is in the storage CRS and Content-Crs header says so", func(t *testing.T) {
		// no CRS requested, so the storage CRS of the collection (EPSG:28992) is used.
		rr := doGet(router, mapPath, "image/png")
		require.Equal(t, http.StatusOK, rr.Code)

		// Content-Crs is mandatory since the storage CRS isn't CRS84.
		assert.Equal(t, "<"+rd28992URI+">", rr.Header().Get("Content-Crs"))
	})

	t.Run("CRS84 is supported", func(t *testing.T) {
		captured := make(chan url.Values, 1)
		wms := newMockWMS(t, func(r *http.Request) {
			captured <- r.URL.Query()
		})
		router := newTestRouter(t, wms.URL, nil)
		rr := doGet(router, mapPath+"?crs="+url.QueryEscape(crs84URI), "image/png")
		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "<"+crs84URI+">", rr.Header().Get(engine.HeaderContentCrs))
		select {
		case params := <-captured:
			require.Equal(t, "CRS:84", params.Get("CRS"))
			require.Equal(t, params.Get("BBOX"), rr.Header().Get(engine.HeaderContentBbox))
		default:
			t.Fatal("WMS backend was not called")
		}
		bbox := parseContentBbox(t, rr.Header().Get(engine.HeaderContentBbox))
		require.Greater(t, bbox[0], 0.0)
		require.Less(t, bbox[2], 10.0)
		require.Greater(t, bbox[1], 50.0)
		require.Less(t, bbox[3], 60.0)
		_, err := png.Decode(bytes.NewReader(rr.Body.Bytes()))
		require.NoError(t, err)
	})

	t.Run("Content-Datetime header, when present, is a valid time instant or interval", func(t *testing.T) {
		// Header is only mandatory when the collection has a temporal aspect. The test collection
		// has no temporal extent, so only verify the format in case an implementation sets it anyway.
		rr := doGet(router, mapPath, "image/png")
		require.Equal(t, http.StatusOK, rr.Code)

		if got := rr.Header().Get("Content-Datetime"); got != "" {
			assertValidContentDatetime(t, got)
		}
	})

	t.Run("body contains a map in the negotiated format", func(t *testing.T) {
		rr := doGet(router, mapPath, "image/png")
		require.Equal(t, http.StatusOK, rr.Code)

		assert.Equal(t, "image/png", rr.Header().Get("Content-Type"))
		img, err := png.Decode(bytes.NewReader(rr.Body.Bytes()))
		require.NoError(t, err, "body must be a decodable PNG")
		assert.Positive(t, img.Bounds().Dx())
		assert.Positive(t, img.Bounds().Dy())
	})
	t.Run("Content-Bbox matches the WMS rendering extent", func(t *testing.T) {
		tests := []struct {
			name               string
			query              string
			omitCollectionBbox bool
			want               [4]float64
			tolerance          float64
		}{
			{
				name: "default bbox",
				want: [4]float64{10000, 300000, 280000, 620000},
			},
			{
				name:  "explicit bbox",
				query: "?bbox=131869,451318,140816,461366&bbox-crs=" + url.QueryEscape(rd28992URI),
				want:  [4]float64{131869, 451318, 140816, 461366},
			},
			{
				name:      "transformed bbox",
				query:     "?bbox=5.05,52.05,5.18,52.14",
				want:      [4]float64{131869.474, 451352.254, 140815.394, 461332.147},
				tolerance: 0.01,
			},
			{
				name:               "no collection bbox",
				omitCollectionBbox: true,
				want:               [4]float64{10000, 300000, 280000, 620000},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				capturedBbox := make(chan string, 1)
				wms := newMockWMS(t, func(r *http.Request) {
					capturedBbox <- r.URL.Query().Get("BBOX")
				})
				router := newTestRouter(t, wms.URL,
					func(mapsConfig *config.OgcAPIMaps) {
						if tt.omitCollectionBbox {
							mapsConfig.Collections[0].Bbox = nil
						}
					},
				)

				rr := doGet(router, mapPath+tt.query, "image/png")
				require.Equal(t, http.StatusOK, rr.Code)
				header := rr.Header().Get(engine.HeaderContentBbox)
				select {
				case bbox := <-capturedBbox:
					require.NotEmpty(t, bbox)
					require.Equal(t, bbox, header)
				default:
					t.Fatal("WMS backend was not called")
				}
				got := parseContentBbox(t, header)
				if tt.tolerance > 0 {
					require.InDeltaSlice(t, tt.want[:], got[:], tt.tolerance)
				} else {
					require.Equal(t, tt.want, got)
				}
			})
		}
	})
}

// Verifies the Content-Datetime syntax rules of Abstract test A.2, independent of the implementation.
func TestContentDatetimeFormat(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		// time instants
		{"2024", true},
		{"2024-05", true},
		{"2024-05-17", true},
		{"2024-05-17T10Z", true},
		{"2024-05-17T10-04:00", true},
		{"2024-05-17T10:30Z", true},
		{"2024-05-17T10:30-04:00", true},
		{"2024-05-17T10:30:15Z", true},
		{"2024-05-17T10:30:15.123+02:00", true},
		// intervals use "time-instant / time-instant"
		{"2023-01-01T00:00:00Z/2024-01-01T00:00:00Z", true},
		{"2023/2024", true},
		// invalid
		{"", false},
		{"17-05-2024", false},
		{"2024-05-17 10:30:15", false},
		{"2023-01-01T00:00:00Z,2024-01-01T00:00:00Z", false},
		{"2023-01-01T00:00:00Z/2024-01-01T00:00:00Z/2025-01-01T00:00:00Z", false},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			assert.Equal(t, tt.valid, datetimeRe.MatchString(tt.value))
		})
	}
}

// newMockWMS starts a fake WMS server that responds with a small image in the requested format.
func newMockWMS(t *testing.T, onRequest ...func(*http.Request)) *httptest.Server {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for x := range 4 {
		for y := range 4 {
			img.Set(x, y, color.RGBA{R: 0x1f, G: 0x77, B: 0xb4, A: 0xff})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	var jpegBuf bytes.Buffer
	require.NoError(t, jpeg.Encode(&jpegBuf, img, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, callback := range onRequest {
			if callback != nil {
				callback(r)
			}
			if r.URL.Query().Get("FORMAT") == "image/jpeg" {
				w.Header().Set("Content-Type", "image/jpeg")
				_, _ = w.Write(jpegBuf.Bytes())
				return
			}
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(buf.Bytes())
	}))
	t.Cleanup(ts.Close)

	return ts
}

func newTestEngine(t *testing.T, wmsURL string) *engine.Engine {
	t.Helper()

	theme, err := config.NewTheme("internal/engine/testdata/test_theme.yaml")
	require.NoError(t, err)

	wms, err := url.Parse(wmsURL)
	require.NoError(t, err)

	return engine.NewEngineWithConfig(&config.Config{
		Version:            "0.4.0",
		Title:              "Test API",
		Abstract:           "Test API description",
		Resources:          &config.Resources{Directory: new("/fakedirectory")},
		AvailableLanguages: []config.Language{{Tag: language.Dutch}},
		BaseURL:            config.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
		OgcAPI: config.OgcAPI{
			Maps: &config.OgcAPIMaps{
				WMSServer: &config.URL{URL: wms},
				MaxWidth:  4000,
				MaxHeight: 4000,
				DefaultMap: config.DefaultMap{
					CRS:        "EPSG:28992",
					Width:      506,
					Height:     600,
					Bbox:       []string{"10000", "300000", "280000", "620000"},
					MMPerPixel: "0.28",
				},
				Collections: []config.MapsCollection{
					{
						ID:           testCollection,
						SupportedSrs: testSupportedSrs,
						Bbox:         []string{"10000", "300000", "280000", "620000"},
						Layers: []config.MapLayer{
							{
								ID:     "layer1",
								Name:   "layer1",
								Styles: []config.MapStyle{{ID: testStyle, WMSStyle: "default"}},
							},
						},
					},
				},
			},
		},
	}, theme, "", false, true)
}

func newTestRouter(t *testing.T, wmsURL string, configure func(*config.OgcAPIMaps)) http.Handler {
	t.Helper()

	e := newTestEngine(t, wmsURL)
	if configure != nil {
		configure(e.Config.OgcAPI.Maps)
	}
	NewMaps(e)

	return e.Router
}

func doGet(h http.Handler, target string, accept string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080"+target, nil)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	return rr
}

// parseContentBbox parses a Content-Bbox header: four comma-separated numbers (minx,miny,maxx,maxy).
func parseContentBbox(t *testing.T, header string) [4]float64 {
	t.Helper()
	require.NotEmpty(t, header, "Content-Bbox header must be present")

	parts := strings.Split(header, ",")
	require.Len(t, parts, 4, "Content-Bbox must contain four comma-separated numbers")

	var bbox [4]float64
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		require.NoError(t, err, "Content-Bbox coordinate %q must be a number", p)
		bbox[i] = v
	}

	return bbox
}

func assertValidContentDatetime(t *testing.T, value string) {
	t.Helper()

	assert.Truef(t, datetimeRe.MatchString(value),
		"Content-Datetime %q must be an RFC 3339 time instant or an interval of the form 'instant/instant'", value)
}

type httpCase struct {
	name           string
	query          string
	wantStatus     int
	wantContentCrs string
	knownDeviation string
}

func TestMapParamsHTTP(t *testing.T) {
	router := newTestRouter(t, newMockWMS(t).URL, nil)
	mapPath := "/collections/" + testCollection + "/map"
	rd := "<" + rd28992URI + ">"

	tests := []httpCase{
		{name: "none", query: "", wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "width only", query: "width=800", wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "height only", query: "height=600", wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "crs only", query: "crs=" + rd28992URI, wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "bbox only", query: "bbox=" + bboxCRS84, wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "bbox-crs only without bbox is ignored", query: "bbox-crs=" + domain.WGS84CrsURI,
			wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "bbox + bbox-crs CRS84", query: "bbox=" + bboxCRS84 + "&bbox-crs=" + domain.WGS84CrsURI,
			wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "bbox + bbox-crs EPSG:4326", query: "bbox=" + bboxLatLon + "&bbox-crs=" + crsEPSG4326URI,
			wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "bbox + bbox-crs EPSG:28992", query: "bbox=" + bboxRD + "&bbox-crs=" + rd28992URI,
			wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "crs CRS84", query: "crs=" + domain.WGS84CrsURI, wantStatus: http.StatusOK,
			wantContentCrs: "<" + crs84URI + ">"},
		{name: "crs + bbox", query: "crs=" + rd28992URI + "&bbox=" + bboxCRS84, wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "crs + bbox + bbox-crs", query: "crs=" + rd28992URI + "&bbox=" + bboxRD + "&bbox-crs=" + rd28992URI,
			wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "width + height", query: "width=800&height=600", wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "all five", query: "crs=" + rd28992URI + "&bbox=" + bboxRD + "&bbox-crs=" + rd28992URI + "&width=800&height=600",
			wantStatus: http.StatusOK, wantContentCrs: rd},
		{name: "width equal to max", query: "crs=" + rd28992URI + "&width=4000&height=4000", wantStatus: http.StatusOK,
			wantContentCrs: rd, knownDeviation: "D2: value equal to max is accepted"},
		{name: "invalid width non-numeric", query: "width=abc", wantStatus: http.StatusBadRequest},
		{name: "invalid width zero", query: "width=0", wantStatus: http.StatusBadRequest,
			knownDeviation: "D1: 20-058 /req/scaling/width-definition C"},
		{name: "invalid width negative", query: "width=-5", wantStatus: http.StatusBadRequest,
			knownDeviation: "D1: 20-058 /req/scaling/width-definition C"},
		{name: "invalid height zero", query: "height=0", wantStatus: http.StatusBadRequest,
			knownDeviation: "D1: 20-058 /req/scaling/height-definition C"},
		{name: "invalid width above max", query: "width=4001", wantStatus: http.StatusRequestEntityTooLarge},
		{name: "invalid height above max", query: "height=4001", wantStatus: http.StatusRequestEntityTooLarge},
		{name: "invalid crs unknown", query: "crs=" + crsUnknownURI, wantStatus: http.StatusBadRequest},
		{name: "invalid bbox 3 values", query: "bbox=1,2,3", wantStatus: http.StatusBadRequest},
		{name: "invalid bbox non-numeric", query: "bbox=a,b,c,d", wantStatus: http.StatusBadRequest},
		{name: "invalid bbox-crs not a URI with bbox", query: "bbox=" + bboxCRS84 + "&bbox-crs=foo",
			wantStatus: http.StatusBadRequest},
		{name: "invalid bbox + valid crs and size", query: "crs=" + rd28992URI + "&bbox=1,2,3&width=800&height=600",
			wantStatus: http.StatusBadRequest},
		{name: "invalid bbox and width non-numeric", query: "bbox=1,2,3&width=abc", wantStatus: http.StatusBadRequest},
		{name: "invalid width and height both above max", query: "width=9000&height=9000", wantStatus: http.StatusRequestEntityTooLarge},
	}

	seen := make(map[string]struct{}, len(tests))
	for _, tt := range tests {
		_, dup := seen[tt.name]
		require.False(t, dup, "duplicate case name: %s", tt.name)
		seen[tt.name] = struct{}{}

		t.Run(tt.name, func(t *testing.T) {
			rr := doGet(router, mapPath+"?"+tt.query, "image/png")

			require.Equal(t, tt.wantStatus, rr.Code, tt.knownDeviation)
			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, tt.wantContentCrs, rr.Header().Get("Content-Crs"), tt.knownDeviation)

				return
			}
		})
	}

}

func TestMapLimitsHTTP(t *testing.T) {
	for _, mapPath := range []string{
		"/collections/" + testCollection + "/map",
		"/collections/" + testCollection + "/styles/" + testStyle + "/map",
	} {
		for _, format := range []string{"png", "jpeg", "html"} {
			for _, tt := range []struct {
				name                string
				query               string
				maxWidth, maxHeight int
				wantStatus          int
				wantDetail          string
			}{
				{name: "exact custom limits above old ceiling", query: "width=6000&height=3000",
					maxWidth: 6000, maxHeight: 3000, wantStatus: http.StatusOK},
				{name: "custom width exceeded", query: "width=6001", maxWidth: 6000, maxHeight: 3000,
					wantStatus: http.StatusRequestEntityTooLarge, wantDetail: "maxWidth (6000)"},
				{name: "custom height exceeded", query: "height=3001", maxWidth: 6000, maxHeight: 3000,
					wantStatus: http.StatusRequestEntityTooLarge, wantDetail: "maxHeight (3000)"},
				{name: "pixel count exceeded", query: "width=6001&height=3000", maxWidth: 6000, maxHeight: 3000,
					wantStatus: http.StatusRequestEntityTooLarge, wantDetail: "maxPixels (18000000)"},
				{name: "default width checked", maxWidth: 505, maxHeight: 3000,
					wantStatus: http.StatusRequestEntityTooLarge, wantDetail: "maxWidth (505)"},
				{name: "default height checked", maxWidth: 6000, maxHeight: 599,
					wantStatus: http.StatusRequestEntityTooLarge, wantDetail: "maxHeight (599)"},
				{name: "malformed dimension remains bad request", query: "width=abc", maxWidth: 6000, maxHeight: 3000,
					wantStatus: http.StatusBadRequest, wantDetail: "not a valid integer"},
				{name: "non-size validation remains bad request", query: "mm-per-pixel=-1", maxWidth: 6000, maxHeight: 3000,
					wantStatus: http.StatusBadRequest, wantDetail: "not a finite positive float"},
				{name: "calculated height exceeds limit", query: "width=3000", maxWidth: 6000, maxHeight: 3000,
					wantStatus: http.StatusRequestEntityTooLarge, wantDetail: "maxHeight (3000)"},
				{name: "calculated dimension overflow", query: "bbox=0,0,1,2&bbox-crs=" + rd28992URI + "&width=" + strconv.Itoa(math.MaxInt),
					maxWidth: 6000, maxHeight: 3000, wantStatus: http.StatusRequestEntityTooLarge,
					wantDetail: "calculated image dimension"},
				{name: "pixel count multiplication overflow", query: "width=" + strconv.Itoa(math.MaxInt) + "&height=" + strconv.Itoa(math.MaxInt),
					maxWidth: 6000, maxHeight: 3000, wantStatus: http.StatusRequestEntityTooLarge,
					wantDetail: "maxPixels (18000000)"},
				{name: "invalid ratio extent", query: "bbox=0,0,0,0&bbox-crs=" + rd28992URI,
					maxWidth: 6000, maxHeight: 3000, wantStatus: http.StatusBadRequest, wantDetail: "no surface area"},
			} {
				t.Run(mapPath+"/"+format+"/"+tt.name, func(t *testing.T) {
					var calls atomic.Int32
					wms := newMockWMS(t, func(*http.Request) { calls.Add(1) })
					router := newTestRouter(t, wms.URL, func(m *config.OgcAPIMaps) {
						m.MaxWidth, m.MaxHeight = tt.maxWidth, tt.maxHeight
					})
					rr := doGet(router, mapPath+"?f="+format+"&"+tt.query, "")
					require.Equal(t, tt.wantStatus, rr.Code, rr.Body.String())
					if tt.wantStatus != http.StatusOK {
						require.Contains(t, rr.Header().Get("Content-Type"), "application/problem+json")
						require.Contains(t, rr.Body.String(), tt.wantDetail)
						require.Zero(t, calls.Load(), "invalid maps must not reach WMS")
					}
				})
			}
		}
	}
}

func TestMapViewportHTTP(t *testing.T) {
	for _, mapPath := range []string{
		"/collections/" + testCollection + "/map",
		"/collections/" + testCollection + "/styles/" + testStyle + "/map",
	} {
		for _, format := range []string{"png", "jpeg"} {
			for _, tt := range []struct {
				name          string
				query         string
				width, height int
				bbox          string
				crs           string
				defaults      *config.DefaultMap
			}{
				{name: "defaults", width: 506, height: 600, bbox: "10000,300000,280000,620000"},
				{name: "width", query: "width=600", width: 600, height: 711, bbox: "10000,300000,280000,620000"},
				{name: "height", query: "height=700", width: 591, height: 700, bbox: "10000,300000,280000,620000"},
				{name: "width and height", query: "width=1200&height=900", width: 1200, height: 900,
					bbox: "-175000,220000,465000,700000"},
				{name: "bbox only", query: "bbox=0,0,2,1&bbox-crs=" + rd28992URI,
					width: 1200, height: 600, bbox: "0,0,2,1"},
				{name: "bbox and width", query: "bbox=0,0,2,1&bbox-crs=" + rd28992URI + "&width=800",
					width: 800, height: 400, bbox: "0,0,2,1"},
				{name: "bbox and height", query: "bbox=0,0,2,1&bbox-crs=" + rd28992URI + "&height=300",
					width: 600, height: 300, bbox: "0,0,2,1"},
				{name: "explicit distortion", query: "bbox=0,0,2,1&bbox-crs=" + rd28992URI + "&width=100&height=100",
					width: 100, height: 100, bbox: "0,0,2,1"},
				{name: "native EPSG4326 axes", query: "crs=" + crsEPSG4326URI + "&bbox=50,4,51,6&bbox-crs=" + crsEPSG4326URI + "&width=600",
					width: 600, height: 300, bbox: "50,4,51,6", crs: crsEPSG4326URI},
				{name: "CRS84 to EPSG4326", query: "crs=" + crsEPSG4326URI + "&bbox=4,50,6,51&width=600",
					width: 600, height: 300, bbox: "50,4,51,6", crs: crsEPSG4326URI},
				{name: "default bbox to CRS84", query: "crs=" + crs84URI + "&width=600",
					width: 600, height: 300, bbox: "4,50,6,51", crs: crs84URI,
					defaults: &config.DefaultMap{CRS: "EPSG:4326", Width: 800, Height: 400,
						Bbox: []string{"50", "4", "51", "6"}, MMPerPixel: "0.28"}},
				{name: "projected bbox to geographic", query: "crs=" + crs84URI + "&bbox=131869.474,451352.254,140815.394,461332.147&bbox-crs=" + rd28992URI + "&width=600",
					width: 600, height: 415, bbox: "5.05,52.05,5.18,52.14", crs: crs84URI},
				{name: "implicit bbox in EPSG4326", query: "crs=" + crsEPSG4326URI + "&width=600&height=600",
					width: 600, height: 600, bbox: "49.75,4.25,51.25,5.75", crs: crsEPSG4326URI,
					defaults: &config.DefaultMap{CRS: "EPSG:4326", Width: 800, Height: 400,
						Bbox: []string{"50", "4", "51", "6"}, MMPerPixel: "0.28"}},
			} {
				t.Run(mapPath+"/"+format+"/"+tt.name, func(t *testing.T) {
					var params url.Values
					wms := newMockWMS(t, func(r *http.Request) { params = r.URL.Query() })
					router := newTestRouter(t, wms.URL, func(m *config.OgcAPIMaps) {
						if tt.defaults != nil {
							m.DefaultMap = *tt.defaults
						}
					})
					rr := doGet(router, mapPath+"?f="+format+"&mm-per-pixel=0.254&"+tt.query, "")
					require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
					require.Equal(t, strconv.Itoa(tt.width), params.Get("WIDTH"))
					require.Equal(t, strconv.Itoa(tt.height), params.Get("HEIGHT"))
					requireBBoxWithinTolerance(t, tt.bbox, params.Get("BBOX"), 1e-6)
					require.Equal(t, params.Get("BBOX"), rr.Header().Get("Content-Bbox"))
					crs := tt.crs
					if crs == "" {
						crs = rd28992URI
					}
					require.Equal(t, "<"+crs+">", rr.Header().Get("Content-Crs"))
					require.Equal(t, "100.000000", params.Get("map_resolution"))
				})
			}
		}
	}
}

func TestMapTransparencyHTTP(t *testing.T) {
	for _, mapPath := range []string{
		"/collections/" + testCollection + "/map",
		"/collections/" + testCollection + "/styles/" + testStyle + "/map",
	} {
		for _, tc := range []struct {
			query   string
			want    bool
			invalid bool
		}{
			{query: "", want: true},
			{query: "transparent=true", want: true},
			{query: "transparent=false", want: false},
			{query: "transparent=true&transparent=false", want: true},
			{query: "transparent=", invalid: true},
			{query: "transparent=yes", invalid: true},
			{query: "transparent=1", invalid: true},
			{query: "transparent=TRUE", want: true},
			{query: "transparent=FALSE", want: false},
			{query: "transparent=False", want: false},
		} {
			t.Run(mapPath+"?"+tc.query, func(t *testing.T) {
				var received url.Values
				wms := newMockWMS(t, func(r *http.Request) {
					received = r.URL.Query()
				})
				router := newTestRouter(t, wms.URL, nil)
				rr := doGet(router, mapPath+"?"+tc.query, "image/png")
				if tc.invalid {
					require.Equal(t, http.StatusBadRequest, rr.Code)
					assert.Contains(t, rr.Body.String(), "transparent must be true or false")
					assert.Nil(t, received, "invalid input must not reach WMS")
					return
				}
				require.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, strconv.FormatBool(tc.want), received.Get("TRANSPARENT"))
			})
		}
	}
}
