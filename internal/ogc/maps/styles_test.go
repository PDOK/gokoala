package maps

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const relMap = "https://www.opengis.net/def/rel/ogc/1.0/map"

type stylesLink struct {
	Rel  string `json:"rel"`
	Type string `json:"type"`
	Href string `json:"href"`
}

type stylesResponse struct {
	Links   []stylesLink `json:"links"`
	Default string       `json:"default"`
	Styles  []struct {
		ID    string       `json:"id"`
		Title string       `json:"title"`
		Links []stylesLink `json:"links"`
	} `json:"styles"`
}

func TestStylesDescLinks(t *testing.T) {
	router := newTestRouter(t)

	rr := doGet(router, "/collections/"+testCollection+"/styles", "application/json")
	require.Equal(t, http.StatusOK, rr.Code)

	var stylesRes stylesResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &stylesRes))
	assert.Equal(t, testStyle, stylesRes.Default)
	require.Len(t, stylesRes.Styles, 1)
	assert.Equal(t, testStyle, stylesRes.Styles[0].ID)

	var mapLinks []stylesLink
	for _, link := range stylesRes.Styles[0].Links {
		if link.Rel == relMap {
			mapLinks = append(mapLinks, link)
		}
	}
	require.NotEmpty(t, mapLinks, "style must have a link with rel %s", relMap)
	for _, link := range mapLinks {
		assert.Contains(t, link.Href, "/collections/"+testCollection+"/styles/"+testStyle+"/map")
		assert.NotEmpty(t, link.Type)
	}
}

func TestStylesFormatsAndErrors(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		name            string
		path            string
		accept          string
		wantStatus      int
		wantContentType string
	}{
		{
			name:            "JSON",
			path:            "/collections/" + testCollection + "/styles",
			accept:          "application/json",
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
		},
		{
			name:            "HTML",
			path:            "/collections/" + testCollection + "/styles?f=html",
			wantStatus:      http.StatusOK,
			wantContentType: "text/html",
		},
		{
			name:       "unknown collection",
			path:       "/collections/does-not-exist/styles",
			accept:     "application/json",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "image format isn't supported on the styles resource",
			path:       "/collections/" + testCollection + "/styles?f=png",
			wantStatus: http.StatusNotAcceptable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := doGet(router, tt.path, tt.accept)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantContentType != "" {
				assert.Contains(t, rr.Header().Get("Content-Type"), tt.wantContentType)
			}
		})
	}
}
