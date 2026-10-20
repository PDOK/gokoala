package maps

import (
	"net/http"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/twpayne/go-geom"
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

func (m imageMaps) renderMapsImage(w http.ResponseWriter, r *http.Request, format string, crs domain.ContentCrs, srid domain.SRID, bbox *geom.Bounds, width int, height int, err error) {
	panic("Not implemented")
}
