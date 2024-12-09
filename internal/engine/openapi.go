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

	gomagpieconfig "github.com/PDOK/gomagpie/config"
	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/PDOK/gomagpie/internal/engine/util"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

const (
	specPath          = templatesDir + "openapi/"
	preamble          = specPath + "preamble.go.json"
	problems          = specPath + "problems.go.json"
	commonCollections = specPath + "common-collections.go.json"
	commonSpec        = specPath + "common.go.json"
	HTMLRegex         = `<[/]?([a-zA-Z]+).*?>`
)

type OpenAPI struct {
	spec     *openapi3.T
	SpecJSON []byte

	config *gomagpieconfig.Config
	router routers.Router
}

func newOpenAPI(config *gomagpieconfig.Config) *OpenAPI {
	setupRequestResponseValidation()
	ctx := context.Background()

	// order matters, see mergeSpecs for details.
	defaultOpenAPIFiles := []string{commonSpec}
	if config.AllCollections() != nil {
		defaultOpenAPIFiles = append(defaultOpenAPIFiles, commonCollections)
	}

	// add preamble first
	openAPIFiles := []string{preamble}
	openAPIFiles = append(openAPIFiles, defaultOpenAPIFiles...)

	resultSpec, resultSpecJSON := mergeSpecs(ctx, config, openAPIFiles, nil)
	validateSpec(ctx, resultSpec, resultSpecJSON)

	for _, server := range resultSpec.Servers {
		server.URL = normalizeBaseURL(server.URL)
	}

	return &OpenAPI{
		config:   config,
		spec:     resultSpec,
		SpecJSON: util.PrettyPrintJSON(resultSpecJSON, ""),
		router:   newOpenAPIRouter(resultSpec),
	}
}

func setupRequestResponseValidation() {
	htmlRegex := regexp.MustCompile(HTMLRegex)

	openapi3filter.RegisterBodyDecoder(MediaTypeHTML,
		func(body io.Reader, _ http.Header, _ *openapi3.SchemaRef,
			_ openapi3filter.EncodingFn) (any, error) {

			data, err := io.ReadAll(body)
			if err != nil {
				return nil, errors.New("failed to read response body")
			}
			if !htmlRegex.Match(data) {
				return nil, errors.New("response doesn't contain HTML")
			}
			return string(data), nil
		})

	for _, mediaType := range MediaTypeJSONFamily {
		openapi3filter.RegisterBodyDecoder(mediaType,
			func(body io.Reader, _ http.Header, _ *openapi3.SchemaRef,
				_ openapi3filter.EncodingFn) (any, error) {
				var value any
				dec := json.NewDecoder(body)
				dec.UseNumber()
				if err := dec.Decode(&value); err != nil {
					return nil, errors.New("response doesn't contain valid JSON")
				}
				return value, nil
			})
	}
}

// mergeSpecs merges the given OpenAPI specs.
//
// Order matters! We start with the preamble, it is highest in rank and there's no way to override it.
// Then the files are merged according to their given order. Files that are merged first
// have a higher change of getting their changes in the final spec than files that follow later.
//
// The OpenAPI spec optionally provided through the CLI should be the second (after preamble) item in the
// `files` slice since it allows the user to override other/default specs.
func mergeSpecs(ctx context.Context, config *gomagpieconfig.Config, files []string, params any) (*openapi3.T, []byte) {
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: false}

	if len(files) < 1 {
		log.Fatalf("files can't be empty, at least OGC Common is expected")
	}
	var resultSpecJSON []byte
	var resultSpec *openapi3.T

	for _, file := range files {
		if file == "" {
			continue
		}
		specJSON := renderOpenAPITemplate(config, file, params)
		var mergedJSON []byte
		if resultSpecJSON == nil {
			mergedJSON = specJSON
		} else {
			var err error
			mergedJSON, err = util.MergeJSON(resultSpecJSON, specJSON, orderByOpenAPIConvention)
			if err != nil {
				log.Print(string(mergedJSON))
				log.Fatalf("failed to merge OpenAPI specs: %v", err)
			}
		}
		resultSpecJSON = mergedJSON
		resultSpec = loadSpec(loader, mergedJSON)
	}
	return resultSpec, resultSpecJSON
}

func orderByOpenAPIConvention(output map[string]any) any {
	result := orderedmap.New[string, any]()
	// OpenAPI specs are commonly ordered according to the following sequence.
	desiredOrder := []string{"openapi", "info", "servers", "paths", "components"}
	for _, order := range desiredOrder {
		for k, v := range output {
			if k == order {
				result.Set(k, v)
			}
		}
	}
	// add remaining keys
	for k, v := range output {
		result.Set(k, v)
	}
	return result
}

