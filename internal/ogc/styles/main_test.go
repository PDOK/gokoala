package styles

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/types"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/PDOK/gokoala/internal/engine"
	"golang.org/x/text/language"

	"github.com/stretchr/testify/assert"
)

var tilePath = "/foo/12/34/56.pbf"

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestNewStyles(t *testing.T) {
	// given
	theme, err := config.NewTheme("internal/engine/testdata/test_theme.yaml")
	require.NoError(t, err)
	type args struct {
		e *engine.Engine
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test render templates with OGC Styles config",
			args: args{
				e: engine.NewEngineWithConfig(&config.Config{
					Version:  "0.4.0",
					Title:    "Test API",
					Abstract: "Test API description",
					Resources: &config.Resources{
						Directory: types.PtrTo("/fakedirectory"),
					},
					AvailableLanguages: []config.Language{{Tag: language.Dutch}},
					BaseURL:            config.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
					OgcAPI: config.OgcAPI{
						GeoVolumes: nil,
						Tiles: &config.OgcAPITiles{
							DatasetTiles: &config.Tiles{
								TileServer: config.URL{URL: &url.URL{Scheme: "https", Host: "tiles.foobar.example", Path: "/somedataset"}},
								Types:      []config.TilesType{config.TilesTypeVector},
								SupportedSrs: []config.SupportedSrs{
									{
										Srs: "EPSG:28992",
										ZoomLevelRange: config.ZoomLevelRange{
											Start: 12,
											End:   12,
										},
									},
								},
								HealthCheck: config.HealthCheck{Enabled: types.PtrTo(true), Srs: "EPSG:28992", TilePath: &tilePath},
							},
						},
						Styles: &config.OgcAPIStyles{
							Default: "foo",
							SupportedStyles: []config.Style{
								{
									ID:    "foo",
									Title: "bar",
								},
							},
						},
					},
				}, theme, "", false, true),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			styles := NewStyles(test.args.e)
			assert.NotEmpty(t, styles.engine.Templates.RenderedTemplates)
		})
	}
}

func TestNewStyles_MapsOnly(t *testing.T) {
	theme, err := config.NewTheme("internal/engine/testdata/test_theme.yaml")
	require.NoError(t, err)
	e := engine.NewEngineWithConfig(&config.Config{
		Version:            "0.4.0",
		Title:              "Test API",
		Abstract:           "Test API description",
		Resources:          &config.Resources{Directory: types.PtrTo("/fakedirectory")},
		AvailableLanguages: []config.Language{{Tag: language.Dutch}},
		BaseURL:            config.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
		OgcAPI:             config.OgcAPI{Maps: testMapsConfig()},
	}, theme, "", false, true)

	// when
	var styles *Styles
	require.NotPanics(t, func() { styles = NewStyles(e) })

	// then
	assert.NotEmpty(t, styles.engine.Templates.RenderedTemplates)
	for _, format := range []string{"json", "html", "md"} {
		rr := doGet(e.Router, "/styles?f="+format)
		assert.Equal(t, http.StatusOK, rr.Code, "format %s", format)
	}

	// Tiles specific routes shouldn't exist without Tiles/Styles
	for _, path := range []string{"/styles/foo", "/styles/foo/metadata", "/styles/foo/legend"} {
		rr := doGet(e.Router, path)
		assert.Equal(t, http.StatusNotFound, rr.Code, "path %s", path)
	}
}

func TestNewStyles_TilesAndMaps(t *testing.T) {
	// given: both OGC API Tiles/Styles and OGC API Maps
	e, err := engine.NewEngine("internal/ogc/styles/testdata/config_minimal_styles.yaml",
		"internal/engine/testdata/test_theme.yaml", "", false, true)
	require.NoError(t, err)
	e.Config.OgcAPI.Maps = testMapsConfig()

	// when
	require.NotPanics(t, func() { NewStyles(e) })

	// then: global styles endpoint and the Tiles specific routes are available
	for _, path := range []string{
		"/styles?f=json", "/styles?f=html", "/styles?f=md",
		"/styles/default__webmercatorquad?f=mapbox",
		"/styles/default__webmercatorquad/metadata?f=json",
	} {
		rr := doGet(e.Router, path)
		assert.Equal(t, http.StatusOK, rr.Code, "path %s", path)
	}
}

