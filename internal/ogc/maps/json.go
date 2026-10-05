package maps

import (
	"github.com/PDOK/gokoala/internal/engine"
)

type jsonMaps struct {
	engine *engine.Engine
}

func newJSONMaps(e *engine.Engine) *jsonMaps {
	return &jsonMaps{
		engine: e,
	}
}
