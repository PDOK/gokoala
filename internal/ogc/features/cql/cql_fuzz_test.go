package cql

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// Run with: go test -fuzz=Fuzz -fuzztime=10s -run=^$
func FuzzParseToSQL(f *testing.F) {
	testcases := []string{
		"floors>5 AND swimming_pool=true",
		"avg(windSpeed)",
		"updated >= date('1970-01-01')",
		"S_INTERSECTS(geometry,POINT(36.319836 32.288087))",
		"T_INTERSECTS(event_time, INTERVAL('1969-07-16T05:32:00Z', '1969-07-24T16:50:35Z'))",
	}
	for _, tc := range testcases {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, input string) {
		result := ParseToSQL(input)
		assert.Truef(t, utf8.ValidString(result), "valid string")
		if strings.TrimSpace(input) != "" {
			assert.NotEmpty(t, result)
		}
	})
}
