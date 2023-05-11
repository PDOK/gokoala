package engine

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"log"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	sprig "github.com/go-task/slim-sprig"
	gomarkdown "github.com/gomarkdown/markdown"
	gomarkdownhtml "github.com/gomarkdown/markdown/html"
	gomarkdownparser "github.com/gomarkdown/markdown/parser"
	stripmd "github.com/writeas/go-strip-markdown/v2"
)

const (
	layoutFile = "layout.go.html"
	FormatHTML = "html"
	FormatJSON = "json"
)

var (
	customFuncs = texttemplate.FuncMap{
		// custom template functions
		"markdown":   markdown,
		"unmarkdown": unmarkdown,
	}
	sprigFuncs    = sprig.FuncMap()
	combinedFuncs = combinedFuncMap(customFuncs, sprigFuncs)
)

// TemplateKey unique key to register and lookup Go templates
type TemplateKey struct {
	// Name of the template, the filename including extension
	Name string

	// Directory in which the template resides
	Directory string

	// Format the file format based on the filename extension, 'html' or 'json'
	Format string

	// Optional. Only required when you want to render the same template multiple times (with different content).
	// By specifying an 'instance name' you can refer to a certain instance of a rendered template later on.
	InstanceName string
}

// TemplateData the data/variables passed as an argument into the template.
type TemplateData struct {
	Config *Config

	// Params optional parameters not part of GoKoala's configfile
	Params interface{}

	// Crumb path to the page, in key-value pairs of name, path
	Breadcrumbs []Breadcrumb
}

type Breadcrumb struct {
	Name string
	Path string
}

// NewTemplateKey build TemplateKeys
func NewTemplateKey(path string) TemplateKey {
	return NewTemplateKeyWithName(path, "")
}

// NewTemplateKeyWithName build TemplateKey with InstanceName (see docs in struct)
func NewTemplateKeyWithName(path string, instanceName string) TemplateKey {
	cleanPath := filepath.Clean(path)
	return TemplateKey{
		Name:         filepath.Base(cleanPath),
		Directory:    filepath.Dir(cleanPath),
		Format:       strings.TrimPrefix(filepath.Ext(path), "."),
		InstanceName: instanceName,
	}
}

type Templates struct {
	RenderedTemplates map[TemplateKey][]byte
	config            *Config
}

func newTemplates(config *Config) *Templates {
	return &Templates{
		RenderedTemplates: make(map[TemplateKey][]byte),
		config:            config,
	}
}

// GetRenderedTemplate returns a pre-rendered template, or error if none is found for the given TemplateKey
func (t *Templates) GetRenderedTemplate(key TemplateKey) ([]byte, error) {
	if renderedTemplate, ok := t.RenderedTemplates[key]; ok {
		return renderedTemplate, nil
	}
	return nil, fmt.Errorf("no rendered template with name %s", key.Name)
}

func (t *Templates) renderHTMLTemplate(key TemplateKey, breadcrumbs []Breadcrumb, params interface{}) {
	file := filepath.Clean(filepath.Join(key.Directory, key.Name))
	compiled := htmltemplate.Must(htmltemplate.New(layoutFile).Funcs(combinedFuncs).ParseFiles(templatesDir+layoutFile, file))
	var rendered bytes.Buffer

	if err := compiled.Execute(&rendered, &TemplateData{
		Config:      t.config,
		Params:      params,
		Breadcrumbs: breadcrumbs,
	}); err != nil {
		log.Fatalf("failed to execute HTML template %s, error: %v", file, err)
	}

	t.RenderedTemplates[key] = rendered.Bytes()
}

func (t *Templates) renderNonHTMLTemplate(key TemplateKey, params interface{}) {
	file := filepath.Clean(filepath.Join(key.Directory, key.Name))
	compiled := texttemplate.Must(texttemplate.New(filepath.Base(file)).Funcs(combinedFuncs).ParseFiles(file))
	var rendered bytes.Buffer

	if err := compiled.Execute(&rendered, &TemplateData{
		Config: t.config,
		Params: params,
	}); err != nil {
		log.Fatalf("failed to execute template %s, error: %v", file, err)
	}

	var result = rendered.Bytes()
	if strings.Contains(key.Format, FormatJSON) {
		// pretty print all JSON (or derivatives like TileJSON)
		result = PrettyPrintJSON(result, key.Name)
	}
	t.RenderedTemplates[key] = result
}

// combine applicable FuncMaps
func combinedFuncMap(customFuncs map[string]interface{}, sprigFuncs map[string]interface{}) map[string]interface{} {
	cfm := make(map[string]interface{}, len(customFuncs)+len(sprigFuncs))
	for k, v := range sprigFuncs {
		cfm[k] = v
	}
	for k, v := range customFuncs {
		cfm[k] = v
	}
	return cfm
}

// markdown turn Markdown into HTML
func markdown(s *string) htmltemplate.HTML {
	if s == nil {
		return ""
	}
	// always normalize newlines, this library only supports Unix LF newlines
	md := gomarkdown.NormalizeNewlines([]byte(*s))

	// create Markdown parser
	extensions := gomarkdownparser.CommonExtensions
	parser := gomarkdownparser.NewWithExtensions(extensions)

	// parse Markdown into AST tree
	doc := parser.Parse(md)

	// create HTML renderer
	htmlFlags := gomarkdownhtml.CommonFlags | gomarkdownhtml.HrefTargetBlank | gomarkdownhtml.SkipHTML
	renderer := gomarkdownhtml.NewRenderer(gomarkdownhtml.RendererOptions{Flags: htmlFlags})

	return htmltemplate.HTML(gomarkdown.Render(doc, renderer)) //nolint:gosec
}

// unmarkdown remove Markdown, so we can use the given string in non-HTML (JSON) output
func unmarkdown(s *string) string {
	if s == nil {
		return ""
	}
	withoutMarkdown := stripmd.Strip(*s)
	withoutLinebreaks := strings.ReplaceAll(withoutMarkdown, "\n", " ")
	return withoutLinebreaks
}
