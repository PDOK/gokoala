package maps

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (m *Maps) Legend() http.HandlerFunc {
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

		http.ServeFile(w, r, style.LegendPath)
	}
}
