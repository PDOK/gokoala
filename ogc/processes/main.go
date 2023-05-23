package processes

import (
	"net/http"

	"github.com/PDOK/gokoala/engine"

	"github.com/go-chi/chi/v5"
)

type Processes struct {
	engine *engine.Engine
}

func NewProcesses(e *engine.Engine, router *chi.Mux) *Processes {
	processes := &Processes{engine: e}
	router.Handle("/jobs*", processes.forwarder(e.Config.OgcAPI.Processes.ProcessesServer))
	router.Handle("/processes*", processes.forwarder(e.Config.OgcAPI.Processes.ProcessesServer))
	router.Handle("/api*", processes.forwarder(e.Config.OgcAPI.Processes.ProcessesServer))
	return processes
}

func (p *Processes) forwarder(processServer engine.YAMLURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURL := *processServer.URL
		targetURL.Path = processServer.URL.Path + r.URL.Path
		targetURL.RawQuery = r.URL.RawQuery
		p.engine.ReverseProxy(w, r, &targetURL, false, "")
	}
}