func loadSpec(loader *openapi3.Loader, mergedJSON []byte, fileName ...string) *openapi3.T {
	resultSpec, err := loader.LoadFromData(mergedJSON)
	if err != nil {
		log.Print(string(mergedJSON))
		log.Fatalf("failed to load merged OpenAPI spec %s, due to %v", fileName, err)
	}
	return resultSpec
}

func validateSpec(ctx context.Context, finalSpec *openapi3.T, finalSpecRaw []byte) {
	// Validate OGC OpenAPI spec. Note: the examples provided in the official spec aren't valid.
	err := finalSpec.Validate(ctx, openapi3.DisableExamplesValidation())
	if err != nil {
		log.Print(string(finalSpecRaw))
		log.Fatalf("invalid OpenAPI spec: %v", err)
	}
}

func newOpenAPIRouter(doc *openapi3.T) routers.Router {
	openAPIRouter, err := gorillamux.NewRouter(doc)
	if err != nil {
		log.Fatalf("failed to setup OpenAPI router: %v", err)
	}
	return openAPIRouter
}

func renderOpenAPITemplate(config *gomagpieconfig.Config, fileName string, params any) []byte {
	file := filepath.Clean(fileName)
	files := []string{problems, file} // add problems template too since it's an "include" template
	parsed := texttemplate.Must(texttemplate.New(filepath.Base(file)).Funcs(GlobalTemplateFuncs).ParseFiles(files...))

	var rendered bytes.Buffer
	if err := parsed.Execute(&rendered, &TemplateData{Config: config, Params: params}); err != nil {
		log.Fatalf("failed to render %s, error: %v", file, err)
	}
	return rendered.Bytes()
}

func (o *OpenAPI) ValidateRequest(r *http.Request) error {
	requestValidationInput, _ := o.getRequestValidationInput(r)
	if requestValidationInput != nil {
		err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput)
		if err != nil {
			var schemaErr *openapi3.SchemaError
			// Don't fail on maximum constraints because OGC has decided these are soft limits, for instance
			// in features: "If the value of the limit parameter is larger than the maximum value, this
			// SHALL NOT result in an error (instead use the maximum as the parameter value)."
			if errors.As(err, &schemaErr) && schemaErr.SchemaField == "maximum" {
				return nil
			}
			return fmt.Errorf("request doesn't conform to OpenAPI spec: %w", err)
		}
	}
	return nil
}

func (o *OpenAPI) ValidateResponse(contentType string, body []byte, r *http.Request) error {
	requestValidationInput, _ := o.getRequestValidationInput(r)
	if requestValidationInput != nil {
		responseHeaders := http.Header{HeaderContentType: []string{contentType}}
		responseCode := 200

		responseValidationInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: requestValidationInput,
			Status:                 responseCode,
			Header:                 responseHeaders,
		}
		responseValidationInput.SetBodyBytes(body)
		err := openapi3filter.ValidateResponse(context.Background(), responseValidationInput)
		if err != nil {
			return fmt.Errorf("response doesn't conform to OpenAPI spec: %w", err)
		}
	}
	return nil
}

func (o *OpenAPI) getRequestValidationInput(r *http.Request) (*openapi3filter.RequestValidationInput, error) {
	route, pathParams, err := o.router.FindRoute(r)
	if err != nil {
		log.Printf("route not found in OpenAPI spec for url %s (host: %s), "+
			"skipping OpenAPI validation", r.URL, r.Host)
		return nil, err
	}
	opts := &openapi3filter.Options{
		SkipSettingDefaults: true,
	}
	opts.WithCustomSchemaErrorFunc(func(err *openapi3.SchemaError) string {
		return err.Reason
	})
	return &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
		Options:    opts,
	}, nil
}

// normalizeBaseURL normalizes the given base URL so our OpenAPI validator is able to match
// requests against the OpenAPI spec. This involves:
//
//   - striping the context root (path) from the base URL. If you use a context root we expect
//     you to have a proxy fronting Gomagpie, therefore we also  need to strip it from the base
//     URL used during OpenAPI validation
//
//   - replacing HTTPS scheme with HTTP. Since Gomagpie doesn't support HTTPS we always perform
//     OpenAPI validation against HTTP requests. Note: it's possible to offer Gomagpie over HTTPS, but you'll
//     need to take care of that in your proxy server (or loadbalancer/service mesh/etc) fronting Gomagpie.
func normalizeBaseURL(baseURL string) string {
	serverURL, _ := url.Parse(baseURL)
	result := strings.Replace(baseURL, serverURL.Scheme, "http", 1)
	result = strings.Replace(result, serverURL.Path, "", 1)
	return result
}
