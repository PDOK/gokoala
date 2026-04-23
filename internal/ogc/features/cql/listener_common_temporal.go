package cql

import (
	"fmt"
	"strings"
)

//nolint:revive // keep these inline with spec.
const (
	T_AFTER        = "T_AFTER"
	T_BEFORE       = "T_BEFORE"
	T_EQUALS       = "T_EQUALS"
	T_DISJOINT     = "T_DISJOINT"
	T_INTERSECTS   = "T_INTERSECTS"
	T_DURING       = "T_DURING"
	T_CONTAINS     = "T_CONTAINS"
	T_FINISHES     = "T_FINISHES"
	T_FINISHEDBY   = "T_FINISHEDBY"
	T_MEETS        = "T_MEETS"
	T_METBY        = "T_METBY"
	T_OVERLAPS     = "T_OVERLAPS"
	T_OVERLAPPEDBY = "T_OVERLAPPEDBY"
	T_STARTS       = "T_STARTS"
	T_STARTEDBY    = "T_STARTEDBY"
)

const (
	// temporalIntervalSeparator custom separator to encode a temporal interval, so we can split it later.
	temporalIntervalSeparator = "##"
	// temporalIntervalUnbounded to indicate an unbounded interval in CQL.
	temporalIntervalUnbounded = ".."
)

const (
	errUnboundedSecondParamNotAllowed = "temporal function %s requires a second parameter, can't be unbounded ('..')"
	errInstantNotAllowed              = "temporal function %s only allows intervals, not instants (timestamp/date)"
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

func (cl *CommonListener) popIntervalOrInstant() (firstStart string, firstEnd string, secondStart string, secondEnd string) {
	second := cl.stack.Pop()
	first := cl.stack.Pop()

	splitWhenInterval := func(s string) (start, end string) {
		if idx := strings.Index(s, temporalIntervalSeparator); idx >= 0 {
			// it's an interval, split into start and end
			return s[:idx], s[idx+(len(temporalIntervalSeparator)):]
		}
		return s, s // not an interval but an instant (timestamp, date)
	}

	firstStart, firstEnd = splitWhenInterval(first)
	secondStart, secondEnd = splitWhenInterval(second)
	return
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
	if secondEnd == temporalIntervalUnbounded {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_AFTER)
		return
	}
	cl.pushTemporal("AND", condition(firstStart, ">", secondEnd))
}

// temporalBefore inverse of temporalAfter (T_AFTER)
// see https://www.w3.org/TR/owl-time/#time:before
func (cl *CommonListener) temporalBefore(firstEnd string, secondStart string) {
	if secondStart == temporalIntervalUnbounded {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_BEFORE)
		return
	}
	cl.pushTemporal("AND", condition(firstEnd, "<", secondStart))
}

// temporalEquals first and second are equal
// see https://www.w3.org/TR/owl-time/#time:intervalEquals
func (cl *CommonListener) temporalEquals(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	if isInstant(firstStart, firstEnd, secondStart, secondEnd) {
		// in case it's an instant (not an interval), just push a simple equal condition
		cl.stack.Push(condition(firstStart, "=", secondStart))
	} else {
		if secondStart == temporalIntervalUnbounded || secondEnd == temporalIntervalUnbounded {
			cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_EQUALS)
			return
		}
		cl.pushTemporal("AND",
			condition(firstStart, "=", secondStart),
			condition(firstEnd, "=", secondEnd),
		)
	}
}

// temporalDisjoint first and second do not overlap.
// see https://www.w3.org/TR/owl-time/#time:intervalDisjoint
func (cl *CommonListener) temporalDisjoint(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	var sql []string
	if secondStart != temporalIntervalUnbounded {
		sql = append(sql, condition(firstEnd, "<", secondStart))
	}
	if secondEnd != temporalIntervalUnbounded {
		sql = append(sql, condition(firstStart, ">", secondEnd))
	}
	if len(sql) == 0 {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_DISJOINT)
		return
	}
	cl.pushTemporal("OR", sql...)
}

// temporalIntersects inverse of temporalDisjoint (T_DISJOINT)
func (cl *CommonListener) temporalIntersects(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	var sql []string
	if secondEnd != temporalIntervalUnbounded {
		sql = append(sql, condition(firstStart, "<=", secondEnd))
	}
	if secondStart != temporalIntervalUnbounded {
		sql = append(sql, condition(firstEnd, ">=", secondStart))
	}
	if len(sql) == 0 {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_INTERSECTS)
		return
	}
	cl.pushTemporal("AND", sql...)
}

// temporalDuring first falls within second
// see https://www.w3.org/TR/owl-time/#time:intervalDuring
func (cl *CommonListener) temporalDuring(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	if isInstant(firstStart, firstEnd, secondStart, secondEnd) {
		cl.errorListener.Errorf(errInstantNotAllowed, T_DURING)
		return
	}

	if secondStart == temporalIntervalUnbounded || secondEnd == temporalIntervalUnbounded {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_DURING)
		return
	}
	cl.pushTemporal("AND",
		condition(firstStart, ">", secondStart),
		condition(firstEnd, "<", secondEnd),
	)
}

// temporalContains inverse of temporalDuring (T_DURING)
// see https://www.w3.org/TR/owl-time/#time:intervalContains
func (cl *CommonListener) temporalContains(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	if isInstant(firstStart, firstEnd, secondStart, secondEnd) {
		cl.errorListener.Errorf(errInstantNotAllowed, T_CONTAINS)
		return
	}

	if secondStart == temporalIntervalUnbounded || secondEnd == temporalIntervalUnbounded {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_CONTAINS)
		return
	}
	cl.pushTemporal("AND",
		condition(firstStart, "<", secondStart),
		condition(firstEnd, ">", secondEnd),
	)
}

