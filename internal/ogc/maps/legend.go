package maps

import (
	"fmt"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (m *Maps) Legends() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")
		collection, ok := m.configuredCollections[collectionID]
		if !ok {
			handleCollectionNotFound(w, collectionID)
			return
		}

		styleID := chi.URLParam(r, "styleId")
		style := collection.GetStyle(styleID)
		if style == nil {
			handleStyleNotFound(w, styleID)
			return
		}

		target := *m.engine.Config.OgcAPI.Maps.WMSServer.URL
		target.Path = path.Join(target.Path, "legend", style.LegendURL)

		m.engine.ReverseProxyAndValidate(w, r, &target, false, "", false, validateGetLegendResponse)
	}
}

func validateGetLegendResponse(response *http.Response) error {
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("WMS returned HTTP %d", response.StatusCode)
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf("invalid WMS Content-Type: %w", err)
	}
	if !strings.HasPrefix(mediaType, "image/") {
		return fmt.Errorf("WMS returned media type %q, expected image/*", mediaType)
	}
	return nil
}
