package maps

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/go-chi/chi/v5"
)

func (m *Maps) Maps() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := chi.URLParam(r, "collectionId")
		collection, ok := m.configuredCollections[collectionID]
		if !ok {
			handleCollectionNotFound(w, collectionID)
			return
		}

		styleID := chi.URLParam(r, "styleId")
		var style *config.MapStyle
		if styleID != "" {
			style = collection.GetStyle(styleID)
			if style == nil {
				handleStyleNotFound(w, styleID)
				return
			}
		}

		url := mapsURL{
			baseURL: *m.engine.Config.BaseURL.URL,
			params:  r.URL.Query(),
		}

		request, err := url.parse()
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}

		request.mapType = CollectionMap
		request.collection = collectionID
		request.style = styleID

		if err := request.setDefaults(m.engine.Config.OgcAPI.Maps.DefaultMap); err != nil {
			engine.RenderProblem(engine.ProblemServerError, w)
			return
		}

		if err := request.validate(); err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}

		w.Header().Add(engine.HeaderContentCrs, request.crs.ToLink())

		format := m.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatHTML:
			m.html.maps(w, r, format, collection, request)
			break
		case engine.FormatJPEG, engine.FormatPNG:
			request.format = format
			m.image.renderMapsImage(w, r, request)
			break
		default:
			handleFormatNotSupported(w, format)
		}
	}
}

func handleCollectionNotFound(w http.ResponseWriter, collectionID string) {
	msg := fmt.Sprintf("collection %s doesn't exist in this features service", collectionID)
	log.Println(msg)
	engine.RenderProblem(engine.ProblemNotFound, w, msg)
}

func handleStyleNotFound(w http.ResponseWriter, styleID string) {
	msg := fmt.Sprintf("style %s doesn't exist in this features service", styleID)
	log.Println(msg)
	engine.RenderProblem(engine.ProblemNotFound, w, msg)
}

func handleFormatNotSupported(w http.ResponseWriter, format string) {
	msg := fmt.Sprintf("format %s is not supported", format)
	log.Println(msg)
	engine.RenderProblem(engine.ProblemNotAcceptable, w, msg)
}
