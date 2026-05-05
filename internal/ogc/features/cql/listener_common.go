package cql

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/internal/engine/types"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/features/cql/parser"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
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
	queryables []domain.Field

	// srid the filter spatial reference identifier (SRID).
	srid domain.SRID

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
	log.Println("WARNING: using '*' as queryable, this is not recommended")
	return len(cl.queryables) == 1 && cl.queryables[0].Name == "*"
}

// isQueryable checks if a column name is allowed in the query.
func (l *GeoPackageListener) isQueryable(name string) bool {
	for _, q := range l.queryables {
		if q.Name == name || q.IsPrimaryGeometry {
			return true
		}
	}
	return false
}

// hasWildcard checks if a pattern contains a SQL wildcard: % or _.
func (cl *CommonListener) hasWildcard(pattern string, symbol string) bool {
	var namedParam string

	// we're only interested in the named param part of the pattern
	parts := strings.Fields(pattern)
	if len(parts) > 0 {
		namedParam = parts[0]
	} else {
		namedParam = pattern
	}

	// remove symbol
	if strings.HasPrefix(pattern, symbol) {
		namedParam = namedParam[len(symbol):]
	}

	// look up the actual value of the named parameter
	patternValue, ok := cl.namedParams[namedParam]
	if !ok {
		return false
	}
	patternValueStr := fmt.Sprintf("%v", patternValue)
	return strings.Contains(patternValueStr, "%") ||
		strings.Contains(patternValueStr, "_")
}

// parseNumber parses a number from a string, supports both integers and floats.
func parseNumber(s string) (any, error) {
	s = strings.TrimSpace(s)

	if i, err := strconv.ParseInt(s, 0, 64); err == nil {
		return i, nil
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, nil
	}
	return nil, fmt.Errorf("%s is not a valid numeric type", s)
}

// stripSingleQuotes removes single quotes from a literal.
func stripSingleQuotes(s string) string {
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return strings.ReplaceAll(s[1:len(s)-1], "''", "'")
	}
	return s
}
