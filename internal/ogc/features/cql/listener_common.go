package cql

import (
	"log"

	"github.com/PDOK/gokoala/internal/engine/types"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/features/cql/parser"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

// CommonListener shared logic between CQL listeners.
type CommonListener struct {
	*parser.BaseCqlParserListener

	// stack holds the current SQL clause being built.
	stack *types.Stack

	// namedParams holds named parameters used in the SQL clause (to protect against SQL injection).
	namedParams map[string]any

	// queryables the list of allowed columns in the datasource that can be queried.
	queryables []string

	// randomizer is used to generate unique named parameters.
	randomizer util.Randomizer

	// errorListener is used to collect parse errors.
	errorListener *ErrorListener
}

// AddErrorListener adds an ErrorListener to this listener.
func (cl *CommonListener) AddErrorListener(errorListener *ErrorListener) {
	cl.errorListener = errorListener
}

// generateNamedParam generates a unique named parameter (e.g. :abc or @abc)
// for parameter binding in SQL prepared statements.
func (cl *CommonListener) generateNamedParam(symbol string) (withoutSymbol, withSymbol string) {
RETRY:
	chars := make([]byte, 4)
	for i := range chars {
		chars[i] = alphabet[cl.randomizer.IntN(len(alphabet))]
	}

	withoutSymbol = "cql_" + string(chars) // for example "cql_xmzq" or "cql_abri"
	withSymbol = symbol + withoutSymbol    // for example "@cql_xmzq" or ":cql_abri"
	_, exists := cl.namedParams[withoutSymbol]
	if exists {
		log.Printf("WARNING: generated duplicate named parameter: '%s', retrying...", withoutSymbol)
		goto RETRY
	}
	return
}

func (cl *CommonListener) allowAllQueryables() bool {
	return len(cl.queryables) == 1 && cl.queryables[0] == "*"
}
