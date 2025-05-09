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
	"golang.org/x/text/language"

	"github.com/stretchr/testify/assert"
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
	engine, err := NewEngine("internal/engine/testdata/config_minimal.yaml", "", false, true)
	assert.NoError(t, err)

	templateKey := NewTemplateKey("internal/ogc/common/core/templates/landing-page.go.json") // Using the backward-compatible API
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

func TestTemplateKeyWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  []TemplateKeyOption
		expected TemplateKey
	}{
		{
			name:    "default values",
			options: nil,
			expected: TemplateKey{
				Name:               "landing-page.go.json",
				Directory:          "internal/ogc/common/core/templates",
				Format:             "json",
				Language:           language.Dutch,
				InstanceName:       "",
				MediaTypeOverwrite: "",
			},
		},
		{
			name:    "with language",
			options: []TemplateKeyOption{WithLanguage(language.English)},
			expected: TemplateKey{
				Name:               "landing-page.go.json",
				Directory:          "internal/ogc/common/core/templates",
				Format:             "json",
				Language:           language.English,
				InstanceName:       "",
				MediaTypeOverwrite: "",
			},
		},
		{
			name:    "with instance name",
			options: []TemplateKeyOption{WithInstanceName("test-instance")},
			expected: TemplateKey{
				Name:               "landing-page.go.json",
				Directory:          "internal/ogc/common/core/templates",
				Format:             "json",
				Language:           language.Dutch,
				InstanceName:       "test-instance",
				MediaTypeOverwrite: "",
			},
		},
		{
			name:    "with media type",
			options: []TemplateKeyOption{WithMediaTypeOverwrite("application/docx")},
			expected: TemplateKey{
				Name:               "landing-page.go.json",
				Directory:          "internal/ogc/common/core/templates",
				Format:             "json",
				Language:           language.Dutch,
				InstanceName:       "",
				MediaTypeOverwrite: "application/docx",
			},
		},
		{
			name: "with multiple options",
			options: []TemplateKeyOption{
				WithLanguage(language.English),
				WithInstanceName("test-instance"),
				WithMediaTypeOverwrite("application/docx"),
			},
			expected: TemplateKey{
				Name:               "landing-page.go.json",
				Directory:          "internal/ogc/common/core/templates",
				Format:             "json",
				Language:           language.English,
				InstanceName:       "test-instance",
				MediaTypeOverwrite: "application/docx",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := NewTemplateKey("internal/ogc/common/core/templates/landing-page.go.json", tt.options...)
			assert.Equal(t, tt.expected.Name, key.Name)
			assert.Equal(t, tt.expected.Directory, key.Directory)
			assert.Equal(t, tt.expected.Format, key.Format)
			assert.Equal(t, tt.expected.Language, key.Language)
			assert.Equal(t, tt.expected.InstanceName, key.InstanceName)
			assert.Equal(t, tt.expected.MediaTypeOverwrite, key.MediaTypeOverwrite)
		})
	}
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
	assert.Equal(t, rec.Body.String(), "Mock response, received header https://api.foobar.example/")
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
			assert.NoError(t, err)

			// Check that the shutdown hook was called
			mockHook.mutex.Lock()
			called := mockHook.called
			mockHook.mutex.Unlock()
			assert.True(t, called)
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
	}
	targetURL, _ := url.Parse(mockTargetServer.URL)
	return engine, targetURL
}

func makeAPICall(t *testing.T, mockTargetServer string) (*httptest.ResponseRecorder, *http.Request) {
	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, mockTargetServer+"/some/path", nil)
	if err != nil {
		t.Fatal(err)
	}
	return rec, req
}
