package maps

import (
	"fmt"
	"log"
	"mime"
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
	getMapRequest, err := buildGetMapRequest(&request, mapsConfig)
	if err != nil {
		log.Printf("failed to build WMS request: %v", err)
		engine.RenderProblem(engine.ProblemServerError, w)
		return
	}

	target := *mapsConfig.WMSServer.URL
	params := target.Query()
	getMapParams := getMapRequest.ToQueryParameters()
	for key, values := range getMapParams {
		params[key] = values
	}
	target.RawQuery = params.Encode()

	m.engine.ReverseProxyAndValidate(w, r, &target, false, "", false, func(response *http.Response) error {
		if err := validateGetMapResponse(response, request.format); err != nil {
			return err
		}
		response.Header.Set(engine.HeaderContentCrs, request.crs.ToLink())
		response.Header.Set(engine.HeaderContentBbox, getMapParams.Get("BBOX"))
		return nil
	})
}

func validateGetMapResponse(response *http.Response, expectedFormat string) error {
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("WMS returned HTTP %d", response.StatusCode)
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf("invalid WMS Content-Type: %w", err)
	}
	if mediaType != expectedFormat {
		return fmt.Errorf("WMS returned media type %q, expected %q", mediaType, expectedFormat)
	}
	return nil
}
