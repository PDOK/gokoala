package maps

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapsHTML(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		name       string
		path       string
		wantMapURL string // regex, since the test base URL ends with a slash
		wantBbox   string
	}{
		{
			name:       "map of collection",
			path:       "/collections/" + testCollection + "/map?f=html",
			wantMapURL: `map-url="https://api.foobar.example/+collections/` + testCollection + `/map"`,
			wantBbox:   `bbox="10000,300000,280000,620000"`,
		},
		{
			name:       "map of collection with style",
			path:       "/collections/" + testCollection + "/styles/" + testStyle + "/map?f=html",
			wantMapURL: `map-url="https://api.foobar.example/+collections/` + testCollection + `/styles/` + testStyle + `/map"`,
			wantBbox:   `bbox="10000,300000,280000,620000"`,
		},
		{
			name:       "map of collection with requested bbox",
			path:       "/collections/" + testCollection + "/map?f=html&bbox=150000,400000,160000,410000.5",
			wantMapURL: `map-url="https://api.foobar.example/+collections/` + testCollection + `/map"`,
			wantBbox:   `bbox="150000,400000,160000,410000.5"`,
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
		})
	}
}
