package records

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func jsonValue(value any) config.JSONValue {
	encoded, _ := json.Marshal(value)
	return config.JSONValue(encoded)
}

func jsonProperties(values map[string]any) map[string]config.JSONValue {
	result := make(map[string]config.JSONValue, len(values))
	for key, value := range values {
		result[key] = jsonValue(value)
	}
	return result
}

func TestItemsEmitsSTACItemAndAssetExtensionFields(t *testing.T) {
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(filepath.Clean(filepath.Join(workingDirectory, "../../.."))))
	t.Cleanup(func() {
		assert.NoError(t, os.Chdir(workingDirectory))
	})

	serverConfig := &config.Config{
		BaseURL: config.URL{URL: &url.URL{Scheme: "https", Host: "api.example.org"}},
		OgcAPI: config.OgcAPI{
			Records: &config.OgcAPIRecords{
				Collections: config.RecordsCollections{{
					ID: "datasets",
					Records: config.Records{{
						ID:             "bgt",
						STACVersion:    "1.1.0",
						STACExtensions: []string{"https://stac-extensions.github.io/table/v1.3.0/schema.json"},
						Properties:     jsonProperties(map[string]any{"datetime": nil, "start_datetime": "2025-01-01T00:00:00Z", "end_datetime": "2025-12-31T23:59:59Z"}),
						Assets: map[string]config.RecordsAsset{
							"geoparquet": {
								Href:   "https://example.org/bgt.parquet",
								Type:   "application/vnd.apache.parquet",
								Roles:  []string{"data"},
								Fields: jsonProperties(map[string]any{"table:primary_geometry": "geometry"}),
							},
						},
					}},
				}},
			},
		},
	}
	testEngine := engine.NewEngineWithConfig(serverConfig, &config.Theme{
		Logo:  &config.ThemeLogo{},
		Color: &config.ThemeColors{Primary: "#000000", Secondary: "#000000", Link: "#000000"},
	}, "", false, true)
	recordsAPI, err := NewRecords(testEngine)
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "https://api.example.org/collections/datasets/items", nil)
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("collectionId", "datasets")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
	recorder := httptest.NewRecorder()
	recordsAPI.Items().ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	feature := response["features"].([]any)[0].(map[string]any)
	assert.Equal(t, "Feature", feature["type"])
	assert.Equal(t, "1.1.0", feature["stac_version"])
	assert.Equal(t, []any{"https://stac-extensions.github.io/table/v1.3.0/schema.json"}, feature["stac_extensions"])
	asset := feature["assets"].(map[string]any)["geoparquet"].(map[string]any)
	assert.Equal(t, "geometry", asset["table:primary_geometry"])
	assert.NotContains(t, asset, "fields")
}

func TestSelectRecordsCoreQueryParameters(t *testing.T) {
	records := config.Records{
		{ID: "road", Properties: jsonProperties(map[string]any{
			"title":       "BGT road network",
			"type":        "dataset",
			"keywords":    []any{"roads", "transport"},
			"externalIds": []any{map[string]any{"scheme": "pdok", "value": "road-1"}},
		})},
		{ID: "water", Properties: jsonProperties(map[string]any{
			"title":       "Water features",
			"type":        "dataset",
			"externalIds": []any{map[string]any{"scheme": "pdok", "value": "water-1"}},
		})},
		{ID: "service", Properties: jsonProperties(map[string]any{
			"title": "BGT road service",
			"type":  "service",
		})},
	}
	tests := []struct {
		name  string
		query url.Values
		want  []string
	}{
		{name: "ids are alternatives", query: url.Values{"ids": {"road,water"}}, want: []string{"road", "water"}},
		{name: "empty ids selects nothing", query: url.Values{"ids": {""}}, want: []string{}},
		{name: "q comma terms are alternatives and case insensitive", query: url.Values{"q": {"BGT,WATER"}}, want: []string{"road", "water", "service"}},
		{name: "q whitespace terms match in order", query: url.Values{"q": {"bgt road"}}, want: []string{"road", "service"}},
		{name: "types are alternatives", query: url.Values{"type": {"service,dataset"}}, want: []string{"road", "water", "service"}},
		{name: "external ID supports scheme and value", query: url.Values{"externalIds": {"pdok:road-1"}}, want: []string{"road"}},
		{name: "queryable equality", query: url.Values{"title": {"Water features"}}, want: []string{"water"}},
		{name: "different predicates are combined with AND", query: url.Values{"type": {"dataset"}, "q": {"water"}}, want: []string{"water"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selected, err := selectRecords(records, test.query)
			require.NoError(t, err)
			got := make([]string, 0, len(selected))
			for _, record := range selected {
				got = append(got, record.ID)
			}
			assert.Equal(t, test.want, got)
		})
	}
}

