package maps

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/stretchr/testify/require"
)

func TestMapWMSResponse(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var pngBody, jpegBody bytes.Buffer
	require.NoError(t, png.Encode(&pngBody, img))
	require.NoError(t, jpeg.Encode(&jpegBody, img, nil))

	tests := []struct {
		name        string
		status      int    // returned by the WMS backend
		contentType string // returned by the WMS backend
		body        string // returned by the WMS backend
		accept      string
		wantStatus  int
	}{
		// Successful responses
		{"PNG", 200, "image/png", pngBody.String(), "image/png", 200},
		{"JPEG", 200, "image/jpeg", jpegBody.String(), "image/jpeg", 200},
		{"media type parameters", 200, "image/png; charset=binary", pngBody.String(), "image/png", 200},
		// Error responses
		{"WMS bad request", 400, "text/xml", "<ServiceException>backend details</ServiceException>", "image/png", 502},
		{"WMS server error", 500, "text/plain", "backend details", "image/png", 502},
		{"WMS XML exception with 200", 200, "text/xml", "<ServiceException>backend details</ServiceException>", "image/png", 502},
		{"HTML with 200", 200, "text/html", "backend details", "image/png", 502},
		{"wrong image format", 200, "image/jpeg", jpegBody.String(), "image/png", 502},
		{"missing media type", 200, "", "backend details", "image/png", 502},
		{"malformed media type", 200, "image/png; invalid", "backend details", "image/png", 502},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wms := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header()["Content-Type"] = []string{tt.contentType}
				w.Header().Set(engine.HeaderContentCrs, "backend CRS")
				w.Header().Set(engine.HeaderContentBbox, "backend bbox")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(wms.Close)
			router := newTestRouter(t, wms.URL, nil)
			rr := doGet(router, "/collections/"+testCollection+"/map", tt.accept)
			if tt.wantStatus != http.StatusOK {
				requireMapProblem(t, rr, tt.wantStatus)
				require.NotContains(t, rr.Body.String(), "backend details")
				return
			}
			require.Equal(t, http.StatusOK, rr.Code)
			require.Equal(t, tt.contentType, rr.Header().Get("Content-Type"))
			require.Equal(t, "<"+rd28992URI+">", rr.Header().Get(engine.HeaderContentCrs))
			require.Equal(t, "10000.000000,300000.000000,280000.000000,620000.000000",
				rr.Header().Get(engine.HeaderContentBbox))
			decoded, _, err := image.Decode(rr.Body)
			require.NoError(t, err)
			require.Equal(t, img.Bounds(), decoded.Bounds())
		})
	}
}

func TestMapWMSConnectionFailure(t *testing.T) {
	wms := newMockWMS(t)
	wms.Close()
	router := newTestRouter(t, wms.URL, nil)
	rr := doGet(router, "/collections/"+testCollection+"/map", "image/png")
	requireMapProblem(t, rr, http.StatusBadGateway)
}

func TestMapLocalErrors(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		configure  func(*config.OgcAPIMaps)
		wantStatus int
	}{
		{
			name: "invalid configured default CRS",
			configure: func(c *config.OgcAPIMaps) {
				c.DefaultMap.CRS = "invalid backend configuration"
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "invalid configured supported CRS",
			configure: func(c *config.OgcAPIMaps) {
				c.Collections[0].SupportedSrs = []config.MapsSupportedSrs{
					{Srs: "invalid backend configuration"},
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "invalid configured default bbox",
			configure: func(c *config.OgcAPIMaps) {
				c.DefaultMap.Bbox[0] = "invalid backend configuration"
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "inconsistent configured layer styles",
			configure: func(c *config.OgcAPIMaps) {
				c.Collections[0].Layers = append(c.Collections[0].Layers,
					config.MapLayer{ID: "other", Name: "other"})
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid client dimensions",
			query:      "?width=abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid client CRS",
			query:      "?crs=" + crsUnknownURI,
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := make(chan struct{}, 1)
			wms := newMockWMS(t, func(_ *http.Request) {
				called <- struct{}{}
			})
			router := newTestRouter(t, wms.URL, tt.configure)
			rr := doGet(router, "/collections/"+testCollection+"/styles/"+testStyle+"/map"+tt.query, "image/png")
			requireMapProblem(t, rr, tt.wantStatus)
			require.NotContains(t, rr.Body.String(), "invalid backend configuration")
			select {
			case <-called:
				t.Fatal("WMS backend was called for a local error")
			default:
			}
		})
	}
}

func requireMapProblem(t *testing.T, rr *httptest.ResponseRecorder, status int) {
	t.Helper()
	require.Equal(t, status, rr.Code)
	require.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))
	require.Empty(t, rr.Header().Get(engine.HeaderContentCrs))
	require.Empty(t, rr.Header().Get(engine.HeaderContentBbox))
	var problem struct {
		Status int `json:"status"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &problem))
	require.Equal(t, status, problem.Status)
}
