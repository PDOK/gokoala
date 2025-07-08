package engine

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockReverseProxy struct {
	mock.Mock
}

func (m *MockReverseProxy) Proxy(w http.ResponseWriter, r *http.Request, target *url.URL, prefer204 bool, overwrite string) {
	m.Called(w, r, target, prefer204, overwrite)
}

func TestDir(t *testing.T) {
	tests := []struct {
		name           string
		resourcesDir   string
		urlParam       string
		expectedStatus int
		expectedLog    string
	}{
		{
			name:           "valid url",
			resourcesDir:   "docs",
			urlParam:       "foo.txt",
			expectedStatus: http.StatusOK,
			expectedLog:    "",
		},
		{
			name:           "invalid url",
			resourcesDir:   "docs",
			urlParam:       "non-existing-file",
			expectedStatus: http.StatusNotFound,
			expectedLog:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			engine, err := NewEngine("internal/engine/testdata/config_resources_dir.yaml", "internal/engine/testdata/test_theme.yaml", "", false, true)
			assert.NoError(t, err)
			r := httptest.NewRequest(http.MethodGet, "/resources/"+tt.urlParam, nil)
			w := httptest.NewRecorder()
			var logOutput strings.Builder
			log.SetOutput(&logOutput)

			// when
			newResourcesEndpoint(engine)
			engine.Router.ServeHTTP(w, r)

			// then
			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
			if tt.expectedLog != "" {
				assert.Contains(t, logOutput.String(), tt.expectedLog)
			}
		})
	}
}

func TestProxy(t *testing.T) {
	tests := []struct {
		name           string
		resourcesURL   string
		urlParam       string
		expectedStatus int
		expectedLog    string
		expectProxy    bool
	}{
		{
			name:           "valid url",
			resourcesURL:   "http://example.com/resources",
			urlParam:       "file",
			expectedStatus: http.StatusOK,
			expectedLog:    "",
			expectProxy:    true,
		},
		{
			name:           "invalid url",
			resourcesURL:   "foo bar",
			urlParam:       "file",
			expectedStatus: http.StatusInternalServerError,
			expectedLog:    "invalid target url, can't proxy resources",
			expectProxy:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			mockReverseProxy := MockReverseProxy{}
			if tt.expectProxy {
				mockReverseProxy.On("Proxy", mock.Anything, mock.Anything, mock.Anything, false, "").Return()
			}
			r := httptest.NewRequest(http.MethodGet, "/resources/"+tt.urlParam, nil)
			w := httptest.NewRecorder()
			var logOutput strings.Builder
			log.SetOutput(&logOutput)

			// when
			proxyHandler := proxy(mockReverseProxy.Proxy, tt.resourcesURL)
			proxyHandler(w, r)

			// then
			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
			if tt.expectedLog != "" {
				assert.Contains(t, logOutput.String(), tt.expectedLog)
			}
			if tt.expectProxy {
				mockReverseProxy.AssertCalled(t, "Proxy", w, r, mock.Anything, false, "")
			} else {
				mockReverseProxy.AssertNotCalled(t, "Proxy", w, r, mock.Anything, false, "")
			}
		})
	}
}
