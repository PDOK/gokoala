package maps

import (
	"log"
	"net/http"
	"os"

	"github.com/PDOK/gokoala/internal/engine"
)

const (
	placeholderImagePNG  = "examples/resources/placeholder.png"
	placeholderImageJPEG = "examples/resources/placeholder.jpeg"
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

	usePlaceholderImage := true

	if usePlaceholderImage {
		var path string
		if request.format == "image/jpeg" {
			// temporary placeholder
			path = placeholderImageJPEG
		} else {
			// temporary placeholder
			path = placeholderImagePNG
		}
		img, err := os.ReadFile(path)
		if err != nil {
			log.Printf("error reading file: %v", err)
			engine.RenderProblem(engine.ProblemServerError, w)
			return
		}
		w.Header().Set("Content-Type", request.format)
		if _, err = w.Write(img); err != nil {
			log.Printf("error writing response: %v", err)
			engine.RenderProblem(engine.ProblemServerError, w)
			return
		}
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