func TestStyles_JSON(t *testing.T) {
	tests := []struct {
		name         string
		withTiles    bool
		withMaps     bool
		mapsNoStyles bool // first maps collection has no styles
		wantIDs      []string
	}{
		{
			name:      "tiles only",
			withTiles: true,
			wantIDs:   []string{"default__netherlandsrdnewquad", "default__webmercatorquad"},
		},
		{
			name:     "maps only",
			withMaps: true,
			wantIDs:  []string{"min_gemeenten_geannoteerd", "min_province_standaard"},
		},
		{
			name:      "tiles and maps",
			withTiles: true,
			withMaps:  true,
			wantIDs: []string{
				"default__netherlandsrdnewquad", "default__webmercatorquad",
				"min_gemeenten_geannoteerd", "min_province_standaard",
			},
		},
		{
			name:         "tiles and maps, with a maps collection without styles",
			withTiles:    true,
			withMaps:     true,
			mapsNoStyles: true,
			wantIDs:      []string{"default__netherlandsrdnewquad", "default__webmercatorquad", "min_province_standaard"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestStylesEngine(t, tt.withTiles, tt.withMaps, tt.mapsNoStyles)
			NewStyles(e)

			rr := doGet(e.Router, "/styles?f=json")
			require.Equal(t, http.StatusOK, rr.Code)

			var body struct {
				Default string `json:"default"`
				Styles  []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
					Links []struct {
						Rel   string `json:"rel"`
						Type  string `json:"type"`
						Href  string `json:"href"`
						Title string `json:"title"`
					} `json:"links"`
				} `json:"styles"`
			}
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body), "response must be valid JSON")

			gotIDs := make([]string, 0, len(body.Styles))
			for _, s := range body.Styles {
				gotIDs = append(gotIDs, s.ID)
				assert.NotEmpty(t, s.Title)
				assert.NotEmpty(t, s.Links)
				if strings.HasPrefix(s.ID, "min_") {
					// maps style: map link in all three formats
					var types, titles []string
					for _, l := range s.Links {
						types = append(types, l.Type)
						titles = append(titles, l.Title)
					}
					assert.Equal(t, []string{"image/png", "image/jpeg", "text/html"}, types, s.ID)
					assert.Equal(t, []string{s.Title + " as PNG", s.Title + " as JPEG", s.Title + " as HTML"}, titles, s.ID)
				}
			}
			assert.Equal(t, tt.wantIDs, gotIDs)
			if tt.withTiles {
				assert.Equal(t, "default", body.Default)
			} else {
				assert.Empty(t, body.Default)
			}
		})
	}
}

func newTestStylesEngine(t *testing.T, withTiles, withMaps, mapsNoStyles bool) *engine.Engine {
	t.Helper()
	e, err := engine.NewEngine("internal/ogc/styles/testdata/config_minimal_styles.yaml",
		"internal/engine/testdata/test_theme.yaml", "", false, true)
	require.NoError(t, err)
	if !withTiles {
		e.Config.OgcAPI.Tiles = nil
		e.Config.OgcAPI.Styles = nil
	}
	if withMaps {
		e.Config.OgcAPI.Maps = testMapsConfig()
		e.Config.OgcAPI.Maps.Collections = []config.MapsCollection{
			{ID: "gemeenten", Layers: []config.MapLayer{{ID: "l", Name: "l", Styles: []config.MapStyle{
				{ID: "geannoteerd", Title: "Met labels", WMSStyle: "default"},
			}}}},
			{ID: "province", Layers: []config.MapLayer{{ID: "l", Name: "l", Styles: []config.MapStyle{
				{ID: "standaard", Title: "Standaard", WMSStyle: "default"},
			}}}},
		}
		if mapsNoStyles {
			e.Config.OgcAPI.Maps.Collections[0].Layers[0].Styles = nil
		}
	}

	return e
}

