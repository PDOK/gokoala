package features

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/records"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeaturesDispatchesRecordCatalogToRecordsHandler(t *testing.T) {
	serverConfig := &config.Config{
		BaseURL: config.URL{URL: &url.URL{Scheme: "https", Host: "api.example.org"}},
		OgcAPI: config.OgcAPI{
			Records: &config.OgcAPIRecords{
				Collections: config.RecordsCollections{{
					ID: "datasets",
					Records: config.Records{{
						ID:          "bgt",
						STACVersion: "1.1.0",
						Properties:  map[string]config.JSONValue{"datetime": config.JSONValue(`"2025-01-01T00:00:00Z"`)},
						Assets:      map[string]config.RecordsAsset{},
					}},
				}},
			},
		},
	}
	features := &Features{}
	recordsAPI, err := records.NewRecords(&engine.Engine{Config: serverConfig})
	require.NoError(t, err)
	features.SetRecords(recordsAPI)
	router := chi.NewRouter()
	router.Get("/collections/{collectionId}/items", features.Features())
	router.Get("/collections/{collectionId}/items/{featureId}", features.Feature())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/collections/datasets/items", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, "FeatureCollection", response["type"])
	feature := response["features"].([]any)[0].(map[string]any)
	assert.Equal(t, "bgt", feature["id"])

	itemRecorder := httptest.NewRecorder()
	itemRequest := httptest.NewRequest(http.MethodGet, "/collections/datasets/items/bgt", nil)
	router.ServeHTTP(itemRecorder, itemRequest)
	require.Equal(t, http.StatusOK, itemRecorder.Code)
	var item map[string]any
	require.NoError(t, json.Unmarshal(itemRecorder.Body.Bytes(), &item))
	assert.Equal(t, "bgt", item["id"])
}
