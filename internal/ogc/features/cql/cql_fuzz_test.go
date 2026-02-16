package cql

import (
	"testing"
	"unicode/utf8"

	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/stretchr/testify/assert"
)

// Test to make sure the parser doesn't crash on invalid input.
// Run with: go test -fuzz=Fuzz -fuzztime=10s -run=^$
func FuzzParseToSQL(f *testing.F) {
	queryables := []string{
		"floors",
		"swimming_pool",
		"updated",
		"geometry",
		"event_time",
		// 'created' is not a queryable
	}

	testcases := []string{
		"floors>5 AND swimming_pool=true",
		"avg(windSpeed)",
		"updated >= date('1970-01-01')",
		"created <= date('2050-01-01')",
		"S_INTERSECTS(geometry,POINT(36.319836 32.288087))",
		"T_INTERSECTS(event_time, INTERVAL('1969-07-16T05:32:00Z', '1969-07-24T16:50:35Z'))",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, input string) {
		// when
		result, _, err := ParseToSQL(input, NewGeoPackageListener(&util.DefaultRandomizer, queryables))

		// then
		assert.Truef(t, utf8.ValidString(result), "valid string")
		if err == nil {
			assert.NotNil(t, result)

			// validate idempotency
			result2, _, err2 := ParseToSQL(input, NewGeoPackageListener(&util.DefaultRandomizer, queryables))
			if err2 != nil {
				assert.NotNil(t, result)
				assert.Equal(t, result, result2)
			}
		}
	})
}