func TestStyles_Markdown(t *testing.T) {
	tests := []struct {
		name         string
		withTiles    bool
		withMaps     bool
		mapsNoStyles bool
		contains     []string
		notContains  []string
	}{
		{
			name:        "tiles only",
			withTiles:   true,
			contains:    []string{"Available Styles", "Test style (NetherlandsRDNewQuad)", "Mapbox style"},
			notContains: []string{"min_", "/map?f="},
		},
		{
			name:        "maps only",
			withMaps:    true,
			contains:    []string{"Available Styles", "### Met labels", "min_gemeenten_geannoteerd", "### Standaard", "min_province_standaard", "/collections/gemeenten/styles/geannoteerd/map?f=png", "/map?f=jpeg", "/map?f=html"},
			notContains: []string{"Mapbox style"},
		},
		{
			name:      "tiles and maps",
			withTiles: true,
			withMaps:  true,
			contains:  []string{"Test style (WebMercatorQuad)", "### Met labels", "min_province_standaard"},
		},
		{
			name:         "maps collection without styles",
			withMaps:     true,
			mapsNoStyles: true,
			contains:     []string{"### Standaard"},
			notContains:  []string{"Met labels", "min_gemeenten_geannoteerd"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestStylesEngine(t, tt.withTiles, tt.withMaps, tt.mapsNoStyles)
			NewStyles(e)

			rr := doGet(e.Router, "/styles?f=md")
			require.Equal(t, http.StatusOK, rr.Code)
			for _, c := range tt.contains {
				assert.Contains(t, rr.Body.String(), c)
			}
			for _, c := range tt.notContains {
				assert.NotContains(t, rr.Body.String(), c)
			}
		})
	}
}

func TestStyles_HTML(t *testing.T) {
	extraStyles := []config.Style{
		{ID: "bare", Title: "No formats"},
	}
	tests := []struct {
		name         string
		withTiles    bool
		withMaps     bool
		mapsNoStyles bool
		extraStyles  []config.Style
		wantOptions  []string // substrings, one per expected <option>, in order
		notContains  []string
		contains     []string
	}{
		{
			name:      "tiles only",
			withTiles: true,
			wantOptions: []string{
				`data-kind="mapbox" data-url="styles/default__netherlandsrdnewquad" data-metadata="styles/default__netherlandsrdnewquad/metadata" data-format="Mapbox style" data-proj="NetherlandsRDNewQuad" data-viewer="tiles">Test style (NetherlandsRDNewQuad)`,
				`data-kind="mapbox" data-url="styles/default__webmercatorquad" data-metadata="styles/default__webmercatorquad/metadata" data-format="Mapbox style" data-proj="WebMercatorQuad" data-viewer="tiles">Test style (WebMercatorQuad)`,
			},
			contains:    []string{`id="styles-viewer"`, "<app-vectortile-view"},
			notContains: []string{`data-kind="map"`},
		},
		{
			name:     "maps only",
			withMaps: true,
			wantOptions: []string{
				`data-kind="map" data-url="collections/gemeenten/styles/geannoteerd/map" data-metadata="collections/gemeenten/styles" data-format="PNG/JPEG">Met labels`,
				`data-kind="map" data-url="collections/province/styles/standaard/map" data-metadata="collections/province/styles" data-format="PNG/JPEG">Standaard`,
			},
			notContains: []string{`data-kind="mapbox"`, "data-viewer", `id="styles-viewer"`, "<app-vectortile-view"},
		},
		{
			name:      "tiles and maps, tiles first",
			withTiles: true,
			withMaps:  true,
			wantOptions: []string{
				"Test style (NetherlandsRDNewQuad)", "Test style (WebMercatorQuad)", "Met labels", "Standaard",
			},
			contains: []string{`id="styles-viewer"`},
		},
		{
			name:         "maps collection without styles",
			withMaps:     true,
			mapsNoStyles: true,
			wantOptions:  []string{`data-url="collections/province/styles/standaard/map"`},
			notContains:  []string{"Met labels"},
		},
		{
			name:        "viewer only for styles that offer a mapbox format",
			withTiles:   true,
			extraStyles: extraStyles,
			wantOptions: []string{
				`styles/default__netherlandsrdnewquad" data-metadata` + `="styles/default__netherlandsrdnewquad/metadata" data-format="Mapbox style" data-proj="NetherlandsRDNewQuad" data-viewer="tiles">`,
				`styles/default__webmercatorquad" data-metadata` + `="styles/default__webmercatorquad/metadata" data-format="Mapbox style" data-proj="WebMercatorQuad" data-viewer="tiles">`,
				`data-url="styles/bare__netherlandsrdnewquad" data-metadata="styles/bare__netherlandsrdnewquad/metadata" data-format="Mapbox style" data-proj="NetherlandsRDNewQuad">No formats`,
				`data-url="styles/bare__webmercatorquad" data-metadata="styles/bare__webmercatorquad/metadata" data-format="Mapbox style" data-proj="WebMercatorQuad">No formats`,
			},
		},
	}
	optionRegexp := regexp.MustCompile(`(?s)<option\b(.*?)</option>`)
	whitespace := regexp.MustCompile(`\s+`)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestStylesEngine(t, tt.withTiles, tt.withMaps, tt.mapsNoStyles)
			if tt.withTiles {
				e.Config.OgcAPI.Styles.SupportedStyles = append(e.Config.OgcAPI.Styles.SupportedStyles, tt.extraStyles...)
				e.Config.OgcAPI.Styles.StylesDir = "internal/ogc/styles/testdata/resources"
			}
			NewStyles(e)

			rr := doGet(e.Router, "/styles?f=html")
			require.Equal(t, http.StatusOK, rr.Code)
			body := rr.Body.String()

			var options []string
			for _, match := range optionRegexp.FindAllStringSubmatch(body, -1) {
				options = append(options, strings.TrimSpace(whitespace.ReplaceAllString(match[1], " ")))
			}
			require.Len(t, options, len(tt.wantOptions), "options: %v", options)
			for i, want := range tt.wantOptions {
				assert.Contains(t, options[i], want)
			}
			assert.NotContains(t, body, "<option value=")
			for _, c := range tt.contains {
				assert.Contains(t, body, c)
			}
			for _, c := range tt.notContains {
				assert.NotContains(t, body, c)
			}
		})
	}
}

