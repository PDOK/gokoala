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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	"github.com/PDOK/gokoala/engine/util"
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
)

var (
	globalTemplateFuncs texttemplate.FuncMap
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
	// Config set during startup based on the given config file
	Config *Config

	// Params optional parameters not part of GoKoala's config file. You can use
	// this to provide extra data to a template at rendering time.
	Params interface{}

	// Breadcrumb path to the page, in key-value pairs of name->path
	Breadcrumbs []Breadcrumb

	// Request URL
	url *url.URL
}

// AvailableFormats returns the output formats available for the current page
func (td *TemplateData) AvailableFormats() map[string]string {
	if td.url != nil && strings.Contains(td.url.Path, "/items") {
		return td.AvailableFormatsFeatures()
	}
	return OutputFormatDefault
}

// AvailableFormatsFeatures convenience function
func (td *TemplateData) AvailableFormatsFeatures() map[string]string {
	return OutputFormatFeatures
}

// QueryString returns ?=foo=a&bar=b style query string of the current page
func (td *TemplateData) QueryString(format string) string {
	if td.url != nil {
		q := td.url.Query()
		if format != "" {
			q.Set(FormatParam, format)
		}
		return "?" + q.Encode()
	}
	return fmt.Sprintf("?%s=%s", FormatParam, format)
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

func ExpandTemplateKey(key TemplateKey, language language.Tag) TemplateKey {
	copyKey := key
	copyKey.Language = language
	return copyKey
}

type Templates struct {
	// ParsedTemplates templates loaded from disk and parsed to an in-memory Go representation.
	ParsedTemplates map[TemplateKey]interface{}

	// RenderedTemplates templates parsed + rendered to their actual output format like JSON, HTMl, etc.
	// We prefer pre-rendered templates whenever possible. These are stored in this map.
	RenderedTemplates map[TemplateKey][]byte

	config     *Config
	localizers map[language.Tag]i18n.Localizer
}

func newTemplates(config *Config) *Templates {
	templates := &Templates{
		ParsedTemplates:   make(map[TemplateKey]interface{}),
		RenderedTemplates: make(map[TemplateKey][]byte),
		config:            config,
		localizers:        newLocalizers(config.AvailableLanguages),
	}
	customFuncs := texttemplate.FuncMap{
		// custom template functions
		"markdown":   markdown,
		"unmarkdown": unmarkdown,
	}
	// we also support https://github.com/go-task/slim-sprig functions
	sprigFuncs := sprig.FuncMap()
	globalTemplateFuncs = combineFuncMaps(customFuncs, sprigFuncs)
	return templates
}

func (t *Templates) getParsedTemplate(key TemplateKey) (interface{}, error) {
	if parsedTemplate, ok := t.ParsedTemplates[key]; ok {
		return parsedTemplate, nil
	}
	return nil, fmt.Errorf("no parsed template with name %s", key.Name)
}

func (t *Templates) getRenderedTemplate(key TemplateKey) ([]byte, error) {
	if RenderedTemplate, ok := t.RenderedTemplates[key]; ok {
		return RenderedTemplate, nil
	}
	return nil, fmt.Errorf("no rendered template with name %s", key.Name)
}

func (t *Templates) parseAndSaveTemplate(key TemplateKey) {
	for lang := range t.localizers {
		keyWithLang := ExpandTemplateKey(key, lang)
		if key.Format == FormatHTML {
			_, parsed := t.parseHTMLTemplate(keyWithLang, lang)
			t.ParsedTemplates[keyWithLang] = parsed
		} else {
			_, parsed := t.parseNonHTMLTemplate(keyWithLang, lang)
			t.ParsedTemplates[keyWithLang] = parsed
		}
	}
}

func (t *Templates) renderAndSaveTemplate(key TemplateKey, breadcrumbs []Breadcrumb, params interface{}) {
	for lang := range t.localizers {
		var result []byte
		if key.Format == FormatHTML {
			file, parsed := t.parseHTMLTemplate(key, lang)
			result = t.renderHTMLTemplate(parsed, nil, params, breadcrumbs, file)
		} else {
			file, parsed := t.parseNonHTMLTemplate(key, lang)
			result = t.renderNonHTMLTemplate(parsed, params, key, file)
		}

		// Store rendered template per language
		key.Language = lang
		t.RenderedTemplates[key] = result
	}
}

func (t *Templates) parseHTMLTemplate(key TemplateKey, lang language.Tag) (string, *htmltemplate.Template) {
	file := filepath.Clean(filepath.Join(key.Directory, key.Name))
	templateFuncs := t.createTemplateFuncs(lang)
	parsed := htmltemplate.Must(htmltemplate.New(layoutFile).
		Funcs(templateFuncs).ParseFiles(templatesDir+layoutFile, file))
	return file, parsed
}

func (t *Templates) renderHTMLTemplate(parsed *htmltemplate.Template, URL *url.URL,
	params interface{}, breadcrumbs []Breadcrumb, file string) []byte {

	var rendered bytes.Buffer
	if err := parsed.Execute(&rendered, &TemplateData{
		Config:      t.config,
		Params:      params,
		Breadcrumbs: breadcrumbs,
		url:         URL,
	}); err != nil {
		log.Fatalf("failed to execute HTML template %s, error: %v", file, err)
	}
	return rendered.Bytes()
}

func (t *Templates) parseNonHTMLTemplate(key TemplateKey, lang language.Tag) (string, *texttemplate.Template) {
	file := filepath.Clean(filepath.Join(key.Directory, key.Name))
	templateFuncs := t.createTemplateFuncs(lang)
	parsed := texttemplate.Must(texttemplate.New(filepath.Base(file)).
		Funcs(templateFuncs).Parse(t.readFile(file)))
	return file, parsed
}

func (t *Templates) renderNonHTMLTemplate(parsed *texttemplate.Template, params interface{}, key TemplateKey, file string) []byte {
	var rendered bytes.Buffer
	if err := parsed.Execute(&rendered, &TemplateData{
		Config: t.config,
		Params: params,
	}); err != nil {
		log.Fatalf("failed to execute template %s, error: %v", file, err)
	}

	var result = rendered.Bytes()
	if strings.Contains(key.Format, FormatJSON) {
		// pretty print all JSON (or derivatives like TileJSON)
		result = util.PrettyPrintJSON(result, key.Name)
	}
	return result
}

func (t *Templates) createTemplateFuncs(lang language.Tag) map[string]interface{} {
	return combineFuncMaps(globalTemplateFuncs, texttemplate.FuncMap{
		// create func just-in-time based on TemplateKey
		"i18n": func(messageID string) htmltemplate.HTML {
			localizer := t.localizers[lang]
			translated := localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID})
			return htmltemplate.HTML(translated) //nolint:gosec // since we trust our language files
		},
	})
}

