package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	texttemplate "text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

const (
	specPath          = templatesDir + "openapi/"
	preamble          = specPath + "preamble.go.json"
	commonCollections = specPath + "common-collections.go.json"
	tilesSpec         = specPath + "tiles.go.json"
	stylesSpec        = specPath + "styles.go.json"
	geoVolumesSpec    = specPath + "3dgeovolumes.go.json"
	commonSpec        = specPath + "common.go.json"
	HTMLRegex         = `<[/]?([a-zA-Z]+).*?>`
)

type OpenAPI struct {
	config   *Config
	spec     *openapi3.T
	SpecJSON []byte
	router   routers.Router
}

func newOpenAPI(config *Config, openAPIFile string) *OpenAPI {
	setupRequestResponseValidation()
	ctx := context.Background()

	// order matters, see mergeSpecs for details.
	defaultOpenAPIFiles := []string{commonSpec}
	if config.OgcAPI.GeoVolumes != nil {
		defaultOpenAPIFiles = append(defaultOpenAPIFiles, commonCollections)
	}
	if config.OgcAPI.Tiles != nil {
		defaultOpenAPIFiles = append(defaultOpenAPIFiles, tilesSpec)
	}
	if config.OgcAPI.Styles != nil {
		defaultOpenAPIFiles = append(defaultOpenAPIFiles, stylesSpec)
	}
	if config.OgcAPI.GeoVolumes != nil {
		defaultOpenAPIFiles = append(defaultOpenAPIFiles, geoVolumesSpec)
	}
	// add preamble first
	openAPIFiles := []string{preamble}
	if openAPIFile != "" {
		// add provided spec thereafter, to allow it to override defaults of following specs
		openAPIFiles = append(openAPIFiles, openAPIFile)
	}
	openAPIFiles = append(openAPIFiles, defaultOpenAPIFiles...)

	resultSpec, resultSpecJSON := mergeSpecs(ctx, config, openAPIFiles)
	validateSpec(ctx, resultSpec, resultSpecJSON)

	for _, server := range resultSpec.Servers {
		server.URL = normalizeBaseURL(server.URL)
		log.Printf("URL used for OpenAPI validation: %v", server.URL)
	}

	return &OpenAPI{
		config:   config,
		spec:     resultSpec,
		SpecJSON: PrettyPrintJSON(resultSpecJSON, ""),
		router:   newOpenAPIRouter(resultSpec),
	}
}

func setupRequestResponseValidation() {
	htmlRegex := regexp.MustCompile(HTMLRegex)

	openapi3filter.RegisterBodyDecoder("text/html",
		func(body io.Reader, header http.Header, ref *openapi3.SchemaRef,
			fn openapi3filter.EncodingFn) (interface{}, error) {

			data, err := io.ReadAll(body)
			if err != nil {
				return nil, errors.New("failed to read response body")
			}
			if !htmlRegex.Match(data) {
				return nil, errors.New("response doesn't contain HTML")
			}
			return string(data), nil
		})

	openapi3filter.RegisterBodyDecoder("application/vnd.mapbox.tile+json",
		func(body io.Reader, header http.Header, schema *openapi3.SchemaRef,
			fn openapi3filter.EncodingFn) (interface{}, error) {
			var value interface{}
			dec := json.NewDecoder(body)
			dec.UseNumber()
			if err := dec.Decode(&value); err != nil {
				return nil, errors.New("response doesn't contain valid JSON")
			}
			return value, nil
		})
}

// mergeSpecs merges the given OpenAPI specs.
//
// Order matters! We start with the preamble, it is highest in rank and there's no way to override it.
// Then the files are merged according to their given order. Files that are merged first
// have a higher change of getting their changes in the final spec than files that follow later.
//
// The OpenAPI spec optionally provided through the CLI should be the second (after preamble) item in the
// `files` slice since it allows the user to override other/default specs.
func mergeSpecs(ctx context.Context, config *Config, files []string) (*openapi3.T, []byte) {
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}

	if len(files) < 1 {
		log.Fatalf("files can't be empty, at least OGC Common is expected")
	}
	var resultSpecJSON []byte
	var resultSpec *openapi3.T

	for _, file := range files {
		specJSON := renderOpenAPITemplate(config, file)
		_ = loadSpec(loader, specJSON)
		var mergedJSON []byte
		if resultSpecJSON == nil {
			mergedJSON = specJSON
		} else {
			var err error
			mergedJSON, err = mergeJSON(resultSpecJSON, specJSON)
			if err != nil {
				log.Print(string(mergedJSON))
				log.Fatalf("failed to merge openapi specs: %v", err)
			}
		}
		resultSpecJSON = mergedJSON
		resultSpec = loadSpec(loader, mergedJSON)
	}
	return resultSpec, resultSpecJSON
}

