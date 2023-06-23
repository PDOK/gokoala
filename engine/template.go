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
	Config *Config

	// Params optional parameters not part of GoKoala's configfile. You can use
	// this to provide extra data to a template at rendering time.
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

func newTemplates(config *Config) *Templates {
	templates := &Templates{
		RenderedTemplates: make(map[TemplateKey][]byte),
		config:            config,
		localizers:        NewLocalizers(config.AvailableLanguages),
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

// GetRenderedTemplate returns a pre-rendered template, or error if none is found for the given TemplateKey
func (t *Templates) GetRenderedTemplate(key TemplateKey) ([]byte, error) {
	if renderedTemplate, ok := t.RenderedTemplates[key]; ok {
		return renderedTemplate, nil
	}
	return nil, fmt.Errorf("no rendered template with name %s", key.Name)
}

func (t *Templates) renderHTMLTemplate(key TemplateKey, breadcrumbs []Breadcrumb, params interface{}) {
	file := filepath.Clean(filepath.Join(key.Directory, key.Name))

	for lang := range t.localizers {
		templateFuncs := t.createTemplateFuncs(lang)
		compiled := htmltemplate.Must(htmltemplate.New(layoutFile).
			Funcs(templateFuncs).ParseFiles(templatesDir+layoutFile, file))

		var rendered bytes.Buffer
		if err := compiled.Execute(&rendered, &TemplateData{
			Config:      t.config,
			Params:      params,
			Breadcrumbs: breadcrumbs,
		}); err != nil {
			log.Fatalf("failed to execute HTML template %s, error: %v", file, err)
		}

		// Store rendered template per language
		key.Language = lang
		t.RenderedTemplates[key] = rendered.Bytes()
	}
}

func (t *Templates) renderNonHTMLTemplate(key TemplateKey, params interface{}) {
	file := filepath.Clean(filepath.Join(key.Directory, key.Name))

	for lang := range t.localizers {
		templateFuncs := t.createTemplateFuncs(lang)
		compiled := texttemplate.Must(texttemplate.New(filepath.Base(file)).
			Funcs(templateFuncs).Parse(t.readFile(file)))

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

		// Store rendered template per language
		key.Language = lang
		t.RenderedTemplates[key] = result
	}
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
func readPlainContents(filePath string) (string, error) {
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
