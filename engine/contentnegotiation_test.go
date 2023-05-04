package engine

import (
	"context"
	"net/http"
	"testing"
)

func TestContentNegotiation_NegotiateFormat(t *testing.T) {
	// given
	cn := newContentNegotiation()
	chromeAcceptHeader := "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"

	// when/then
	testFormat(t, cn, "application/json", "http://pdok.example/ogc/api", "json")
	testFormat(t, cn, "application/json", "http://pdok.example/ogc/api/", "json")
	testFormat(t, cn, chromeAcceptHeader, "http://pdok.example/ogc/api", "html")
	testFormat(t, cn, chromeAcceptHeader, "http://pdok.example/ogc/api/", "html")
	testFormat(t, cn, "application/json", "http://pdok.example/ogc/api.json", "json")
	testFormat(t, cn, "application/json", "http://pdok.example/ogc/api?f=json", "json")
	testFormat(t, cn, "", "http://pdok.example/ogc/api?f=json", "json")
	testFormat(t, cn, "application/xml, application/json, text/css, text/html", "http://pdok.example/ogc/api/", "json")
}

func testFormat(t *testing.T, cn *ContentNegotiation, acceptHeader string, givenURL string, expectedFormat string) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, givenURL, nil)
	req.Header.Set("Accept", acceptHeader)
	if err != nil {
		t.Fatal(err)
	}
	format := cn.NegotiateFormat(req)
	if format != expectedFormat {
		t.Fatalf("Expected %s for input %s, got %s", expectedFormat, givenURL, format)
	}
}