// read file, return contents as string
func (t *Templates) readFile(filePath string) string {
	gzipFile := filePath + ".gz"
	var fileContents string
	if _, err := os.Stat(gzipFile); !errors.Is(err, fs.ErrNotExist) {
		fileContents, err = readGzipContents(gzipFile)
		if err != nil {
			log.Fatalf("unable to decompress gzip file %s", gzipFile)
		}
	} else {
		fileContents, err = readPlainContents(filePath)
		if err != nil {
			log.Fatalf("unable to read file %s", filePath)
		}
	}
	return fileContents
}

// decompress gzip files, return contents as string
func readGzipContents(filePath string) (string, error) {
	gzipFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(gzipFile *os.File) {
		err := gzipFile.Close()
		if err != nil {
			log.Println("failed to close gzip file")
		}
	}(gzipFile)
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return "", err
	}
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		if err != nil {
			log.Println("failed to close gzip reader")
		}
	}(gzipReader)
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, gzipReader) //nolint:gosec
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// read file, return contents as string
func readPlainContents(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("failed to close file")
		}
	}(file)
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, file)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// combine given FuncMaps
func combineFuncMaps(funcMaps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, funcMap := range funcMaps {
		for k, v := range funcMap {
			result[k] = v
		}
	}
	return result
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
