package engine

import (
	"log"
	"net/http"

	"github.com/elnormous/contenttype"
	"golang.org/x/text/language"
)

const (
	formatParam   = "f"
	languageParam = "lang"
)

type ContentNegotiation struct {
	availableMediaTypes []contenttype.MediaType
	availableLanguages  []language.Tag

	formatsByMediaType map[string]string
	mediaTypesByFormat map[string]string
}

func newContentNegotiation(availableLanguages []language.Tag) *ContentNegotiation {
	availableMediaTypes := []contenttype.MediaType{
		// in order
		contenttype.NewMediaType("application/json"),
		contenttype.NewMediaType("text/html"),
		contenttype.NewMediaType("application/vnd.mapbox.tile+json"),
		contenttype.NewMediaType("application/vnd.mapbox-vector-tile"),
		contenttype.NewMediaType("application/vnd.mapbox.style+json"),
		contenttype.NewMediaType("application/vnd.custom.style+json"),
		contenttype.NewMediaType("application/vnd.ogc.sld+xml;version=1.0"),
	}

	formatsByMediaType := map[string]string{
		"application/json":                        FormatJSON,
		"text/html":                               FormatHTML,
		"application/vnd.mapbox.tile+json":        "tilejson",
		"application/vnd.mapbox-vector-tile":      "pbf", // could also be 'mvt', but 'pbf' is more widely used.
		"application/vnd.mapbox.style+json":       "mapbox",
		"application/vnd.custom.style+json":       "custom",
		"application/vnd.ogc.sld+xml;version=1.0": "sld10",
	}

	mediaTypesByFormat := reverseMap(formatsByMediaType)

	return &ContentNegotiation{
		availableMediaTypes: availableMediaTypes,
		availableLanguages:  availableLanguages,
		formatsByMediaType:  formatsByMediaType,
		mediaTypesByFormat:  mediaTypesByFormat,
	}
}

func (cn *ContentNegotiation) GetSupportedStyleFormats() []string {
	return []string{"mapbox", "custom", "sld10"}
}

func (cn *ContentNegotiation) GetStyleFormatExtension(format string) string {
	extensionsByFormat := map[string]string{
		"mapbox": ".json",
		"custom": ".style",
		"sld10":  ".sld",
	}
	if extension, exists := extensionsByFormat[format]; exists {
		return extension
	}
	return ""
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
		requestedLanguage = cn.getLanguageFromAcceptLanguageHeader(req)
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
	if queryParams.Get(formatParam) != "" {
		requestedFormat = queryParams.Get(formatParam)

		// remove ?f= parameter, to prepare for rewrite
		queryParams.Del(formatParam)
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

		// check for presence of language cookie, create cookie if not present, update if present and language doesn't match
		cookie, err := req.Cookie("lang")
		if err != nil {
			cookie = &http.Cookie{
				Name:     "lang",
				Value:    lang,
				Path:     "/",
				MaxAge:   60 * 60 * 24,
				SameSite: http.SameSiteStrictMode,
				Secure:   true,
			}
		} else if cookie.Value != lang {
			cookie.Value = lang
		}
		http.SetCookie(w, cookie)

		// remove ?lang= parameter, to prepare for rewrite
		queryParams.Del(languageParam)
		req.URL.RawQuery = queryParams.Encode()
	}
	return requestedLanguage
}

func (cn *ContentNegotiation) getLanguageFromCookie(req *http.Request) language.Tag {
	var requestedLanguage = language.Und
	cookie, err := req.Cookie("lang")
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

func (cn *ContentNegotiation) getLanguageFromAcceptLanguageHeader(req *http.Request) language.Tag {
	var requestedLanguage = language.Und
	if req.Header.Get("Accept-Language") != "" {
		accepted, _, err := language.ParseAcceptLanguage(req.Header.Get("Accept-Language"))
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

func reverseMap(input map[string]string) map[string]string {
	output := make(map[string]string)
	for k, v := range input {
		output[v] = k
	}
	return output
}
