package core

import (
	"net/http"

	"github.com/PDOK/gokoala/engine"
	"github.com/go-chi/chi/v5"
)

const (
	templatesDir       = "ogc/common/core/templates/"
	rootPath           = "/"
	apiPath            = "/api"
	alternativeAPIPath = "/openapi.json"
	conformancePath    = "/conformance"
)

type CommonCore struct {
	engine *engine.Engine
}

func NewCommonCore(e *engine.Engine, router *chi.Mux) *CommonCore {
	conformanceBreadcrumbs := []engine.Breadcrumb{
		{
			Name: "Conformance",
			Path: "conformance",
		},
	}
	apiBreadcrumbs := []engine.Breadcrumb{
		{
			Name: "OpenAPI specificatie",
			Path: "api",
		},
	}

	e.RenderTemplates(rootPath,
		nil,
		engine.NewTemplateKey(templatesDir+"landing-page.go.json"),
		engine.NewTemplateKey(templatesDir+"landing-page.go.html"))
	e.RenderTemplates(rootPath,
		apiBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"api.go.html"))
	e.RenderTemplates(conformancePath,
		conformanceBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"conformance.go.json"),
		engine.NewTemplateKey(templatesDir+"conformance.go.html"))
	core := &CommonCore{
		engine: e,
	}

	router.Get(rootPath, core.LandingPage())
	router.Get(apiPath, core.API())
	// implements https://gitdocumentatie.logius.nl/publicatie/api/adr/#api-17
	router.Get(alternativeAPIPath, func(w http.ResponseWriter, r *http.Request) { core.apiAsJSON(w) })
	router.Get(conformancePath, core.Conformance())
	router.Handle("/*", http.FileServer(http.Dir("assets")))

	return core
}

func (c *CommonCore) LandingPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKeyWithLanguage(templatesDir+"landing-page.go."+c.engine.CN.NegotiateFormat(r), c.engine.CN.NegotiateLanguage(w, r))
		c.engine.ServePage(w, r, key)
	}
}

func (c *CommonCore) API() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := c.engine.CN.NegotiateFormat(r)
		if format == engine.FormatHTML {
			c.apiAsHTML(w, r)
			return
		} else if format == engine.FormatJSON {
			c.apiAsJSON(w)
			return
		}
		http.NotFound(w, r)
	}
}

func (c *CommonCore) apiAsHTML(w http.ResponseWriter, r *http.Request) {
	key := engine.NewTemplateKeyWithLanguage(templatesDir+"api.go.html", c.engine.CN.NegotiateLanguage(w, r))
	c.engine.ServePage(w, r, key)
}

func (c *CommonCore) apiAsJSON(w http.ResponseWriter) {
	w.Header().Set(engine.HeaderContentType, engine.MediaTypeOpenAPI)
	engine.SafeWrite(w.Write, c.engine.OpenAPI.SpecJSON)
}

func (c *CommonCore) Conformance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKeyWithLanguage(templatesDir+"conformance.go."+c.engine.CN.NegotiateFormat(r), c.engine.CN.NegotiateLanguage(w, r))
		c.engine.ServePage(w, r, key)
	}
}
