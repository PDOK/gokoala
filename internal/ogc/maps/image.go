package maps

import (
	"log"
	"net/http"
	"os"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/twpayne/go-geom"
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

func (m imageMaps) renderMapsImage(w http.ResponseWriter, r *http.Request, format string, style *config.MapStyle,
	crs domain.ContentCrs, srid domain.SRID, bbox *geom.Bounds, width int, height int, err error) {
	var path string
	if format == engine.FormatJPEG {
		// temporary placeholder, do the WMS call here
		path = placeholderImageJPEG
	} else {
		// temporary placeholder, do the WMS call here
		path = placeholderImagePNG
	}
	img, err := os.ReadFile(path)
	if err != nil {
		log.Printf("error reading file: %v", err)
		engine.RenderProblem(engine.ProblemServerError, w)
		return
	}
	w.Header().Set("Content-Type", "image/"+format)
	if _, err = w.Write(img); err != nil {
		log.Printf("error writing response: %v", err)
		engine.RenderProblem(engine.ProblemServerError, w)
	}
}
