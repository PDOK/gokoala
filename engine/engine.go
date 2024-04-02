package engine

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	texttemplate "text/template"
	"time"

	"github.com/PDOK/gokoala/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	templatesDir    = "engine/templates/"
	shutdownTimeout = 5 * time.Second

	HeaderLink           = "Link"
	HeaderAccept         = "Accept"
	HeaderAcceptLanguage = "Accept-Language"
	HeaderContentType    = "Content-Type"
	HeaderContentLength  = "Content-Length"
	HeaderContentCrs     = "Content-Crs"
	HeaderBaseURL        = "X-BaseUrl"
	HeaderRequestedWith  = "X-Requested-With"
	HeaderAPIVersion     = "API-Version"
)

// Engine encapsulates shared non-OGC API specific logic
type Engine struct {
	Config    *config.Config
	OpenAPI   *OpenAPI
	Templates *Templates
	CN        *ContentNegotiation
	Router    *chi.Mux

	shutdownHooks []func()
}

// NewEngine builds a new Engine
func NewEngine(configFile string, openAPIFile string, enableTrailingSlash bool, enableCORS bool) (*Engine, error) {
	config, err := config.NewConfig(configFile)
	if err != nil {
		return nil, err
	}
	return NewEngineWithConfig(config, openAPIFile, enableTrailingSlash, enableCORS), nil
}

// NewEngineWithConfig builds a new Engine
func NewEngineWithConfig(config *config.Config, openAPIFile string, enableTrailingSlash bool, enableCORS bool) *Engine {
	contentNegotiation := newContentNegotiation(config.AvailableLanguages)
	templates := newTemplates(config)
	openAPI := newOpenAPI(config, []string{openAPIFile}, nil)
	router := newRouter(config.Version, enableTrailingSlash, enableCORS)

	engine := &Engine{
		Config:    config,
		OpenAPI:   openAPI,
		Templates: templates,
		CN:        contentNegotiation,
		Router:    router,
	}

	if config.Resources != nil {
		newResourcesEndpoint(engine) // Resources endpoint to serve static assets
	}
	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		SafeWrite(w.Write, []byte("OK")) // Health endpoint
	})
	return engine
}

// Start the engine by initializing all components and starting the server
func (e *Engine) Start(address string, debugPort int, shutdownDelay int) error {
	// debug server (binds to localhost).
	if debugPort > 0 {
		go func() {
			debugAddress := fmt.Sprintf("localhost:%d", debugPort)
			debugRouter := chi.NewRouter()
			debugRouter.Use(middleware.Logger)
			debugRouter.Mount("/debug", middleware.Profiler())
			err := e.startServer("debug server", debugAddress, 0, debugRouter)
			if err != nil {
				log.Fatalf("debug server failed %v", err)
			}
		}()
	}

	// main server
	return e.startServer("main server", address, shutdownDelay, e.Router)
}

// startServer creates and starts an HTTP server, also takes care of graceful shutdown
func (e *Engine) startServer(name string, address string, shutdownDelay int, router *chi.Mux) error {
	// create HTTP server
	server := http.Server{
		Addr:    address,
		Handler: router,

		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	go func() {
		log.Printf("%s listening on http://%2s", name, address)
		// ListenAndServe always returns a non-nil error. After Shutdown or
		// Close, the returned error is ErrServerClosed
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to shutdown %s: %v", name, err)
		}
	}()

	// listen for interrupt signal and then perform shutdown
	<-ctx.Done()
	stop()

	// execute shutdown hooks
	for _, shutdownHook := range e.shutdownHooks {
		shutdownHook()
	}

	if shutdownDelay > 0 {
		log.Printf("stop signal received, initiating shutdown of %s after %d seconds delay", name, shutdownDelay)
		time.Sleep(time.Duration(shutdownDelay) * time.Second)
	}
	log.Printf("shutting down %s gracefully", name)

	// shutdown with a max timeout.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return server.Shutdown(timeoutCtx)
}

// RegisterShutdownHook register a func to execute during graceful shutdown, e.g. to clean up resources.
func (e *Engine) RegisterShutdownHook(fn func()) {
	e.shutdownHooks = append(e.shutdownHooks, fn)
}

