package maps

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/engine/types"
	"github.com/PDOK/gokoala/internal/ogc/common/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

const (
	// Conformance class URIs, see https://docs.ogc.org/is/20-058/20-058.html Annex A
	confClassMapsCore = "https://www.opengis.net/spec/ogcapi-maps-1/1.0/req/core"

	crs84URI       = "https://www.opengis.net/def/crs/OGC/1.3/CRS84"
	rd28992URI     = "http://www.opengis.net/def/crs/EPSG/0/28992"
	testCollection = "foo"
	testStyle      = "default"
)

// RFC 3339 section 5.6 time instant, extended with the reduced precision
// formats allowed by OGC API Maps: yyyy, yyyy-mm, yyyy-mm-dd, yyyy-mm-ddThhZ
// (or offset) and yyyy-mm-ddThh:mmZ (or offset).
var (
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
	router := newTestRouter(t)

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
	router := newTestRouter(t)
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

	t.Run("Content-Crs header may be omitted for CRS84", func(t *testing.T) {
		rr := doGet(router, mapPath+"?crs="+url.QueryEscape(crs84URI), "image/png")
		require.Equal(t, http.StatusOK, rr.Code)

		got := rr.Header().Get("Content-Crs")
		if got != "" {
			assert.Equal(t, "<"+crs84URI+">", got)
		}
	})

	t.Run("Content-Bbox header holds lower-left and upper-right corner in response CRS", func(t *testing.T) {
		rr := doGet(router, mapPath, "image/png")
		require.Equal(t, http.StatusOK, rr.Code)

		bbox := parseBbox(t, rr.Header().Get("Content-Bbox"))

		// Storage CRS is EPSG:28992 (x/y order): lower-left must be smaller than upper-right
		assert.Less(t, bbox[0], bbox[2], "minx must be smaller than maxx")
		assert.Less(t, bbox[1], bbox[3], "miny must be smaller than maxy")
		// and has to fall in the valid extent of EPSG:28992
		assert.GreaterOrEqual(t, bbox[0], -7000.0)
		assert.LessOrEqual(t, bbox[2], 300000.0)
		assert.GreaterOrEqual(t, bbox[1], 289000.0)
		assert.LessOrEqual(t, bbox[3], 629000.0)
	})

	t.Run("Content-Bbox is in CRS84 (lon/lat order) when CRS84 is requested", func(t *testing.T) {
		rr := doGet(router, mapPath+"?crs="+url.QueryEscape(crs84URI), "image/png")
		require.Equal(t, http.StatusOK, rr.Code)

		bbox := parseBbox(t, rr.Header().Get("Content-Bbox"))

		assert.Less(t, bbox[0], bbox[2])
		assert.Less(t, bbox[1], bbox[3])
		assert.GreaterOrEqual(t, bbox[0], -180.0)
		assert.LessOrEqual(t, bbox[2], 180.0)
		assert.GreaterOrEqual(t, bbox[1], -90.0)
		assert.LessOrEqual(t, bbox[3], 90.0)
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

// Abstract test A.3: /conf/core/conformance-success
//
// Given: a conformance resource in a recognized format, such as the OGC API JSON /conformance resource
// When:  retrieving that resource from the API endpoint
// Then:  assert that the list of conformance classes includes at minimum
// https://www.opengis.net/spec/ogcapi-maps-1/1.0/req/core.
func TestCoreConformanceSuccess(t *testing.T) {
	e := newTestEngine(t, newMockWMS(t).URL)
	core.NewCommonCore(e, core.ExtraConformanceClasses{})
	NewMaps(e)

	rr := doGet(e.Router, "/conformance?f=json", "application/json")

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), confClassMapsCore)
}

// newMockWMS starts a fake WMS server that always responds with a small PNG, regardless of the request.
func newMockWMS(t *testing.T) *httptest.Server {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for x := range 4 {
		for y := range 4 {
			img.Set(x, y, color.RGBA{R: 0x1f, G: 0x77, B: 0xb4, A: 0xff})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
		Resources:          &config.Resources{Directory: types.PtrTo("/fakedirectory")},
		AvailableLanguages: []config.Language{{Tag: language.Dutch}},
		BaseURL:            config.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
		OgcAPI: config.OgcAPI{
			Maps: &config.OgcAPIMaps{
				WMSServer: &config.URL{URL: wms},
				DefaultMap: config.DefaultMap{
					CRS:    "EPSG:28992",
					Width:  800,
					Height: 600,
					Bbox:   []string{"10000", "300000", "280000", "620000"},
				},
				Collections: []config.MapsCollection{
					{
						ID:         testCollection,
						DefaultCRS: "EPSG:28992",
						Bbox:       []string{"10000", "300000", "280000", "620000"},
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

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	e := newTestEngine(t, newMockWMS(t).URL)
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

// parseBbox parses a Content-Bbox header: four comma-separated numbers (minx,miny,maxx,maxy).
func parseBbox(t *testing.T, header string) [4]float64 {
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
