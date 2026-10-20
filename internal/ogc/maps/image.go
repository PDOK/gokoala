package maps

import (
	"log"
	"net/http"

	"github.com/PDOK/gokoala/internal/engine"
)

type imageMaps struct {
	engine *engine.Engine
}

// newImageMaps returns an image with the requested map
func newImageMaps(e *engine.Engine) *imageMaps {
	return &imageMaps{
		engine: e,
	}
}

func (m imageMaps) renderMapsImage(w http.ResponseWriter, r *http.Request, request MapRequest) {
	switch request.format {
	case engine.FormatPNG, engine.FormatJPEG:
		request.format = "image/" + request.format
	}

	mapsConfig := m.engine.Config.OgcAPI.Maps
	wmsRequest, err := buildGetMapRequest(&request, mapsConfig)
	if err != nil {
		log.Printf("failed to build WMS request: %v", err)
		engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
		return
	}

	target := *mapsConfig.WMSServer.URL
	params := target.Query()
	for key, values := range wmsRequest.ToQueryParameters() {
		params[key] = values
	}
	target.RawQuery = params.Encode()

	m.engine.ReverseProxy(w, r, &target, false, "")
}