// RebuildOpenAPI rebuild the full OpenAPI spec with the newly given parameters.
// Use only once during bootstrap for specific use cases! For example: when you want to expand a
// specific part of the OpenAPI spec with data outside the configuration file (e.g. from a database).
func (e *Engine) RebuildOpenAPI(openAPIParams any) {
	e.OpenAPI = newOpenAPI(e.Config, e.OpenAPI.extraOpenAPIFiles, openAPIParams)
}

// ParseTemplate parses both HTML and non-HTML templates depending on the format given in the TemplateKey and
// stores it in the engine for future rendering using RenderAndServePage.
func (e *Engine) ParseTemplate(key TemplateKey) {
	e.Templates.parseAndSaveTemplate(key)
}

// RenderTemplates renders both HTML and non-HTML templates depending on the format given in the TemplateKey.
// This method also performs OpenAPI validation of the rendered template, therefore we also need the URL path.
// The rendered templates are stored in the engine for future serving using ServePage.
func (e *Engine) RenderTemplates(urlPath string, breadcrumbs []Breadcrumb, keys ...TemplateKey) {
	for _, key := range keys {
		e.Templates.renderAndSaveTemplate(key, breadcrumbs, nil)

		// we already perform OpenAPI validation here during startup to catch
		// issues early on, in addition to runtime OpenAPI response validation
		// all templates are created in all available languages, hence all are checked
		for lang := range e.Templates.localizers {
			key.Language = lang
			if err := e.validateStaticResponse(key, urlPath); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// RenderTemplatesWithParams renders both HTMl and non-HTML templates depending on the format given in the TemplateKey.
// This method does not perform OpenAPI validation of the rendered template (will be done during runtime).
func (e *Engine) RenderTemplatesWithParams(params any, breadcrumbs []Breadcrumb, keys ...TemplateKey) {
	for _, key := range keys {
		e.Templates.renderAndSaveTemplate(key, breadcrumbs, params)
	}
}

// RenderAndServePage renders an already parsed HTML or non-HTML template and renders it on-the-fly depending
// on the format in the given TemplateKey. The result isn't store in engine, it's served directly to the client.
//
// NOTE: only used this for dynamic pages that can't be pre-rendered and cached (e.g. with data from a backing store).
func (e *Engine) RenderAndServePage(w http.ResponseWriter, r *http.Request, key TemplateKey,
	params any, breadcrumbs []Breadcrumb) {

	// validate request
	if err := e.OpenAPI.ValidateRequest(r); err != nil {
		log.Printf("%v", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get template
	parsedTemplate, err := e.Templates.getParsedTemplate(key)
	if err != nil {
		log.Printf("%v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// render output
	var output []byte
	if key.Format == FormatHTML {
		htmlTmpl := parsedTemplate.(*htmltemplate.Template)
		output = e.Templates.renderHTMLTemplate(htmlTmpl, r.URL, params, breadcrumbs, "")
	} else {
		jsonTmpl := parsedTemplate.(*texttemplate.Template)
		output = e.Templates.renderNonHTMLTemplate(jsonTmpl, params, key, "")
	}
	contentType := e.CN.formatToMediaType(key.Format)

	// validate response
	if err := e.OpenAPI.ValidateResponse(contentType, output, r); err != nil {
		log.Printf("%v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return response output to client
	if contentType != "" {
		w.Header().Set(HeaderContentType, contentType)
	}
	SafeWrite(w.Write, output)
}

// ServePage serves a pre-rendered template while also validating against the OpenAPI spec
func (e *Engine) ServePage(w http.ResponseWriter, r *http.Request, templateKey TemplateKey) {
	// validate request
	if err := e.OpenAPI.ValidateRequest(r); err != nil {
		log.Printf("%v", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// render output
	output, err := e.Templates.getRenderedTemplate(templateKey)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	contentType := e.CN.formatToMediaType(templateKey.Format)

	// validate response
	if err := e.OpenAPI.ValidateResponse(contentType, output, r); err != nil {
		log.Printf("%v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return response output to client
	if contentType != "" {
		w.Header().Set(HeaderContentType, contentType)
	}
	SafeWrite(w.Write, output)
}

// ServeResponse serves the given response (arbitrary bytes) while also validating against the OpenAPI spec
func (e *Engine) ServeResponse(w http.ResponseWriter, r *http.Request,
	validateRequest bool, validateResponse bool, contentType string, response []byte) {

	if validateRequest {
		if err := e.OpenAPI.ValidateRequest(r); err != nil {
			log.Printf("%v", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if validateResponse {
		if err := e.OpenAPI.ValidateResponse(contentType, response, r); err != nil {
			log.Printf("%v", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// return response output to client
	if contentType != "" {
		w.Header().Set(HeaderContentType, contentType)
	}
	SafeWrite(w.Write, response)
}

// ReverseProxy forwards given HTTP request to given target server, and optionally tweaks response
func (e *Engine) ReverseProxy(w http.ResponseWriter, r *http.Request, target *url.URL,
	prefer204 bool, contentTypeOverwrite string) {
	e.ReverseProxyAndValidate(w, r, target, prefer204, contentTypeOverwrite, false)
}

// ReverseProxy forwards given HTTP request to given target server, and optionally tweaks and validates response
func (e *Engine) ReverseProxyAndValidate(w http.ResponseWriter, r *http.Request, target *url.URL,
	prefer204 bool, contentTypeOverwrite string, validateResponse bool) {

	rewrite := func(r *httputil.ProxyRequest) {
		r.Out.URL = target
		r.Out.Host = ""   // Don't pass Host header (similar to Traefik's passHostHeader=false)
		r.SetXForwarded() // Set X-Forwarded-* headers.
		r.Out.Header.Set(HeaderBaseURL, e.Config.BaseURL.String())
	}

	modifyResponse := func(proxyRes *http.Response) error {
		if prefer204 {
			// OGC spec: If the tile has no content due to lack of data in the area, but is within the data
			// resource its tile matrix sets and tile matrix sets limits, the HTTP response will use the status
			// code either 204 (indicating an empty tile with no content) or a 200
			if proxyRes.StatusCode == http.StatusNotFound {
				proxyRes.StatusCode = http.StatusNoContent
				removeBody(proxyRes)
			}
		}
		if contentTypeOverwrite != "" {
			proxyRes.Header.Set(HeaderContentType, contentTypeOverwrite)
		}
		if contentType := proxyRes.Header.Get(HeaderContentType); contentType == MediaTypeJSON && validateResponse {
			reader, err := gzip.NewReader(proxyRes.Body)
			if err != nil {
				log.Printf("%v", err.Error())
				return err
			}
			res, err := io.ReadAll(reader)
			if err != nil {
				log.Printf("%v", err.Error())
				return err
			}
			e.ServeResponse(w, r, false, true, contentType, res)
		}
		return nil
	}

	reverseProxy := &httputil.ReverseProxy{Rewrite: rewrite, ModifyResponse: modifyResponse}
	reverseProxy.ServeHTTP(w, r)
}

func removeBody(proxyRes *http.Response) {
	buf := bytes.NewBuffer(make([]byte, 0))
	proxyRes.Body = io.NopCloser(buf)
	proxyRes.Header[HeaderContentLength] = []string{"0"}
	proxyRes.Header[HeaderContentType] = []string{}
}

func (e *Engine) validateStaticResponse(key TemplateKey, urlPath string) error {
	template, _ := e.Templates.getRenderedTemplate(key)
	serverURL := normalizeBaseURL(e.Config.BaseURL.String())
	req, err := http.NewRequest(http.MethodGet, serverURL+urlPath, nil)
	if err != nil {
		return fmt.Errorf("failed to construct request to validate %s "+
			"template against OpenAPI spec %v", key.Name, err)
	}
	err = e.OpenAPI.ValidateResponse(e.CN.formatToMediaType(key.Format), template, req)
	if err != nil {
		return fmt.Errorf("validation of template %s failed: %w", key.Name, err)
	}
	return nil
}

// SafeWrite executes the given http.ResponseWriter.Write while logging errors
func SafeWrite(write func([]byte) (int, error), body []byte) {
	_, err := write(body)
	if err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
