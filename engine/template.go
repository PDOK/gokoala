package engine

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	sprig "github.com/go-task/slim-sprig"
	gomarkdown "github.com/gomarkdown/markdown"
	gomarkdownhtml "github.com/gomarkdown/markdown/html"
	gomarkdownparser "github.com/gomarkdown/markdown/parser"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	stripmd "github.com/writeas/go-strip-markdown/v2"
	"golang.org/x/text/language"
)

const (
	layoutFile = "layout.go.html"
	FormatHTML = "html"
	FormatJSON = "json"
)

var (
	customFuncs   texttemplate.FuncMap
	sprigFuncs    texttemplate.FuncMap
	combinedFuncs texttemplate.FuncMap
)

// TemplateKey unique key to register and lookup Go templates
type TemplateKey struct {
	// Name of the template, the filename including extension
	Name string

	// Directory in which the template resides
	Directory string

	// Format the file format based on the filename extension, 'html' or 'json'
	Format string

	// Language of the contents of the template
	Language language.Tag

	// Optional. Only required when you want to render the same template multiple times (with different content).
	// By specifying an 'instance name' you can refer to a certain instance of a rendered template later on.
	InstanceName string
}

// TemplateData the data/variables passed as an argument into the template.
type TemplateData struct {
	Config *Config

	// Params optional parameters not part of GoKoala's configfile. You can use
	// this to provide extra data to a template at rendering time.
	Params interface{}

	// Crumb path to the page, in key-value pairs of name, path
	Breadcrumbs []Breadcrumb

	// Language to render the page in
	Language language.Tag
}

type Breadcrumb struct {
	Name string
	Path string
}

// NewTemplateKey build TemplateKeys
func NewTemplateKey(path string) TemplateKey {
	return NewTemplateKeyWithName(path, "")
}

func NewTemplateKeyWithLanguage(path string, language language.Tag) TemplateKey {
	return NewTemplateKeyWithNameAndLanguage(path, "", language)
}

// NewTemplateKeyWithName build TemplateKey with InstanceName (see docs in struct)
func NewTemplateKeyWithName(path string, instanceName string) TemplateKey {
	return NewTemplateKeyWithNameAndLanguage(path, instanceName, language.Dutch)
}

func NewTemplateKeyWithNameAndLanguage(path string, instanceName string, language language.Tag) TemplateKey {
	cleanPath := filepath.Clean(path)
	return TemplateKey{
		Name:         filepath.Base(cleanPath),
		Directory:    filepath.Dir(cleanPath),
		Format:       strings.TrimPrefix(filepath.Ext(path), "."),
		Language:     language,
		InstanceName: instanceName,
	}
}

type Templates struct {
	RenderedTemplates map[TemplateKey][]byte
	config            *Config
	localizers        map[language.Tag]i18n.Localizer
}

func newTemplates(config *Config, localizers map[language.Tag]i18n.Localizer) *Templates {
	templates := &Templates{
		RenderedTemplates: make(map[TemplateKey][]byte),
		config:            config,
		localizers:        localizers,
	}
	customFuncs = texttemplate.FuncMap{
		// custom template functions
		"markdown":     markdown,
		"unmarkdown":   unmarkdown,
		"i18n":         templates.i18n,
		"unescapeHTML": unescapeHTML,
	}
	sprigFuncs = sprig.FuncMap()
	combinedFuncs = combinedFuncMap(customFuncs, sprigFuncs)
	return templates
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
		Language:    key.Language,
	}); err != nil {
		log.Fatalf("failed to execute HTML template %s, error: %v", file, err)
	}

	t.RenderedTemplates[key] = rendered.Bytes()
}

func (t *Templates) renderNonHTMLTemplate(key TemplateKey, params interface{}) {
	file := filepath.Clean(filepath.Join(key.Directory, key.Name))
	gzipFile := file + ".gz"
	var fileContents string
	if _, err := os.Stat(gzipFile); !errors.Is(err, fs.ErrNotExist) {
		fileContents, err = readGzipContents(gzipFile)
		if err != nil {
			log.Fatalf("unable to decompress gzip file %s", gzipFile)
		}
	} else {
		fileContents, err = readFileContents(file)
		if err != nil {
			log.Fatalf("unable to read file %s", file)
		}
	}
	compiled := texttemplate.Must(texttemplate.New(filepath.Base(file)).Funcs(combinedFuncs).Parse(fileContents))
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

// decompress gzip files, return contents as string
func readGzipContents(filePath string) (string, error) {
	gzipFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer gzipFile.Close()
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return "", err
	}
	defer gzipReader.Close()
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, gzipReader) //nolint:gosec
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// read file, return contents as string
func readFileContents(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, file)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
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

func unescapeHTML(s string) htmltemplate.HTML {
	return htmltemplate.HTML(s) //nolint:gosec
}

func (t *Templates) i18n(messageID string, lang language.Tag) string {
	localizer := t.localizers[lang]
	return localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID})
}
