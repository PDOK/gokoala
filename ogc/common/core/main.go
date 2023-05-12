package core

import (
	"net/http"

	"github.com/PDOK/gokoala/engine"

	"github.com/go-chi/chi/v5"
)

const (
	templatesDir    = "ogc/common/core/templates/"
	rootPath        = "/"
	apiPath         = "/api"
	conformancePath = "/conformance"
)

type CommonCore struct {
	engine *engine.Engine
}

func NewCommonCore(e *engine.Engine, router *chi.Mux) *CommonCore {
	conformanceBreadcrumbs := []engine.Breadcrumb{
		engine.Breadcrumb{
			Name: "Conformance",
			Path: "conformance",
		},
	}

	e.RenderTemplates(rootPath,
		nil,
		engine.NewTemplateKey(templatesDir+"landing-page.go.json"),
		engine.NewTemplateKey(templatesDir+"landing-page.go.html"))
	e.RenderTemplates(conformancePath,
		conformanceBreadcrumbs,
		engine.NewTemplateKey(templatesDir+"conformance.go.json"),
		engine.NewTemplateKey(templatesDir+"conformance.go.html"))

	core := &CommonCore{
		engine: e,
	}

	router.Get(rootPath, core.LandingPage())
	router.Get(apiPath, core.API())
	router.Get(conformancePath, core.Conformance())
	router.Handle("/*", http.FileServer(http.Dir("assets")))

	return core
}

func (c *CommonCore) LandingPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKey(templatesDir + "landing-page.go." + c.engine.CN.NegotiateFormat(r))
		c.engine.ServePage(w, r, key)
	}
}

func (c *CommonCore) API() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := c.engine.CN.NegotiateFormat(r)
		if format == engine.FormatHTML {
			// skip templating for swagger UI
			http.ServeFile(w, r, templatesDir+"api.html")
			return
		} else if format == engine.FormatJSON {
			w.Header().Set("Content-Type", "application/vnd.oai.openapi+json;version=3.0")
			engine.SafeWrite(w.Write, c.engine.OpenAPI.SpecJSON)
			return
		}
		http.NotFound(w, r)
	}
}

func (c *CommonCore) Conformance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := engine.NewTemplateKey(templatesDir + "conformance.go." + c.engine.CN.NegotiateFormat(r))
		c.engine.ServePage(w, r, key)
	}
}
