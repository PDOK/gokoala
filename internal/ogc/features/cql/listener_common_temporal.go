package cql

import (
	"fmt"
	"strings"
)

const (
	// temporalIntervalSeparator custom separator to encode a temporal interval, so we can split it later.
	temporalIntervalSeparator = "##"
	// temporalIntervalUnbounded to indicate an unbounded interval in CQL.
	temporalIntervalUnbounded = ".."
)

func isInstant(firstStart string, firstEnd string, secondStart string, secondEnd string) bool {
	return firstStart == firstEnd && secondStart == secondEnd
}

func condition(first, comparisonOperator, second string) string {
	return fmt.Sprintf("%s %s %s", first, comparisonOperator, second)
}

//nolint:unparam // can be removed once we've implemented the Postgres cql listener
func (cl *CommonListener) addTemporalLiteral(literal string, symbol string) {
	withoutSymbol, withSymbol := cl.generateNamedParam(symbol)
	cl.namedParams[withoutSymbol] = strings.Trim(literal, "'")
	cl.stack.Push(withSymbol)
}

func (cl *CommonListener) pushTemporal(booleanOperator string, sql ...string) {
	if len(sql) == 1 {
		cl.stack.Push(sql[0])
	} else {
		cl.stack.Push("(" + strings.Join(sql, " "+booleanOperator+" ") + ")")
	}
}

// temporalAfter first starts after the end of second
// see https://www.w3.org/TR/owl-time/#time:after
func (cl *CommonListener) temporalAfter(firstStart string, secondEnd string) {
	if secondEnd != temporalIntervalUnbounded {
		cl.pushTemporal("AND", condition(firstStart, ">", secondEnd))
	}
}

// temporalBefore inverse of temporalAfter (T_AFTER)
// see https://www.w3.org/TR/owl-time/#time:before
func (cl *CommonListener) temporalBefore(firstEnd string, secondStart string) {
	if secondStart != temporalIntervalUnbounded {
		cl.pushTemporal("AND", condition(firstEnd, "<", secondStart))
	}
}

// temporalEquals first and second are equal
// see https://www.w3.org/TR/owl-time/#time:intervalEquals
func (cl *CommonListener) temporalEquals(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	if isInstant(firstStart, firstEnd, secondStart, secondEnd) {
		// in case it's an instant (not an interval), just push a simple equal condition
		cl.stack.Push(condition(firstStart, "=", secondStart))
	} else {
		if secondStart == temporalIntervalUnbounded || secondEnd == temporalIntervalUnbounded {
			return
		}
		cl.pushTemporal("AND",
			condition(firstStart, "=", secondStart),
			condition(firstEnd, "=", secondEnd),
		)
	}
}