func TestSelectRecordsSpatialAndTemporalIntersections(t *testing.T) {
	records := config.Records{
		{ID: "local", Geometry: jsonValue(map[string]any{"type": "Point"}), Bbox: []float64{4, 52, 5, 53}, Time: &config.RecordTime{Date: "2025-06-01"}},
		{ID: "antimeridian", Geometry: jsonValue(map[string]any{"type": "Point"}), Bbox: []float64{175, -5, -175, 5}, Time: &config.RecordTime{Interval: []string{"2025-01-01T00:00:00Z", "2025-12-31T23:59:59Z"}}},
		{ID: "unknown", Geometry: nil, Time: nil},
		{ID: "remote", Geometry: jsonValue(map[string]any{"type": "Point"}), Bbox: []float64{20, 10, 21, 11}, Time: &config.RecordTime{Timestamp: "2020-01-01T00:00:00Z"}},
	}
	tests := []struct {
		name  string
		query url.Values
		want  []string
	}{
		{name: "bbox intersection includes unknown geometry", query: url.Values{"bbox": {"4.5,52.5,5.5,53.5"}}, want: []string{"local", "unknown"}},
		{name: "bbox handles antimeridian", query: url.Values{"bbox": {"170,-10,-170,10"}}, want: []string{"antimeridian", "unknown"}},
		{name: "datetime instant intersects date and keeps unknown time", query: url.Values{"datetime": {"2025-06-01T12:00:00Z"}}, want: []string{"local", "antimeridian", "unknown"}},
		{name: "datetime interval intersects open record interval", query: url.Values{"datetime": {"2025-06-01T00:00:00Z/2025-06-02T00:00:00Z"}}, want: []string{"local", "antimeridian", "unknown"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selected, err := selectRecords(records, test.query)
			require.NoError(t, err)
			got := make([]string, 0, len(selected))
			for _, record := range selected {
				got = append(got, record.ID)
			}
			assert.Equal(t, test.want, got)
		})
	}
}

func TestPageRecordsHonorsLimitOffsetAndMaximum(t *testing.T) {
	records := make(config.Records, maximumLimit+2)
	for index := range records {
		records[index] = &config.Record{ID: strconv.Itoa(index)}
	}
	tests := []struct {
		name       string
		query      url.Values
		wantFirst  string
		wantCount  int
		wantOffset *int
	}{
		{name: "default limit", query: url.Values{}, wantFirst: "0", wantCount: defaultLimit, wantOffset: intPointer(defaultLimit)},
		{name: "requested limit and offset", query: url.Values{"limit": {"2"}, "offset": {"3"}}, wantFirst: "3", wantCount: 2, wantOffset: intPointer(5)},
		{name: "limit above maximum is capped", query: url.Values{"limit": {"10001"}}, wantFirst: "0", wantCount: maximumLimit, wantOffset: intPointer(maximumLimit)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			page, nextOffset, err := pageRecords(records, test.query)
			require.NoError(t, err)
			require.Len(t, page, test.wantCount)
			assert.Equal(t, test.wantFirst, page[0].ID)
			assert.Equal(t, test.wantOffset, nextOffset)
		})
	}
}

