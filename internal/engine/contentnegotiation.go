package engine

import (
	"log"
	"net/http"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/elnormous/contenttype"
	"golang.org/x/text/language"
)

const (
	FormatParam   = "f"
	languageParam = "lang"

	MediaTypeJSON    = "application/json"
	MediaTypeXML     = "application/xml"
	MediaTypeHTML    = "text/html"
	MediaTypeOpenAPI = "application/vnd.oai.openapi+json;version=3.0"
	MediaTypeGeoJSON = "application/geo+json"
	MediaTypeJSONFG  = "application/vnd.ogc.fg+json" // https://docs.ogc.org/per/21-017r1.html#toc17

	FormatHTML    = "html"
	FormatXML     = "xml"
	FormatJSON    = "json"
	FormatGeoJSON = "geojson" // ?=json should also work for geojson
	FormatJSONFG  = "jsonfg"
	FormatGzip    = "gzip"
)

var (
	MediaTypeJSONFamily    = []string{MediaTypeGeoJSON, MediaTypeJSONFG}
	OutputFormatDefault    = map[string]string{FormatJSON: "JSON"}
	OutputFormatFeatures   = map[string]string{FormatJSON: "GeoJSON", FormatJSONFG: "JSON-FG"}
	CompressibleMediaTypes = []string{
		MediaTypeJSON,
		MediaTypeGeoJSON,
		MediaTypeJSONFG,
		MediaTypeOpenAPI,
		MediaTypeHTML,
		// common web media types
		"text/css",
		"text/plain",
		"text/javascript",
		"application/javascript",
		"image/svg+xml",
	}
)

type ContentNegotiation struct {
	availableMediaTypes []contenttype.MediaType
	availableLanguages  []language.Tag

	formatsByMediaType map[string]string
	mediaTypesByFormat map[string]string
}

func newContentNegotiation(availableLanguages []config.Language) *ContentNegotiation {
	availableMediaTypes := []contenttype.MediaType{
		// in order
		contenttype.NewMediaType(MediaTypeJSON),
		contenttype.NewMediaType(MediaTypeXML),
		contenttype.NewMediaType(MediaTypeHTML),
		contenttype.NewMediaType(MediaTypeGeoJSON),
		contenttype.NewMediaType(MediaTypeJSONFG),
	}

	formatsByMediaType := map[string]string{
		MediaTypeJSON:    FormatJSON,
		MediaTypeXML:     FormatXML,
		MediaTypeHTML:    FormatHTML,
		MediaTypeGeoJSON: FormatGeoJSON,
		MediaTypeJSONFG:  FormatJSONFG,
	}

	mediaTypesByFormat := util.Inverse(formatsByMediaType)

	languageTags := make([]language.Tag, 0, len(availableLanguages))
	for _, availableLanguage := range availableLanguages {
		languageTags = append(languageTags, availableLanguage.Tag)
	}

	return &ContentNegotiation{
		availableMediaTypes: availableMediaTypes,
		availableLanguages:  languageTags,
		formatsByMediaType:  formatsByMediaType,
		mediaTypesByFormat:  mediaTypesByFormat,
	}
}

// NegotiateFormat performs content negotiation, not idempotent (since it removes the ?f= param)
func (cn *ContentNegotiation) NegotiateFormat(req *http.Request) string {
	requestedFormat := cn.getFormatFromQueryParam(req)
	if requestedFormat == "" {
		requestedFormat = cn.getFormatFromAcceptHeader(req)
	}
	if requestedFormat == "" {
		requestedFormat = FormatJSON // default
	}
	return requestedFormat
}

// NegotiateLanguage performs language negotiation, not idempotent (since it removes the ?lang= param)
func (cn *ContentNegotiation) NegotiateLanguage(w http.ResponseWriter, req *http.Request) language.Tag {
	requestedLanguage := cn.getLanguageFromQueryParam(w, req)
	if requestedLanguage == language.Und {
		requestedLanguage = cn.getLanguageFromCookie(req)
	}
	if requestedLanguage == language.Und {
		requestedLanguage = cn.getLanguageFromHeader(req)
	}
	if requestedLanguage == language.Und {
		requestedLanguage = language.Dutch // default
	}
	return requestedLanguage
}

func (cn *ContentNegotiation) formatToMediaType(format string) string {
	return cn.mediaTypesByFormat[format]
}

func (cn *ContentNegotiation) getFormatFromQueryParam(req *http.Request) string {
	var requestedFormat = ""
	queryParams := req.URL.Query()
	if queryParams.Get(FormatParam) != "" {
		requestedFormat = queryParams.Get(FormatParam)

		// remove ?f= parameter, to prepare for rewrite
		queryParams.Del(FormatParam)
		req.URL.RawQuery = queryParams.Encode()
	}
	return requestedFormat
}

func (cn *ContentNegotiation) getFormatFromAcceptHeader(req *http.Request) string {
	accepted, _, err := contenttype.GetAcceptableMediaType(req, cn.availableMediaTypes)
	if err != nil {
		log.Printf("Failed to parse Accept header: %v. Continuing\n", err)
		return ""
	}
	return cn.formatsByMediaType[accepted.String()]
}

func (cn *ContentNegotiation) getLanguageFromQueryParam(w http.ResponseWriter, req *http.Request) language.Tag {
	var requestedLanguage = language.Und
	queryParams := req.URL.Query()
	if queryParams.Get(languageParam) != "" {
		lang := queryParams.Get(languageParam)
		accepted, _, err := language.ParseAcceptLanguage(lang)
		if err != nil {
			return requestedLanguage
		}
		m := language.NewMatcher(cn.availableLanguages)
		_, langIndex, _ := m.Match(accepted...)
		requestedLanguage = cn.availableLanguages[langIndex]
		// override for use in cookie
		lang = requestedLanguage.String()

		// set requested language in cookie
		setLanguageCookie(w, lang)

		// remove ?lang= parameter, to prepare for rewrite
		queryParams.Del(languageParam)
		req.URL.RawQuery = queryParams.Encode()
	}
	return requestedLanguage
}

func setLanguageCookie(w http.ResponseWriter, lang string) {
	cookie := &http.Cookie{
		Name:     languageParam,
		Value:    lang,
		Path:     "/",
		MaxAge:   config.CookieMaxAge,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	}
	http.SetCookie(w, cookie)
}

func (cn *ContentNegotiation) getLanguageFromCookie(req *http.Request) language.Tag {
	var requestedLanguage = language.Und
	cookie, err := req.Cookie(languageParam)
	if err != nil {
		return requestedLanguage
	}
	lang := cookie.Value
	accepted, _, err := language.ParseAcceptLanguage(lang)
	if err != nil {
		return requestedLanguage
	}
	m := language.NewMatcher(cn.availableLanguages)
	_, langIndex, _ := m.Match(accepted...)
	requestedLanguage = cn.availableLanguages[langIndex]
	return requestedLanguage
}

func (cn *ContentNegotiation) getLanguageFromHeader(req *http.Request) language.Tag {
	var requestedLanguage = language.Und
	if req.Header.Get(HeaderAcceptLanguage) != "" {
		accepted, _, err := language.ParseAcceptLanguage(req.Header.Get(HeaderAcceptLanguage))
		if err != nil {
			log.Printf("Failed to parse Accept-Language header: %v. Continuing\n", err)
			return requestedLanguage
		}
		m := language.NewMatcher(cn.availableLanguages)
		_, langIndex, _ := m.Match(accepted...)
		requestedLanguage = cn.availableLanguages[langIndex]
	}
	return requestedLanguage
}
