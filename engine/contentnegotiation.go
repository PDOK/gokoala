package engine

import (
	"log"
	"net/http"

	"github.com/elnormous/contenttype"
)

const formatParam = "f"

type ContentNegotiation struct {
	availableMediaTypes []contenttype.MediaType

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

func reverseMap(input map[string]string) map[string]string {
	output := make(map[string]string)
	for k, v := range input {
		output[v] = k
	}
	return output
}