func testMapsConfig() *config.OgcAPIMaps {
	return &config.OgcAPIMaps{
		WMSServer: &config.URL{URL: &url.URL{Scheme: "https", Host: "wms.foobar.example", Path: "/wms"}},
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
				ID:           "gemeenten",
				SupportedSrs: []config.MapsSupportedSrs{{Srs: "EPSG:28992"}},
				Bbox:         []string{"10000", "300000", "280000", "620000"},
				Layers: []config.MapLayer{
					{ID: "l1", Name: "l1", Styles: []config.MapStyle{{ID: "geannoteerd", Title: "Met labels", WMSStyle: "default"}}},
				},
			},
		},
	}
}

func doGet(h http.Handler, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080"+target, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	return rr
}

func TestStyles_Style(t *testing.T) {
	type fields struct {
		configFile string
		url        string
		style      string
	}
	type want struct {
		bodyContains []string
		statusCode   int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "styles/default__netherlandsrdnewquad",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style",
				style:      "default__netherlandsrdnewquad",
			},
			want: want{
				bodyContains: []string{"\"id\": \"default\"", "tiles/NetherlandsRDNewQuad/{z}/{y}/{x}?f=mvt"},
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "styles/default__webmercatorquad",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style",
				style:      "default__webmercatorquad",
			},
			want: want{
				bodyContains: []string{"\"id\": \"default\"", "tiles/WebMercatorQuad/{z}/{y}/{x}?f=mvt"},
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "styles/default (backwards comp)",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style",
				style:      "default",
			},
			want: want{
				bodyContains: []string{"\"id\": \"default\"", "tiles/NetherlandsRDNewQuad/{z}/{y}/{x}?f=mvt"},
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "styles/default__netherlandsrdnewquad as Markdown",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style?f=md",
				style:      "default__netherlandsrdnewquad",
			},
			want: want{
				bodyContains: []string{"**Style URL:** [http://localhost:8080/styles/default?f=mapbox]"},
				statusCode:   http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createStyleRequest(tt.fields.url, tt.fields.style)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "internal/engine/testdata/test_theme.yaml", "", false, true)
			require.NoError(t, err)
			styles := NewStyles(newEngine)
			handler := styles.Style()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			for _, c := range tt.want.bodyContains {
				assert.Contains(t, rr.Body.String(), c)
			}
		})
	}
}

