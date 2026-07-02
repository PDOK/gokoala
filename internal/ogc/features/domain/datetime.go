package domain

import (
	"time"
)

// DateTime holds a parsed 'datetime' query param, which can be either an instant or an interval
type DateTime struct {
	// Instant a specific moment in time, a reference date
	Instant *time.Time

	// IntervalStart and IntervalEnd are set for interval queries. When 'nil' the interval is open/unbounded ("..")
	IntervalStart *time.Time
	IntervalEnd   *time.Time

	// IsInterval distinguishes an interval from an instant
	IsInterval bool
}