func intPointer(value int) *int {
	return &value
}

func TestSortRecordsUsesConfiguredAndRequestedOrder(t *testing.T) {
	collection := config.RecordsCollection{
		Sortables: []string{"title", "rank"},
		DefaultSortOrder: []config.RecordsSortOrder{
			{Field: "title", Direction: "asc"},
		},
		Records: config.Records{
			{ID: "b", Properties: jsonProperties(map[string]any{"title": "Beta", "rank": 2})},
			{ID: "a", Properties: jsonProperties(map[string]any{"title": "Alpha", "rank": 1})},
			{ID: "c", Properties: jsonProperties(map[string]any{"title": "Gamma", "rank": 3})},
		},
	}
	defaultOrder, err := sortRecords(collection.Records, collection, url.Values{})
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, recordIDs(defaultOrder))

	requestedOrder, err := sortRecords(collection.Records, collection, url.Values{"sortby": {"-rank"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"c", "b", "a"}, recordIDs(requestedOrder))

	_, err = sortRecords(collection.Records, collection, url.Values{"sortby": {"unknown"}})
	require.Error(t, err)
}

func recordIDs(records config.Records) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	return ids
}

func TestSortablesReturnsJsonSchemaForConfiguredProperties(t *testing.T) {
	serverConfig := &config.Config{
		BaseURL: config.URL{URL: &url.URL{Scheme: "https", Host: "api.example.org"}},
		OgcAPI: config.OgcAPI{
			Records: &config.OgcAPIRecords{
				Collections: config.RecordsCollections{{
					ID:        "datasets",
					Sortables: []string{"title", "rank"},
					Records: config.Records{{
						ID:         "bgt",
						Properties: jsonProperties(map[string]any{"title": "BGT", "rank": 4}),
					}},
				}},
			},
		},
	}
	recordsAPI, err := NewRecords(&engine.Engine{Config: serverConfig})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "/collections/datasets/sortables", nil)
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("collectionId", "datasets")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
	recorder := httptest.NewRecorder()
	recordsAPI.Sortables().ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/schema+json", recorder.Header().Get(engine.HeaderContentType))
	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, "https://api.example.org/collections/datasets/sortables", response["$id"])
	properties := response["properties"].(map[string]any)
	assert.Equal(t, "string", properties["title"].(map[string]any)["type"])
	assert.Equal(t, "integer", properties["rank"].(map[string]any)["type"])
}

