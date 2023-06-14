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

func newContentNegotiation() *ContentNegotiation {
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

	availableLanguages := []language.Tag{
		// in order
		language.Dutch,
		language.English,
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
func (cn *ContentNegotiation) NegotiateLanguage(req *http.Request) language.Tag {
	requestedLanguage, err := cn.getLanguageFromQueryParam(req)
	if err != nil || requestedLanguage == language.Und {
		requestedLanguage, err = cn.getLanguageFromAcceptLanguageHeader(req)
	}
	if err != nil || requestedLanguage == language.Und {
		requestedLanguage = language.Dutch // default
	}
	log.Printf("dutch language: %v", language.Dutch)
	log.Printf("negotiated language: %v", requestedLanguage)
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

func (cn *ContentNegotiation) getLanguageFromQueryParam(req *http.Request) (language.Tag, error) {
	var requestedLanguage = language.Und
	queryParams := req.URL.Query()
	if queryParams.Get(languageParam) != "" {
		lang := queryParams.Get(languageParam)
		accepted, _, err := language.ParseAcceptLanguage(lang)
		if err != nil {
			return language.Und, err
		}
		m := language.NewMatcher(cn.availableLanguages)
		_, langIndex, _ := m.Match(accepted...)
		requestedLanguage = cn.availableLanguages[langIndex]

		// remove ?lang= parameter, to prepare for rewrite
		queryParams.Del(languageParam)
		req.URL.RawQuery = queryParams.Encode()
	}
	return requestedLanguage, nil
}

func (cn *ContentNegotiation) getLanguageFromAcceptLanguageHeader(req *http.Request) (language.Tag, error) {
	var requestedLanguage = language.Und
	if req.Header.Get("Accept-Language") != "" {
		accepted, _, err := language.ParseAcceptLanguage(req.Header.Get("Accept-Language"))
		if err != nil {
			log.Printf("Failed to parse Accept-Language header: %v. Continuing\n", err)
			return language.Und, err
		}
		m := language.NewMatcher(cn.availableLanguages)
		_, langIndex, _ := m.Match(accepted...)
		requestedLanguage = cn.availableLanguages[langIndex]
	}
	return requestedLanguage, nil
}

func reverseMap(input map[string]string) map[string]string {
	output := make(map[string]string)
	for k, v := range input {
		output[v] = k
	}
	return output
}
