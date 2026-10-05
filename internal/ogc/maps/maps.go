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

		// TODO: Find out how to construct the URL for the maps collection
		// What params? How do we go about type-safing them? Validation?
		url := mapsCollectionURL{}

		// TODO: Get properties from URL and validate them
		// TODO: Prepare/Send WMS request (mock for now)
		// TODO: Get format (JSON/HTML/IMG)

		// TODO: Finish handling the request and set the proper headers
	}
}

func handleCollectionNotFound(w http.ResponseWriter, collectionID string) {
	msg := fmt.Sprintf("collection %s doesn't exist in this features service", collectionID)
	log.Println(msg)
	engine.RenderProblem(engine.ProblemNotFound, w, msg)
}
