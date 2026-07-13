package cql

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/types"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	"github.com/PDOK/gokoala/internal/ogc/features/cql/parser"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/common"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

const collateKeyword = " COLLATE "

const (
	errAdvancedComparisonNotEnabled        = "advanced comparison operators (LIKE, BETWEEN, IN, IS NULL) are not enabled for this collection"
	errCaseInsensitiveOperatorNotEnabled   = "case-insensitive comparison (CASEI) is not enabled for this collection"
	errAccentInsensitiveOperatorNotEnabled = "accent-insensitive comparison (ACCENTI) is not enabled for this collection"
	errSpatialOperatorsNotEnabled          = "spatial operators are not enabled for this collection"
	errTemporalOperatorsNotEnabled         = "temporal operators are not enabled for this collection"
)

// CommonListener shared logic between CQL listeners.
type CommonListener struct {
	*parser.BaseCqlParserListener

	// stack holds the current SQL clause being built.
	stack *types.Stack

	// namedParams holds named parameters used in the SQL clause (to protect against SQL injection).
	namedParams map[string]any

	// cqlConfig settings that enable/disable CQL conformance classes.
	cqlConfig config.CQL

	// queryables the list of allowed columns in the datasource that can be queried.
	queryables []domain.Field

	// srid the filter spatial reference identifier (SRID).
	srid domain.SRID

	// axisOrder the order of axis of the filter spatial reference identifier
	axisOrder domain.AxisOrder

	// randomizer is used to generate unique named parameters.
	randomizer util.Randomizer

	// collectionType specifies if this collection contains features or attributes.
	collectionType geospatial.CollectionType

	// errorListener is used to collect parse errors.
	errorListener *ErrorListener

	// currentWktType is the current WKT type (POINT, POLYGON, etc.) being parsed.
	currentWktType string
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

func (cl *CommonListener) isAllQueryablesAllowed() bool {
	for _, q := range cl.queryables {
		if q.Name == "*" {
			log.Println("WARNING: using '*' as queryable, this is not recommended")
			return true
		}
	}
	return false
}

// isQueryable checks if a column name is allowed in the query.
func (cl *CommonListener) isQueryable(name string) bool {
	if cl.isAllQueryablesAllowed() {
		return true
	}
	for _, q := range cl.queryables {
		if name == q.Name {
			return true
		}
		if name == domain.GeomPropertyName && q.IsPrimaryGeometry {
			return true
		}
	}
	return false
}

// lookupNamedParam looks up a named parameter value
func (cl *CommonListener) lookupNamedParam(namedParam string, symbol string) (any, bool) {
	// remove symbol before looking up the actual value
	namedParam = strings.TrimPrefix(namedParam, symbol)
	val, ok := cl.namedParams[namedParam]
	return val, ok
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

	patternValue, ok := cl.lookupNamedParam(namedParam, symbol)
	if !ok {
		return false
	}
	patternValueAsStr := fmt.Sprintf("%v", patternValue)
	return strings.Contains(patternValueAsStr, "%") ||
		strings.Contains(patternValueAsStr, "_")
}

// isNumericParam checks if a named parameter is a numeric value.
func (cl *CommonListener) isNumericParam(namedParam string, symbol string) bool {
	if namedParam == "" || len(namedParam) < 2 {
		return false
	}
	val, ok := cl.lookupNamedParam(namedParam, symbol)
	if !ok {
		return false
	}

	if types.IsNumeric(val) {
		return true
	}

	// not a number, but perhaps it's a number in a string.
	valAsStr, ok := val.(string)
	if !ok {
		return false
	}
	_, err := parseNumber(valAsStr)
	return err == nil
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

// addCollation adds a COLLATE to the SQL expression to make case and/or accent insensitive comparison possible.
func addCollation(expr, collation string) string {
	suffixCase := collateKeyword + common.IgnoreCaseCollation
	suffixAccent := collateKeyword + common.IgnoreAccentCollation
	suffixAccentCase := collateKeyword + common.IgnoreAccentAndCaseCollation

	switch {
	case hasCollation(expr, common.IgnoreAccentAndCaseCollation):
		return expr
	case hasCollation(expr, common.IgnoreCaseCollation) && collation == common.IgnoreAccentCollation:
		// replace existing case with case + accent
		return strings.Replace(expr, suffixCase, suffixAccentCase, 1)
	case hasCollation(expr, common.IgnoreAccentCollation) && collation == common.IgnoreCaseCollation:
		// replace existing accent with case + accent
		return strings.Replace(expr, suffixAccent, suffixAccentCase, 1)
	default:
		return expr + collateKeyword + collation
	}
}

// removeCollation removes COLLATE from the SQL expression.
func removeCollation(expr string) string {
	collations := []string{common.IgnoreCaseCollation, common.IgnoreAccentCollation, common.IgnoreAccentAndCaseCollation}
	for _, collation := range collations {
		if hasCollation(expr, collation) {
			return strings.TrimSuffix(expr, collateKeyword+collation)
		}
	}
	return expr
}

// hasCollation checks if the SQL expression has a specific COLLATE suffix.
func hasCollation(expr, collation string) bool {
	return strings.HasSuffix(expr, collateKeyword+collation)
}
