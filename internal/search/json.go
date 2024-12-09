package search

import (
	stdjson "encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/PDOK/gomagpie/internal/engine"
	"github.com/PDOK/gomagpie/internal/search/domain"
	perfjson "github.com/goccy/go-json"
)

var (
	now                            = time.Now // allow mocking
	disableJSONPerfOptimization, _ = strconv.ParseBool(os.Getenv("DISABLE_JSON_PERF_OPTIMIZATION"))
)

func featuresAsGeoJSON(w http.ResponseWriter, fc *domain.FeatureCollection) {
	fc.Timestamp = now().Format(time.RFC3339)
	// fc.Links = createFeatureCollectionLinks(engine.FormatGeoJSON, collectionID, cursor, featuresURL) // TODO add links

	// TODO add validation
	// if jf.validateResponse {
	//	jf.serveAndValidateJSON(&fc, engine.MediaTypeGeoJSON, r, w)
	// } else {
	serveJSON(&fc, engine.MediaTypeGeoJSON, w)
	// }
}

// serveJSON serves JSON *WITHOUT* OpenAPI validation by writing directly to the response output stream
func serveJSON(input any, contentType string, w http.ResponseWriter) {
	w.Header().Set(engine.HeaderContentType, contentType)

	if err := getEncoder(w).Encode(input); err != nil {
		handleJSONEncodingFailure(err, w)
		return
	}
}

type jsonEncoder interface {
	Encode(input any) error
}

// Create JSONEncoder. Note escaping of '<', '>' and '&' is disabled (HTMLEscape is false).
// Especially the '&' is important since we use this character in the next/prev links.
func getEncoder(w io.Writer) jsonEncoder {
	if disableJSONPerfOptimization {
		// use Go stdlib JSON encoder
		encoder := stdjson.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		return encoder
	}
	// use ~7% overall faster 3rd party JSON encoder (in case of issues switch back to stdlib using env variable)
	encoder := perfjson.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder
}

func handleJSONEncodingFailure(err error, w http.ResponseWriter) {
	log.Printf("JSON encoding failed: %v", err)
	engine.RenderProblem(engine.ProblemServerError, w, "Failed to write JSON response")
}
