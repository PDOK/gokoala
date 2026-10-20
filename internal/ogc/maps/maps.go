package maps

import (
	"fmt"
	"log"
	"net/http"

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

		url := mapsURL{
			baseURL: *m.engine.Config.BaseURL.URL,
			params:  r.URL.Query(),
		}

		crs, bboxSRID, bbox, width, height, err := url.parse()
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}

		w.Header().Add(engine.HeaderContentCrs, crs.ToLink())

		format := m.engine.CN.NegotiateFormat(r)
		switch format {
		case engine.FormatHTML:
			m.html.maps(w, r, format, collection, crs, bboxSRID, bbox, width, height, err)
			break
		case engine.FormatJPG, engine.FormatPNG:
			m.image.renderMapsImage(w, r, format, crs, bboxSRID, bbox, width, height, err)
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

func handleFormatNotSupported(w http.ResponseWriter, format string) {
	msg := fmt.Sprintf("format %s is not supported", format)
	log.Println(msg)
	engine.RenderProblem(engine.ProblemNotAcceptable, w, msg)
}
