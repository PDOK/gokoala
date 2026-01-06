package util

import (
	stdjson "encoding/json"
	"io"
	"os"
	"strconv"

	perfjson "github.com/goccy/go-json"
)

var disableJSONPerfOptimization, _ = strconv.ParseBool(os.Getenv("DISABLE_JSON_PERF_OPTIMIZATION"))

type JSONEncoder interface {
	Encode(input any) error
}

// GetJSONEncoder Create JSON encoder. Note escaping of '<', '>' and '&' is disabled (HTMLEscape is false).
// Especially the '&' is important since we use this character in, for example, next/prev links.
func GetJSONEncoder(w io.Writer) JSONEncoder {
	if disableJSONPerfOptimization {
		// use Go stdlib JSON encoder
		encoder := stdjson.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		return encoder
	}
	// use ~7% overall faster 3rd party JSON encoder (in case of issues, switch back to stdlib using env variable)
	encoder := perfjson.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder
}