func TestStyles_Metadata(t *testing.T) {
	type fields struct {
		configFile string
		url        string
		style      string
	}
	type want struct {
		bodyContains []string
		statusCode   int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "styles/default__netherlandsrdnewquad",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style/metadata",
				style:      "default__netherlandsrdnewquad",
			},
			want: want{
				bodyContains: []string{"\"id\": \"default\"", "\"title\": \"Mapbox Style\"", "default__netherlandsrdnewquad"},
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "styles/default__webmercatorquad",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style/metadata",
				style:      "default__webmercatorquad",
			},
			want: want{
				bodyContains: []string{"\"id\": \"default\"", "\"title\": \"Mapbox Style\"", "default__webmercatorquad"},
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "styles/default (backwards comp)",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style/metadata",
				style:      "default",
			},
			want: want{
				bodyContains: []string{"\"id\": \"default\"", "\"title\": \"Mapbox Style\"", "default__netherlandsrdnewquad"},
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "styles/default__netherlandsrdnewquad/metadata as Markdown",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles/:style/metadata?f=md",
				style:      "default__netherlandsrdnewquad",
			},
			want: want{
				bodyContains: []string{"Test style Metadata"},
				statusCode:   http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createStyleRequest(tt.fields.url, tt.fields.style)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "internal/engine/testdata/test_theme.yaml", "", false, true)
			require.NoError(t, err)
			styles := NewStyles(newEngine)
			handler := styles.Metadata()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			for _, c := range tt.want.bodyContains {
				assert.Contains(t, rr.Body.String(), c)
			}
		})
	}
}

func TestStyles_Legend(t *testing.T) {
	type fields struct {
		configFile string
		url        string
		style      string
	}
	type want struct {
		bodyContains []string
		statusCode   int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "with legend",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_legend.yaml",
				url:        "http://localhost:8080/styles/default__netherlandsrdnewquad/legend",
				style:      "default__netherlandsrdnewquad",
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "without legend",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_legend.yaml",
				url:        "http://localhost:8080/styles/alternative__webmercatorquad/legend",
				style:      "alternative__webmercatorquad",
			},
			want: want{
				statusCode:   http.StatusNotFound,
				bodyContains: []string{"no legend configured for style alternative"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createStyleRequest(tt.fields.url, tt.fields.style)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "internal/engine/testdata/test_theme.yaml", "", false, true)
			require.NoError(t, err)
			styles := NewStyles(newEngine)
			handler := styles.Legend()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			for _, c := range tt.want.bodyContains {
				assert.Contains(t, rr.Body.String(), c)
			}
		})
	}
}

func TestTile_Styles(t *testing.T) {
	type fields struct {
		configFile string
		url        string
	}
	type want struct {
		bodyContains string
		statusCode   int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "test NetherlandsRDNewQuad present",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles",
			},
			want: want{
				bodyContains: "Test style (NetherlandsRDNewQuad)",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "test WebMercatorQuad present",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles",
			},
			want: want{
				bodyContains: "Test style (WebMercatorQuad)",
				statusCode:   http.StatusOK,
			},
		},
		{
			name: "test Available Styles heading present as Markdown",
			fields: fields{
				configFile: "internal/ogc/styles/testdata/config_minimal_styles.yaml",
				url:        "http://localhost:8080/styles?f=md",
			},
			want: want{
				bodyContains: "Available Styles",
				statusCode:   http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := createStylesRequest(tt.fields.url)
			if err != nil {
				log.Fatal(err)
			}
			rr, ts := createMockServer()
			defer ts.Close()

			newEngine, err := engine.NewEngine(tt.fields.configFile, "internal/engine/testdata/test_theme.yaml", "", false, true)
			require.NoError(t, err)
			styles := NewStyles(newEngine)
			handler := styles.Styles()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.want.bodyContains)
		})
	}
}

func createMockServer() (*httptest.ResponseRecorder, *httptest.Server) {
	rr := httptest.NewRecorder()
	l, err := net.Listen("tcp", "localhost:10090")
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engine.SafeWrite(w.Write, []byte(r.URL.String()))
	}))
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()

	return rr, ts
}

func createStyleRequest(url string, style string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("style", style)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	return req, err
}

func createStylesRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	rctx := chi.NewRouteContext()

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	return req, err
}
