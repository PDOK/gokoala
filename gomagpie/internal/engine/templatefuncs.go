package engine

import (
	htmltemplate "html/template"
	"log"
	"regexp"
	"strconv"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/docker/go-units"

	sprig "github.com/go-task/slim-sprig"
	gomarkdown "github.com/gomarkdown/markdown"
	gomarkdownhtml "github.com/gomarkdown/markdown/html"
	gomarkdownparser "github.com/gomarkdown/markdown/parser"
	stripmd "github.com/writeas/go-strip-markdown/v2"
)

var (
	GlobalTemplateFuncs texttemplate.FuncMap
	linkRegex           = regexp.MustCompile(`^https?://\S+$`)
)

// Initialize functions to be used in html/json/etc templates
func init() {
	customFuncs := texttemplate.FuncMap{
		// custom template functions
		"markdown":   markdown,
		"unmarkdown": unmarkdown,
		"humansize":  humanSize,
		"bytessize":  bytesSize,
		"isdate":     isDate,
		"islink":     isLink,
		"firstupper": firstUpper,
	}
	sprigFuncs := sprig.FuncMap() // we also support https://github.com/go-task/slim-sprig functions
	GlobalTemplateFuncs = combineFuncMaps(customFuncs, sprigFuncs)
}

// combine given FuncMaps
func combineFuncMaps(funcMaps ...map[string]any) map[string]any {
	result := make(map[string]any)
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

// humanSize converts size in bytes to a human-readable size
func humanSize(a any) string {
	if i, ok := a.(int64); ok {
		return units.HumanSize(float64(i))
	} else if f, ok := a.(float64); ok {
		return units.HumanSize(f)
	} else if s, ok := a.(string); ok {
		fs, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return units.HumanSize(fs)
		}
	}
	log.Printf("cannot convert '%v' to float", a)
	return "0"
}

// bytesSize converts human-readable size to size in bytes (base-10, not base-2)
func bytesSize(s string) int64 {
	i, err := units.FromHumanSize(s)
	if err != nil {
		log.Printf("cannot convert '%s' to bytes", s)
		return 0
	}
	return i
}

// isDate true when given input is a date, false otherwise
func isDate(v any) bool {
	if _, ok := v.(time.Time); ok {
		return true
	}
	return false
}

// isLink true when given input is an HTTP(s) URL (without any additional text), false otherwise
func isLink(v any) bool {
	if text, ok := v.(string); ok {
		return linkRegex.MatchString(text)
	}
	return false
}

func firstUpper(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}
