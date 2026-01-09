package engine

import (
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template/config files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestEngine_ServePage_LandingPage(t *testing.T) {
	// given
	engine, err := NewEngine("internal/engine/testdata/config_minimal.yaml", "internal/engine/testdata/test_theme.yaml", "", false, true)
	require.NoError(t, err)

	templateKey := NewTemplateKey("internal/ogc/common/core/templates/landing-page.go.json")
	engine.RenderTemplates("/", nil, templateKey)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engine.Serve(w, r, ServeTemplate(templateKey))
	})

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// when
	handler.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get(HeaderContentType))
	assert.Contains(t, recorder.Body.String(), "This is a minimal OGC API, offering only OGC API Common")
}

func TestEngine_ReverseProxy(t *testing.T) {
	// given
	mockTargetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Mock response, received header " + r.Header.Get(HeaderBaseURL)))
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer mockTargetServer.Close()

	engine, targetURL := makeEngine(mockTargetServer)
	rec, req := makeAPICall(t, mockTargetServer.URL)

	// when
	engine.ReverseProxy(rec, req, targetURL, false, "")

	// then
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Mock response, received header https://api.foobar.example/", rec.Body.String())
}

func TestEngine_ReverseProxyAndValidate(t *testing.T) {
	// given
	mockTargetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Mock response, received header " + r.Header.Get(HeaderBaseURL)))
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer mockTargetServer.Close()

	engine, targetURL := makeEngine(mockTargetServer)
	rec, req := makeAPICall(t, mockTargetServer.URL)

	// when
	engine.ReverseProxyAndValidate(rec, req, targetURL, false, MediaTypeJSON, true)

	// then
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Mock response, received header https://api.foobar.example/", rec.Body.String())
}

func TestEngine_ReverseProxy_Status204(t *testing.T) {
	// given
	mockTargetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockTargetServer.Close()

	engine, targetURL := makeEngine(mockTargetServer)
	rec, req := makeAPICall(t, mockTargetServer.URL)

	// when
	engine.ReverseProxy(rec, req, targetURL, true, "audio/wav")

	// then
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "audio/wav", rec.Header().Get(HeaderContentType))
}

type mockShutdownHook struct {
	mutex  sync.Mutex
	called bool
}

func (m *mockShutdownHook) Shutdown() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.called = true
}

func TestEngine_Start(t *testing.T) {
	mockHook := &mockShutdownHook{}
	tests := []struct {
		name          string
		address       string
		shutdownDelay int
		router        *chi.Mux
		hooks         []func()
	}{
		{"Start/stop no delay", "localhost:8080", 0, chi.NewRouter(), []func(){mockHook.Shutdown}},
		{"Start/stop 1s delay", "localhost:8080", 1, chi.NewRouter(), []func(){mockHook.Shutdown}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{shutdownHooks: tt.hooks}

			// Start the server in a separate goroutine
			errChan := make(chan error, 1)
			randomDebugPort := rand.IntN(9999-9000) + 9000 //nolint:gosec // just for testing
			go func() {
				errChan <- e.Start(tt.address, randomDebugPort, tt.shutdownDelay)
			}()

			// Wait for a moment to ensure the server has started
			time.Sleep(10 * time.Millisecond)

			// Send an interrupt signal to stop the server
			p, _ := os.FindProcess(os.Getpid())
			_ = p.Signal(syscall.SIGINT)

			// Wait for the server to shut down and check that there was no error
			err := <-errChan
			require.NoError(t, err)

			// Check that the shutdown hook was called
			mockHook.mutex.Lock()
			called := mockHook.called
			mockHook.mutex.Unlock()
			assert.True(t, called)
		})
	}
}

func TestEngine_Serve(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	e, _ := makeEngine(mockServer)

	// Pre-populate rendered templates for testing
	templateKey := TemplateKey{Name: "test-template", Format: FormatJSON}
	e.Templates.RenderedTemplates[templateKey] = []byte(`{"template": "rendered"}`)

	tests := []struct {
		name                string
		opts                []ServeOption
		expectedStatus      int
		expectedBody        string
		expectedContentType string
	}{
		{
			name: "Serve JSON",
			opts: []ServeOption{
				ServeValidation(false, false),
				ServeJSON(map[string]string{"foo": "bar"}),
				ServeContentType("application/json"),
			},
			expectedStatus:      http.StatusOK,
			expectedBody:        "{\"foo\":\"bar\"}\n",
			expectedContentType: "application/json",
		},
		{
			name: "Serve JSON with response validation",
			opts: []ServeOption{
				ServeValidation(false, true),
				ServeJSON(map[string]string{"foo": "bar"}),
				ServeContentType("application/json"),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "{\"foo\":\"bar\"}\n",
		},
		{
			name: "Serve template",
			opts: []ServeOption{
				ServeValidation(false, false),
				ServeTemplate(templateKey),
			},
			expectedStatus:      http.StatusOK,
			expectedBody:        `{"template": "rendered"}`,
			expectedContentType: MediaTypeJSON,
		},
		{
			name: "Fail serve template",
			opts: []ServeOption{
				ServeValidation(false, false),
				ServeTemplate(TemplateKey{Name: "non-existent"}),
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Serve ore-rendered output",
			opts: []ServeOption{
				ServeValidation(false, false),
				ServePreRenderedOutput([]byte("raw-data")),
				ServeContentType("text/plain"),
			},
			expectedStatus:      http.StatusOK,
			expectedBody:        "raw-data",
			expectedContentType: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, mockServer.URL, nil)

			// when
			e.Serve(w, r, tt.opts...)

			// then
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
			if tt.expectedContentType != "" {
				assert.Equal(t, tt.expectedContentType, w.Header().Get(HeaderContentType))
			}
		})
	}
}

func makeEngine(mockTargetServer *httptest.Server) (*Engine, *url.URL) {
	cfg := &config.Config{
		BaseURL: config.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
	}
	openAPI := newOpenAPI(cfg, []string{""}, nil)
	engine := &Engine{
		Config:  cfg,
		OpenAPI: openAPI,
		Templates: &Templates{
			RenderedTemplates: map[TemplateKey][]byte{},
		},
		CN: newContentNegotiation(cfg.AvailableLanguages),
	}
	targetURL, _ := url.Parse(mockTargetServer.URL)

	return engine, targetURL
}

func makeAPICall(t *testing.T, mockTargetServer string) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, mockTargetServer+"/some/path", nil)
	if err != nil {
		t.Fatal(err)
	}

	return rec, req
}
