package query_expansion

import (
	"context"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// change working dir to root
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

// Run with: go test -fuzz=Fuzz -fuzztime=10s -run=^$
func FuzzExpand(f *testing.F) {
	queryExpansion, err := NewQueryExpansion("internal/ogc/features_search/testdata/rewrites.csv", "internal/ogc/features_search/testdata/synonyms.csv")
	require.NoError(f, err)

	testcases := []string{"Foo", "Bar", "Baz", "Den Haag", "Fryslân", "Gouverneurstraat", "West", "1e", "tweede", "Oud", "Oude"}
	for _, tc := range testcases {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, input string) {
		expanded, err := queryExpansion.Expand(context.Background(), input)
		require.NoError(t, err)
		query := expanded.ToExactMatchQuery(true)

		assert.Truef(t, utf8.ValidString(query), "valid string")
		if strings.TrimSpace(input) != "" {
			assert.NotEmpty(t, query)
		}
	})
}
