package maps

import (
	"net/http"
	"regexp"
	"strconv"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapsHTML(t *testing.T) {
	router := newTestRouter(t, newMockWMS(t).URL, func(c *config.OgcAPIMaps) {
		c.Collections[0].Bbox = []string{"20000", "310000", "270000", "610000"}
		c.Basemap = "BRT"
	})

	tests := []struct {
		name            string
		path            string
		wantMapURL      string // regex, since the test base URL ends with a slash
		wantBbox        string
		wantCRS         string
		wantTransparent bool
	}{
		{
			name:            "opaque map with style",
			path:            "/collections/" + testCollection + "/styles/" + testStyle + "/map?f=html&transparent=FALSE",
			wantMapURL:      `map-url="https://api.foobar.example/+collections/` + testCollection + `/styles/` + testStyle + `/map"`,
			wantBbox:        "10000,300000,280000,620000",
			wantCRS:         rd28992URI,
			wantTransparent: false,
		},
		{
			name:            "map of collection",
			wantTransparent: true,
			path:            "/collections/" + testCollection + "/map?f=html",
			wantMapURL:      `map-url="https://api.foobar.example/+collections/` + testCollection + `/map"`,
			wantBbox:        "10000,300000,280000,620000",
			wantCRS:         rd28992URI,
		},
		{
			name:            "map of collection with style",
			wantTransparent: true,
			path:            "/collections/" + testCollection + "/styles/" + testStyle + "/map?f=html",
			wantMapURL:      `map-url="https://api.foobar.example/+collections/` + testCollection + `/styles/` + testStyle + `/map"`,
			wantBbox:        "10000,300000,280000,620000",
			wantCRS:         rd28992URI,
		},
		{
			name:            "map of collection with requested bbox",
			wantTransparent: true,
			path:            "/collections/" + testCollection + "/map?f=html&bbox=150000,400000,160000,410000.5&bbox-crs=" + rd28992URI,
			wantMapURL:      `map-url="https://api.foobar.example/+collections/` + testCollection + `/map"`,
			wantBbox:        "150000,400000,160000,410000.5",
			wantCRS:         rd28992URI,
		},
		{
			name:            "resolved CRS and bbox",
			path:            "/collections/" + testCollection + "/map?f=html&bbox=50,4,51,6&bbox-crs=" + crsEPSG4326URI + "&crs=" + crs84URI,
			wantMapURL:      `map-url="https://api.foobar.example/+collections/` + testCollection + `/map"`,
			wantBbox:        "4,50,6,51",
			wantCRS:         crs84URI,
			wantTransparent: true,
		},
		{
			name:            "implicit bbox from width and height",
			path:            "/collections/" + testCollection + "/map?f=html&width=1200&height=900",
			wantMapURL:      `map-url="https://api.foobar.example/+collections/` + testCollection + `/map"`,
			wantBbox:        "-175000,220000,465000,700000",
			wantCRS:         rd28992URI,
			wantTransparent: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := doGet(router, tt.path, "text/html")

			require.Equal(t, http.StatusOK, rr.Code)
			body := rr.Body.String()
			assert.Contains(t, body, "<app-maps-view")
			assert.Contains(t, body, `background-map="BRT"`)
			assert.Regexp(t, tt.wantMapURL, body)
			assert.Contains(t, body, `crs="`+tt.wantCRS+`"`)
			match := regexp.MustCompile(`bbox="([^"]+)"`).FindStringSubmatch(body)
			require.Len(t, match, 2)
			requireBBoxWithinTolerance(t, tt.wantBbox, match[1], 1e-6)
			assert.Contains(t, body, `transparent="`+strconv.FormatBool(tt.wantTransparent)+`"`)
		})
	}
}
