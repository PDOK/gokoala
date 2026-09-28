package engine

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestNegotiateFormat(t *testing.T) {
	cn := newContentNegotiation([]config.Language{{Tag: language.Dutch}, {Tag: language.English}})
	chromeAcceptHeader := "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"

	tests := []struct {
		accept string
		url    string
		format string
	}{
		{"application/json", "http://pdok.example/ogc/api", "json"},
		{"application/json", "http://pdok.example/ogc/api/", "json"},
		{chromeAcceptHeader, "http://pdok.example/ogc/api", "html"},
		{chromeAcceptHeader, "http://pdok.example/ogc/api/", "html"},
		{"application/json", "http://pdok.example/ogc/api.json", "json"},
		{"application/json", "http://pdok.example/ogc/api?f=json", "json"},
		{"this is an invalid accept header &%%$$&#$&##&", "http://pdok.example/ogc/api", "json"}, // fallback
		{"", "http://pdok.example/ogc/api?f=json", "json"},
		{"application/xml, application/json, text/css, text/html", "http://pdok.example/ogc/api/", "xml"},
		{"application/json, application/xml, text/css, text/html", "http://pdok.example/ogc/api/", "json"},
		{"text/markdown", "http://pdok.example/ogc/api", "md"},
		{"application/json", "http://pdok.example/ogc/api?f=md", "md"},
		{"", "http://pdok.example/ogc/api?f=md", "md"},
		{"text/markdown, application/json;q=0.9", "http://pdok.example/ogc/api", "md"},
	}

	for _, tt := range tests {
		t.Run(tt.accept+" @ "+tt.url, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, tt.url, nil)
			req.Header.Set(HeaderAccept, tt.accept)
			require.NoError(t, err)
			format := cn.NegotiateFormat(req)
			assert.Equal(t, tt.format, format)
		})
	}
}

func TestNegotiateLanguage(t *testing.T) {
	cn := newContentNegotiation([]config.Language{{Tag: language.Dutch}, {Tag: language.English}})

	tests := []struct {
		accept   string
		url      string
		language language.Tag
	}{
		{"nl;q=1", "http://pdok.example/ogc/api", language.Dutch},
		{"fr;q=0.8, de;q=0.5", "http://pdok.example/ogc/api", language.Dutch},
		{"en;q=1", "http://pdok.example/ogc/api", language.English},
		{"", "http://pdok.example/ogc/api", language.Dutch},
		{"", "http://pdok.example/ogc/api?lang=fr", language.Dutch},
		{"", "http://pdok.example/ogc/api?lang=en", language.English},
		{"this is an invalid accept language header &%%$$&#$&##&", "http://pdok.example/ogc/api", language.Dutch}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.accept+" @ "+tt.url, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, tt.url, nil)
			req.Header.Set(HeaderAcceptLanguage, tt.accept)
			require.NoError(t, err)
			lang := cn.NegotiateLanguage(httptest.NewRecorder(), req)
			assert.Equal(t, tt.language, lang)
		})
	}
}

func TestNegotiateLanguageWithCookie(t *testing.T) {
	cn := newContentNegotiation([]config.Language{{Tag: language.Dutch}, {Tag: language.English}})

	tests := []struct {
		cookieLang string
		url        string
		language   language.Tag
	}{
		{"en", "http://pdok.example/ogc/api", language.English},
		{"nl", "http://pdok.example/ogc/api", language.Dutch},
		{"", "http://pdok.example/ogc/api", language.Dutch},
		{"this is an invalid cookie value &%%$$&#$&##&", "http://pdok.example/ogc/api", language.Dutch}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.cookieLang+" @ "+tt.url, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, tt.url, nil)
			req.AddCookie(&http.Cookie{
				Name:     languageParam,
				Value:    tt.cookieLang,
				Path:     "/",
				MaxAge:   config.CookieMaxAge,
				SameSite: http.SameSiteStrictMode,
				Secure:   true,
			})
			require.NoError(t, err)
			lang := cn.NegotiateLanguage(httptest.NewRecorder(), req)
			assert.Equal(t, tt.language, lang)
		})
	}
}
