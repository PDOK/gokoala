package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PDOK/gomagpie/config"
	"golang.org/x/text/language"
)

func TestContentNegotiation_NegotiateFormat(t *testing.T) {
	// given
	cn := newContentNegotiation([]config.Language{{Tag: language.Dutch}, {Tag: language.English}})
	chromeAcceptHeader := "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"

	// when/then
	testFormat(t, cn, "application/json", "http://pdok.example/ogc/api", "json")
	testFormat(t, cn, "application/json", "http://pdok.example/ogc/api/", "json")
	testFormat(t, cn, chromeAcceptHeader, "http://pdok.example/ogc/api", "html")
	testFormat(t, cn, chromeAcceptHeader, "http://pdok.example/ogc/api/", "html")
	testFormat(t, cn, "application/json", "http://pdok.example/ogc/api.json", "json")
	testFormat(t, cn, "application/json", "http://pdok.example/ogc/api?f=json", "json")
	testFormat(t, cn, "", "http://pdok.example/ogc/api?f=json", "json")
	testFormat(t, cn, "application/xml, application/json, text/css, text/html", "http://pdok.example/ogc/api/", "xml")
	testFormat(t, cn, "application/json, application/xml, text/css, text/html", "http://pdok.example/ogc/api/", "json")
	testLanguage(t, cn, "nl;q=1", "http://pdok.example/ogc/api", language.Dutch)
	testLanguage(t, cn, "fr;q=0.8, de;q=0.5", "http://pdok.example/ogc/api", language.Dutch)
	testLanguage(t, cn, "en;q=1", "http://pdok.example/ogc/api", language.English)
	testLanguage(t, cn, "", "http://pdok.example/ogc/api", language.Dutch)
	testLanguage(t, cn, "", "http://pdok.example/ogc/api?lang=fr", language.Dutch)
	testLanguage(t, cn, "", "http://pdok.example/ogc/api?lang=en", language.English)
	testLanguageWithCookie(t, cn, "en", "http://pdok.example/ogc/api", language.English)
}

func testFormat(t *testing.T, cn *ContentNegotiation, acceptHeader string, givenURL string, expectedFormat string) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, givenURL, nil)
	req.Header.Set(HeaderAccept, acceptHeader)
	if err != nil {
		t.Fatal(err)
	}
	format := cn.NegotiateFormat(req)
	if format != expectedFormat {
		t.Fatalf("Expected %s for input %s, got %s", expectedFormat, givenURL, format)
	}
}

func testLanguage(t *testing.T, cn *ContentNegotiation, acceptLanguageHeader string, givenURL string, expectedLanguage language.Tag) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, givenURL, nil)
	req.Header.Set(HeaderAcceptLanguage, acceptLanguageHeader)
	if err != nil {
		t.Fatal(err)
	}
	lang := cn.NegotiateLanguage(httptest.NewRecorder(), req)
	if lang != expectedLanguage {
		t.Fatalf("Expected %v for input %s, got %v", expectedLanguage, givenURL, lang)
	}
}

func testLanguageWithCookie(t *testing.T, cn *ContentNegotiation, cookieLanguage string, givenURL string, expectedLanguage language.Tag) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, givenURL, nil)
	req.AddCookie(&http.Cookie{
		Name:     languageParam,
		Value:    cookieLanguage,
		Path:     "/",
		MaxAge:   config.CookieMaxAge,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	lang := cn.NegotiateLanguage(httptest.NewRecorder(), req)
	if lang != expectedLanguage {
		t.Fatalf("Expected %v for input %s, got %v", expectedLanguage, givenURL, lang)
	}
}
