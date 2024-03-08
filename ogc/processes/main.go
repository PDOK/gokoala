package processes

import (
	"net/http"

	"github.com/PDOK/gokoala/config"

	"github.com/PDOK/gokoala/engine"
)

type Processes struct {
	engine *engine.Engine
}

func NewProcesses(e *engine.Engine) *Processes {
	processes := &Processes{engine: e}
	e.Router.Handle("/jobs*", processes.forwarder(e.Config.OgcAPI.Processes.ProcessesServer))
	e.Router.Handle("/processes*", processes.forwarder(e.Config.OgcAPI.Processes.ProcessesServer))
	e.Router.Handle("/api*", processes.forwarder(e.Config.OgcAPI.Processes.ProcessesServer))
	return processes
}

func (p *Processes) forwarder(processServer config.ConfigURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURL := *processServer.URL
		targetURL.Path = processServer.URL.Path + r.URL.Path
		targetURL.RawQuery = r.URL.RawQuery
		p.engine.ReverseProxy(w, r, &targetURL, false, "")
	}
}