// temporalStarts first and second share the same start, and first ends before second
// ee https://www.w3.org/TR/owl-time/#time:intervalStarts
func (cl *CommonListener) temporalStarts(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	if isInstant(firstStart, firstEnd, secondStart, secondEnd) {
		cl.errorListener.Errorf(errInstantNotAllowed, T_STARTS)
		return
	}

	if secondStart == temporalIntervalUnbounded {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_STARTS)
		return
	}
	sql := []string{condition(firstStart, "=", secondStart)}
	if secondEnd != temporalIntervalUnbounded {
		sql = append(sql, condition(firstEnd, "<", secondEnd))
	}
	cl.pushTemporal("AND", sql...)
}

// temporalStartedBy inverse of temporalStarts (T_STARTS)
// see https://www.w3.org/TR/owl-time/#time:intervalStartedBy
func (cl *CommonListener) temporalStartedBy(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	if isInstant(firstStart, firstEnd, secondStart, secondEnd) {
		cl.errorListener.Errorf(errInstantNotAllowed, T_STARTEDBY)
		return
	}

	if secondStart == temporalIntervalUnbounded {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_STARTEDBY)
		return
	}
	sql := []string{condition(firstStart, "=", secondStart)}
	if secondEnd != temporalIntervalUnbounded {
		sql = append(sql, condition(firstEnd, ">", secondEnd))
	}
	cl.pushTemporal("AND", sql...)
}

// temporalOverlaps first starts before second, and first ends within second
// see https://www.w3.org/TR/owl-time/#time:intervalOverlaps
func (cl *CommonListener) temporalOverlaps(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	if isInstant(firstStart, firstEnd, secondStart, secondEnd) {
		cl.errorListener.Errorf(errInstantNotAllowed, T_OVERLAPS)
		return
	}

	if secondStart == temporalIntervalUnbounded || secondEnd == temporalIntervalUnbounded {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_OVERLAPS)
		return
	}
	cl.pushTemporal("AND",
		condition(firstStart, "<", secondStart),
		condition(firstEnd, ">", secondStart),
		condition(firstEnd, "<", secondEnd),
	)
}

// temporalOverlappedBy inverse of temporalOverlaps (T_OVERLAPS)
// see https://www.w3.org/TR/owl-time/#time:intervalOverlappedBy
func (cl *CommonListener) temporalOverlappedBy(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	if isInstant(firstStart, firstEnd, secondStart, secondEnd) {
		cl.errorListener.Errorf(errInstantNotAllowed, T_OVERLAPPEDBY)
		return
	}

	if secondStart == temporalIntervalUnbounded || secondEnd == temporalIntervalUnbounded {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_OVERLAPPEDBY)
		return
	}
	cl.pushTemporal("AND",
		condition(firstStart, ">", secondStart),
		condition(firstStart, "<", secondEnd),
		condition(firstEnd, ">", secondEnd),
	)
}

// temporalMeets the end of first equals the start of second.
// see https://www.w3.org/TR/owl-time/#time:intervalMeets
func (cl *CommonListener) temporalMeets(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	if isInstant(firstStart, firstEnd, secondStart, secondEnd) {
		cl.errorListener.Errorf(errInstantNotAllowed, T_MEETS)
		return
	}

	if secondStart == temporalIntervalUnbounded {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_MEETS)
		return
	}
	cl.pushTemporal("AND", condition(firstEnd, "=", secondStart))
}

// temporalMetBy inverse of temporalMeets (T_MEETS)
// see https://www.w3.org/TR/owl-time/#time:intervalMetBy
func (cl *CommonListener) temporalMetBy(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	if isInstant(firstStart, firstEnd, secondStart, secondEnd) {
		cl.errorListener.Errorf(errInstantNotAllowed, T_METBY)
		return
	}

	if secondEnd == temporalIntervalUnbounded {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_METBY)
		return
	}
	cl.pushTemporal("AND", condition(firstStart, "=", secondEnd))
}

// temporalFinishes first starts after second and they share the same end
// see https://www.w3.org/TR/owl-time/#time:intervalFinishes
func (cl *CommonListener) temporalFinishes(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	if isInstant(firstStart, firstEnd, secondStart, secondEnd) {
		cl.errorListener.Errorf(errInstantNotAllowed, T_FINISHES)
		return
	}

	if secondStart == temporalIntervalUnbounded || secondEnd == temporalIntervalUnbounded {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_FINISHES)
		return
	}
	cl.pushTemporal("AND",
		condition(firstStart, ">", secondStart),
		condition(firstEnd, "=", secondEnd),
	)
}

// temporalFinishedBy inverse of temporalFinishes (T_FINISHES)
// see https://www.w3.org/TR/owl-time/#time:intervalFinishedBy
func (cl *CommonListener) temporalFinishedBy(firstStart string, firstEnd string, secondStart string, secondEnd string) {
	if isInstant(firstStart, firstEnd, secondStart, secondEnd) {
		cl.errorListener.Errorf(errInstantNotAllowed, T_FINISHEDBY)
		return
	}

	if secondStart == temporalIntervalUnbounded || secondEnd == temporalIntervalUnbounded {
		cl.errorListener.Errorf(errUnboundedSecondParamNotAllowed, T_FINISHEDBY)
		return
	}
	cl.pushTemporal("AND",
		condition(firstStart, "<", secondStart),
		condition(firstEnd, "=", secondEnd),
	)
}
