package maps

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapsHTML(t *testing.T) {
	router := newTestRouter(t, newMockWMS(t).URL, func(c *config.OgcAPIMaps) {
		c.Collections[0].Bbox = []string{"20000", "310000", "270000", "610000"}
	})

	tests := []struct {
		name            string
		path            string
		wantMapURL      string // regex, since the test base URL ends with a slash
		wantBbox        string
		wantTransparent bool
	}{
		{
			name:            "opaque map with style",
			path:            "/collections/" + testCollection + "/styles/" + testStyle + "/map?f=html&transparent=FALSE",
			wantMapURL:      `map-url="https://api.foobar.example/+collections/` + testCollection + `/styles/` + testStyle + `/map"`,
			wantBbox:        `bbox="20000,310000,270000,610000"`,
			wantTransparent: false,
		},
		{
			name:            "map of collection",
			wantTransparent: true,
			path:            "/collections/" + testCollection + "/map?f=html",
			wantMapURL:      `map-url="https://api.foobar.example/+collections/` + testCollection + `/map"`,
			wantBbox:        `bbox="20000,310000,270000,610000"`,
		},
		{
			name:            "map of collection with style",
			wantTransparent: true,
			path:            "/collections/" + testCollection + "/styles/" + testStyle + "/map?f=html",
			wantMapURL:      `map-url="https://api.foobar.example/+collections/` + testCollection + `/styles/` + testStyle + `/map"`,
			wantBbox:        `bbox="20000,310000,270000,610000"`,
		},
		{
			name:            "map of collection with requested bbox",
			wantTransparent: true,
			path:            "/collections/" + testCollection + "/map?f=html&bbox=150000,400000,160000,410000.5",
			wantMapURL:      `map-url="https://api.foobar.example/+collections/` + testCollection + `/map"`,
			wantBbox:        `bbox="150000,400000,160000,410000.5"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := doGet(router, tt.path, "text/html")

			require.Equal(t, http.StatusOK, rr.Code)
			body := rr.Body.String()
			assert.Contains(t, body, "<app-maps-view")
			assert.Regexp(t, tt.wantMapURL, body)
			assert.Contains(t, body, `crs="`+rd28992URI+`"`)
			assert.Contains(t, body, tt.wantBbox)
			assert.Contains(t, body, `transparent="`+strconv.FormatBool(tt.wantTransparent)+`"`)
		})
	}
}
