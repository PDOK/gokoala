package cql

import (
	"log"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoop(t *testing.T) {
	// given
	inputCQL := "prop1 = 10 AND prop2 < 5"
	expectedSQL := ""

	// when
	actual, err := ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, []domain.Field{}, 0))

	// then
	require.NoError(t, err)
	require.NotNil(t, actual)
	assert.Empty(t, actual.Params)
	assert.Equal(t, expectedSQL, actual.SQL)
}

// Test CQL examples provided by OGC.
// See https://github.com/opengeospatial/ogcapi-features/tree/64ac2d892b877b711a4570336cb9d42e2afb4ef8/cql2/standard/schema/examples/text
func TestCQLExamplesProvidedByOGC_Postgres(t *testing.T) {
	const (
		ext               = ".txt"
		expectedSuffix    = "_expected_postgres" + ext
		expectedErrSuffix = "_expected_error_postgres" + ext
	)

	ogcExamples := path.Join(pwd, "testdata", "ogc")
	entries, err := os.ReadDir(ogcExamples)
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() ||
			strings.Contains(entry.Name(), "gpkg"+ext) ||
			strings.Contains(entry.Name(), expectedSuffix) ||
			strings.Contains(entry.Name(), expectedErrSuffix) {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			// given
			example, err := os.ReadFile(path.Join(ogcExamples, entry.Name()))
			require.NoError(t, err)

			expectedFile := path.Join(ogcExamples, strings.TrimSuffix(entry.Name(), ext)+expectedSuffix)
			expectedErrFile := path.Join(ogcExamples, strings.TrimSuffix(entry.Name(), ext)+expectedErrSuffix)

			inputCQL := strings.Map(removeNewlinesAndTabs, strings.TrimSpace(string(example)))
			require.NotEmpty(t, inputCQL)
			log.Printf("Parsing CQL: %s", inputCQL)

			if strings.HasPrefix(entry.Name(), "SKIP_") {
				t.Skipf("Skipping %s, since this example is not (yet) supported by our CQL implementation", entry.Name())
			}

			var expectedSQL, expectedErr []byte
			expectedSQL, err = os.ReadFile(expectedFile)
			if os.IsNotExist(err) {
				// no exception file found, assume error is expected
				expectedErr, err = os.ReadFile(expectedErrFile)
				require.NoError(t, err, "file with expected error not found")
			}

			// when
			switch {
			case len(expectedSQL) > 0:
				t.Skip("TODO")
				// TODO: test successful case
			case len(expectedErr) > 0:
				t.Skip("TODO")
				// TODO: test error case
			default:
				require.Fail(t, "expected either an expected SQL result or an expected error, but neither was found")
			}
		})
	}
}