func TestItemsSupportsHTML(t *testing.T) {
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(filepath.Clean(filepath.Join(workingDirectory, "../../.."))))
	t.Cleanup(func() {
		assert.NoError(t, os.Chdir(workingDirectory))
	})

	serverConfig := &config.Config{
		Version:            "1.0.0",
		Title:              "Records API",
		Abstract:           "Dataset records",
		AvailableLanguages: []config.Language{{Tag: language.Dutch}},
		BaseURL:            config.URL{URL: &url.URL{Scheme: "https", Host: "api.example.org"}},
		OgcAPI: config.OgcAPI{
			Records: &config.OgcAPIRecords{
				Collections: config.RecordsCollections{{
					ID: "datasets",
					Records: config.Records{{
						ID:          "bgt",
						STACVersion: "1.1.0",
						Properties: jsonProperties(map[string]any{
							"datetime":  nil,
							"title":     "BGT",
							"keywords":  []string{"GeoParquet", "example"},
							"publisher": map[string]any{"name": "Open Geospatial Consortium"},
							"rank":      4,
							"enabled":   true,
						}),
						Assets: map[string]config.RecordsAsset{"geoparquet": {Href: "https://example.org/bgt.parquet", Type: "application/vnd.apache.parquet"}},
					}},
				}},
			},
		},
	}
	theme := &config.Theme{Logo: &config.ThemeLogo{}, Color: &config.ThemeColors{Primary: "#000000", Secondary: "#000000", Link: "#000000"}, Includes: &config.ThemeIncludes{}}
	testEngine := engine.NewEngineWithConfig(serverConfig, theme, "", false, true)
	recordsAPI, err := NewRecords(testEngine)
	require.NoError(t, err)
	router := chi.NewRouter()
	router.Get("/collections/{collectionId}/items", recordsAPI.Items())
	router.Get("/collections/{collectionId}/items/{featureId}", recordsAPI.Item())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "https://api.example.org/collections/datasets/items?f=html", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Header().Get(engine.HeaderContentType), mediaTypeHTML)
	assert.Contains(t, recorder.Body.String(), "BGT")

	itemRecorder := httptest.NewRecorder()
	itemRequest := httptest.NewRequest(http.MethodGet, "https://api.example.org/collections/datasets/items/bgt?f=html", nil)
	router.ServeHTTP(itemRecorder, itemRequest)
	require.Equal(t, http.StatusOK, itemRecorder.Code)
	assert.Contains(t, itemRecorder.Header().Get(engine.HeaderContentType), mediaTypeHTML)
	assert.Contains(t, itemRecorder.Body.String(), "<h1>BGT</h1>")
	assert.Contains(t, itemRecorder.Body.String(), "<p>bgt</p>")
	assert.Contains(t, itemRecorder.Body.String(), "https://example.org/bgt.parquet")
	assert.Contains(t, itemRecorder.Body.String(), "application/vnd.apache.parquet")
	assert.Contains(t, itemRecorder.Body.String(), "<dt>title</dt><dd>BGT</dd>")
	assert.Contains(t, itemRecorder.Body.String(), "<dt>datetime</dt><dd>null</dd>")
	assert.Contains(t, itemRecorder.Body.String(), "<dt>rank</dt><dd>4</dd>")
	assert.Contains(t, itemRecorder.Body.String(), "<dt>enabled</dt><dd>true</dd>")
	assert.NotContains(t, itemRecorder.Body.String(), "[34 ")

	markdownRecorder := httptest.NewRecorder()
	markdownRequest := httptest.NewRequest(http.MethodGet, "https://api.example.org/collections/datasets/items?f=md", nil)
	router.ServeHTTP(markdownRecorder, markdownRequest)
	require.Equal(t, http.StatusOK, markdownRecorder.Code)
	assert.Contains(t, markdownRecorder.Header().Get(engine.HeaderContentType), mediaTypeMarkdown)
	assert.Contains(t, markdownRecorder.Body.String(), "BGT")

	itemMarkdownRecorder := httptest.NewRecorder()
	itemMarkdownRequest := httptest.NewRequest(http.MethodGet, "https://api.example.org/collections/datasets/items/bgt?f=md", nil)
	router.ServeHTTP(itemMarkdownRecorder, itemMarkdownRequest)
	require.Equal(t, http.StatusOK, itemMarkdownRecorder.Code)
	assert.Contains(t, itemMarkdownRecorder.Header().Get(engine.HeaderContentType), mediaTypeMarkdown)
	assert.Contains(t, itemMarkdownRecorder.Body.String(), "- **title**: BGT")
	assert.Contains(t, itemMarkdownRecorder.Body.String(), "- **keywords**: [\"GeoParquet\",\"example\"]")
	assert.Contains(t, itemMarkdownRecorder.Body.String(), "- **publisher**: {\"name\":\"Open Geospatial Consortium\"}")
	assert.Contains(t, itemMarkdownRecorder.Body.String(), "- **rank**: 4")
	assert.Contains(t, itemMarkdownRecorder.Body.String(), "- **enabled**: true")
	assert.Contains(t, itemMarkdownRecorder.Body.String(), "- **datetime**: null")
}

func TestRetainRecordIDs(t *testing.T) {
	selected := retainRecordIDs(config.Records{
		{ID: "one"},
		{ID: "two"},
		{ID: "three"},
	}, map[string]bool{"two": true})
	require.Equal(t, []string{"two"}, recordIDs(selected))
}