func loadSpec(loader *openapi3.Loader, mergedJSON []byte, fileName ...string) *openapi3.T {
	resultSpec, err := loader.LoadFromData(mergedJSON)
	if err != nil {
		log.Print(string(mergedJSON))
		log.Fatalf("failed to load merged openapi spec %s, due to %v", fileName, err)
	}
	return resultSpec
}

func validateSpec(ctx context.Context, finalSpec *openapi3.T, finalSpecRaw []byte) {
	// Validate OGC OpenAPI spec. Note: the examples provided in the official spec aren't valid.
	err := finalSpec.Validate(ctx, openapi3.DisableExamplesValidation())
	if err != nil {
		log.Print(string(finalSpecRaw))
		log.Fatalf("invalid openapi spec: %v", err)
	}
}

func newOpenAPIRouter(doc *openapi3.T) routers.Router {
	openAPIRouter, err := gorillamux.NewRouter(doc)
	if err != nil {
		log.Fatalf("failed to setup openapi router: %v", err)
	}
	return openAPIRouter
}

func renderOpenAPITemplate(config *Config, fileName string) []byte {
	file := filepath.Clean(fileName)
	compiled := texttemplate.Must(texttemplate.New(filepath.Base(file)).Funcs(customFuncs).ParseFiles(file))

	var rendered bytes.Buffer
	if err := compiled.Execute(&rendered, &TemplateData{Config: config}); err != nil {
		log.Fatalf("failed to render %s, error: %v", file, err)
	}
	return rendered.Bytes()
}

func (o *OpenAPI) validateRequest(r *http.Request) error {
	requestValidationInput, _ := o.getRequestValidationInput(r)
	if requestValidationInput != nil {
		err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput)
		if err != nil {
			return fmt.Errorf("invalid request, doesn't conform to openapi spec %w", err)
		}
	}
	return nil
}

func (o *OpenAPI) validateResponse(contentType string, body []byte, r *http.Request) error {
	requestValidationInput, _ := o.getRequestValidationInput(r)
	if requestValidationInput != nil {
		responseHeaders := http.Header{"Content-Type": []string{contentType}}
		responseCode := 200

		responseValidationInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: requestValidationInput,
			Status:                 responseCode,
			Header:                 responseHeaders,
		}
		responseValidationInput.SetBodyBytes(body)
		err := openapi3filter.ValidateResponse(context.Background(), responseValidationInput)
		if err != nil {
			return fmt.Errorf("response doesn't conform to openapi spec: %w", err)
		}
	}
	return nil
}

func (o *OpenAPI) getRequestValidationInput(r *http.Request) (*openapi3filter.RequestValidationInput, error) {
	route, pathParams, err := o.router.FindRoute(r)
	if err != nil {
		log.Printf("route not found in openapi spec for url %s (host: %s), "+
			"skipping OpenAPI validation", r.URL, r.Host)
		return nil, err
	}
	return &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	}, nil
}

// normalizeBaseURL normalizes the given base URL so our OpenAPI validator is able to match
// requests against the OpenAPI spec. This involves:
//
//   - striping the context root (path) from the base URL. If you use a context root we expect
//     you have a proxying fronting GoKoala it from requests, therefore we also need to strip it from
//     the base URL used during OpenAPI validation
//
//   - replacing HTTPS scheme with HTTP. Since GoKoala doesn't support HTTPS we always perform
//     OpenAPI validation against HTTP requests. Note: it's possible to offer GoKoala over HTTPS, but you'll
//     need to take care of that in your proxy server (or loadbalancer/service mesh/etc) fronting GoKoala.
func normalizeBaseURL(baseURL string) string {
	serverURL, _ := url.Parse(baseURL)
	result := strings.Replace(baseURL, serverURL.Scheme, "http", 1)
	result = strings.Replace(result, serverURL.Path, "", 1)
	return result
}
