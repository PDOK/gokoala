package cql

import (
	"github.com/PDOK/gokoala/internal/engine/types"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/features/cql/parser"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

// CommonListener shared logic between CQL listeners.
type CommonListener struct {
	*parser.BaseCqlParserListener

	stack       *types.Stack
	namedParams map[string]any
	randomizer  util.Randomizer
}

// generateUniqueNamedParam generates a unique named parameter (e.g. :abc or @abc)
// for parameter binding in prepared statements.
func (cl *CommonListener) generateUniqueNamedParam(symbol string) (withoutSymbol, withSymbol string) {
RETRY:
	chars := make([]byte, 4)
	for i := range chars {
		chars[i] = alphabet[cl.randomizer.IntN(len(alphabet))]
	}

	withoutSymbol = "cql_" + string(chars) // for example "cql_xmzq" or "cql_abri"
	withSymbol = symbol + withoutSymbol    // for example "@cql_xmzq" or ":cql_abri"
	_, exists := cl.namedParams[withoutSymbol]
	if exists {
		goto RETRY
	}
	return
}
