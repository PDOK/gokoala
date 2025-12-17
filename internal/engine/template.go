package engine

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const (
	layoutFile = "layout.go.html"
)

// TemplateKey unique key to register and lookup Go templates.
type TemplateKey struct {
	// Name of the template, the filename including extension
	Name string

	// Directory in which the template resides
	Directory string

	// Format the file format based on the filename extension, 'html' or 'json'
	Format string

	// Optional. Use with caution, overwrite the Media-Type associated with the Format of this template.
	MediaTypeOverwrite string

	// Language of the contents of the template
	Language language.Tag

	// Optional. Only required when you want to render the same template multiple times (with different content).
	// By specifying an 'instance name' you can refer to a certain instance of a rendered template later on.
	InstanceName string
}

// TemplateKeyOption implements the functional option pattern for TemplateKey.
type TemplateKeyOption func(*TemplateKey)

// TemplateData the data/variables passed as an argument into the template.
type TemplateData struct {
	// Config set during startup based on the given config file
	Config *config.Config

	// Theme set during startup
	Theme *config.Theme

	// Params optional parameters not part of GoKoala's config file. You can use
	// this to provide extra data to a template at rendering time.
	Params any

	// Breadcrumb path to the page, in key-value pairs of name->path
	Breadcrumbs []Breadcrumb

	// AvailableFormats returns the output formats available for the current page
	AvailableFormats []OutputFormat

	// Request URL
	url *url.URL
}

// QueryString returns ?=foo=a&bar=b style query string of the current page.
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

// WithLanguage sets the language of a TemplateKey.
func WithLanguage(language language.Tag) TemplateKeyOption {
	return func(tk *TemplateKey) {
		tk.Language = language
	}
}

// WithNegotiatedLanguage sets the language of a TemplateKey based on content-negotiation.
func (e *Engine) WithNegotiatedLanguage(w http.ResponseWriter, r *http.Request) TemplateKeyOption {
	return WithLanguage(e.CN.NegotiateLanguage(w, r))
}

// WithInstanceName sets the instance name of a TemplateKey.
func WithInstanceName(instanceName string) TemplateKeyOption {
	return func(tk *TemplateKey) {
		tk.InstanceName = instanceName
	}
}

// WithMediaTypeOverwrite overwrites the mediatype of the Format in a TemplateKey.
func WithMediaTypeOverwrite(mediaType string) TemplateKeyOption {
	return func(tk *TemplateKey) {
		tk.MediaTypeOverwrite = mediaType
	}
}

// NewTemplateKey builds a TemplateKey with the given path and options.
func NewTemplateKey(path string, opts ...TemplateKeyOption) TemplateKey {
	cleanPath := filepath.Clean(path)
	tk := TemplateKey{
		Name:      filepath.Base(cleanPath),
		Directory: filepath.Dir(cleanPath),
		Format:    strings.TrimPrefix(filepath.Ext(path), "."),
		Language:  language.Dutch, // Default language
	}

	for _, opt := range opts {
		opt(&tk)
	}

	return tk
}

func ExpandTemplateKey(key TemplateKey, language language.Tag) TemplateKey {
	copyKey := key
	copyKey.Language = language

	return copyKey
}

type Templates struct {
	// ParsedTemplates templates loaded from disk and parsed to an in-memory Go representation.
	ParsedTemplates map[TemplateKey]any

	// RenderedTemplates templates parsed + rendered to their actual output format like JSON, HTMl, etc.
	// We prefer pre-rendered templates whenever possible. These are stored in this map.
	RenderedTemplates map[TemplateKey][]byte
	Theme             *config.Theme

	config     *config.Config
	localizers map[language.Tag]i18n.Localizer
}

func newTemplates(config *config.Config, theme *config.Theme) *Templates {
	templates := &Templates{
		ParsedTemplates:   make(map[TemplateKey]any),
		RenderedTemplates: make(map[TemplateKey][]byte),
		config:            config,
		Theme:             theme,
		localizers:        newLocalizers(config.AvailableLanguages),
	}

	return templates
}

func (t *Templates) getParsedTemplate(key TemplateKey) (any, error) {
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

func (t *Templates) renderAndSaveTemplate(key TemplateKey, breadcrumbs []Breadcrumb, params any) {
	for lang := range t.localizers {
		var result []byte
		if key.Format == FormatHTML {
			file, parsed := t.parseHTMLTemplate(key, lang)
			result = t.renderHTMLTemplate(parsed, nil, params, breadcrumbs, file, OutputFormatDefault)
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
func (t *Templates) renderHTMLTemplate(parsed *htmltemplate.Template, url *url.URL,
	params any, breadcrumbs []Breadcrumb, file string, availableFormats []OutputFormat) []byte {

	var rendered bytes.Buffer
	if err := parsed.Execute(&rendered, &TemplateData{
		Config:           t.config,
		Theme:            t.Theme,
		Params:           params,
		Breadcrumbs:      breadcrumbs,
		AvailableFormats: availableFormats,
		url:              url,
	}); err != nil {
		log.Fatalf("failed to execute HTML template %s, error: %v", file, err)
	}

	return rendered.Bytes()
}

func (t *Templates) parseNonHTMLTemplate(key TemplateKey, lang language.Tag) (string, *texttemplate.Template) {
	file := filepath.Clean(filepath.Join(key.Directory, key.Name))
	templateFuncs := t.createTemplateFuncs(lang)
	parsed := texttemplate.Must(texttemplate.New(filepath.Base(file)).
		Funcs(templateFuncs).Parse(util.ReadFile(file)))

	return file, parsed
}

func (t *Templates) renderNonHTMLTemplate(parsed *texttemplate.Template, params any, key TemplateKey, file string) []byte {
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

func (t *Templates) createTemplateFuncs(lang language.Tag) map[string]any {
	return combineFuncMaps(GlobalTemplateFuncs, texttemplate.FuncMap{
		// create func just-in-time based on TemplateKey
		"i18n": func(messageID string) htmltemplate.HTML {
			localizer := t.localizers[lang]
			translated := localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID})

			return htmltemplate.HTML(translated) //nolint:gosec // since we trust our language files
		},
	})
}
