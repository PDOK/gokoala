package cql

import (
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/types"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources/geopackage"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	gpkg       = "gpkg"
	postgresql = "postgres"
)

var datasources = []string{gpkg, postgresql}

var cqlConfigAllEnabled = config.CQL{
	Enable:                            types.PtrTo(true),
	EnableAdvancedComparisonOperators: true,
	EnableCaseInsensitiveComparison:   true,
	EnableAccentInsensitiveComparison: true,
	EnableBasicSpatialFunctions:       true,
	EnableBasicSpatialFunctionsPlus:   true,
	EnableSpatialFunctions:            true,
	EnableTemporalFunctions:           true,
}

var pwd string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	pwd = path.Dir(filename)
}

func TestInvalidBooleanQuery(t *testing.T) {
	// given
	inputCQL := "prop1 ==== 1 AND prop2 !!= 5"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var err error

			// when
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, []domain.Field{}, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, []domain.Field{}, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			}

			// then
			require.ErrorContains(t, err, "syntax error at column 7: mismatched input '=' expecting ")
			require.ErrorContains(t, err, "syntax error at column 23: no viable alternative at input 'prop2!'")
		})
	}
}

func TestFailOnNonQueryablePropertyQuery(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}}
	inputCQL := "prop1 = 30 AND prop2 > 77"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var err error

			// when
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			}

			// then
			assert.ErrorContains(t, err, "property 'prop2' cannot be used in CQL filter, is not a queryable property")
		})
	}
}

func TestPreventSQLInjectionAttack(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}}
	inputCQL := "prop1 > 5 OR 1 = 1"
	expectedSQLGeoPackage := "(cast (\"prop1\" as numeric) > :cql_bcde OR cast (:cql_fghi as numeric) = :cql_jklm)"
	expectedSQLPostgres := "(cast (\"prop1\" as numeric) > @cql_bcde OR cast (@cql_fghi as numeric) = @cql_jklm)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error

			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}

			// then
			assert.Equal(t, map[string]any{"cql_bcde": int64(5), "cql_fghi": int64(1), "cql_jklm": int64(1)}, actual.Params)
		})
	}
}

func TestPreventSQLInjectionAttackAdvanced(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "prop5 = 'Square';DROP TABLE cql"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var err error
			// when
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			}
			// then
			assert.ErrorContains(t, err, "syntax error at column 16")
		})
	}
}

func TestBooleanQueryWithNumbers(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 = 10 AND prop2 < 5"
	expectedSQLGeoPackage := "(cast (\"prop1\" as numeric) = :cql_bcde AND cast (\"prop2\" as numeric) < :cql_fghi)"
	expectedSQLPostgres := "(cast (\"prop1\" as numeric) = @cql_bcde AND cast (\"prop2\" as numeric) < @cql_fghi)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": int64(10), "cql_fghi": int64(5)}, actual.Params)
		})
	}
}

func TestAllSimpleComparisionOperators(t *testing.T) {
	// given
	operators := []string{"=", "<", ">", "<=", ">=", "<>"} // note '!=' is not valid CQL, but '<>' is used by CQL.
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}

	for _, operator := range operators {
		t.Run(operator, func(t *testing.T) {
			for _, datasource := range datasources {
				t.Run(datasource, func(t *testing.T) {
					var actual *SQLResult
					var err error
					switch datasource {
					case gpkg:
						expectedSQL := "cast (\"prop1\" as numeric) " + operator + " :cql_bcde"

						// when
						actual, err = ParseToSQL("prop1 "+operator+" 10", NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

						// then
						require.NoError(t, err)
						assertValidSQLiteQuery(t, actual)
						assert.Equal(t, expectedSQL, actual.SQL)
					case postgresql:
						expectedSQL := "cast (\"prop1\" as numeric) " + operator + " @cql_bcde"

						// when
						actual, err = ParseToSQL("prop1 "+operator+" 10", NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

						// then
						require.NoError(t, err)
						assert.Equal(t, expectedSQL, actual.SQL)
					}
					assert.Equal(t, map[string]any{"cql_bcde": int64(10)}, actual.Params)
				})
			}
		})
	}
}

func TestMultipleBooleanQueries(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "(prop1 = 10 OR prop1 = 20) AND NOT (prop2 = 'X')"
	expectedSQLGeoPackage := "((cast (\"prop1\" as numeric) = :cql_bcde OR cast (\"prop1\" as numeric) = :cql_fghi) AND NOT (\"prop2\" = :cql_jklm))"
	expectedSQLPostgres := "((cast (\"prop1\" as numeric) = @cql_bcde OR cast (\"prop1\" as numeric) = @cql_fghi) AND NOT (\"prop2\" = @cql_jklm))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": int64(10), "cql_fghi": int64(20), "cql_jklm": "X"}, actual.Params)
		})
	}
}

func TestBooleanTrueLiteral(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "(prop1 = true AND prop2 = 20)"
	expectedSQLGeoPackage := "(\"prop1\" = 1 AND cast (\"prop2\" as numeric) = :cql_bcde)"
	expectedSQLPostgres := "(\"prop1\" = true AND cast (\"prop2\" as numeric) = @cql_bcde)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": int64(20)}, actual.Params)
		})
	}
}

func TestBooleanFalseLiteral(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "(prop1 = false AND prop2 = 20)"
	expectedSQLGeoPackage := "(\"prop1\" = 0 AND cast (\"prop2\" as numeric) = :cql_bcde)"
	expectedSQLPostgres := "(\"prop1\" = false AND cast (\"prop2\" as numeric) = @cql_bcde)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": int64(20)}, actual.Params)
		})
	}
}

func TestMultipleBooleanQueriesWithStrings(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}, {Name: "prop3"}}
	inputCQL := "(prop1 = 'foo' AND prop2 = 'bar') OR prop3 = 'abc'"
	expectedSQLGeoPackage := "((\"prop1\" = :cql_bcde AND \"prop2\" = :cql_fghi) OR \"prop3\" = :cql_jklm)"
	expectedSQLPostgres := "((\"prop1\" = @cql_bcde AND \"prop2\" = @cql_fghi) OR \"prop3\" = @cql_jklm)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "foo", "cql_fghi": "bar", "cql_jklm": "abc"}, actual.Params)
		})
	}
}

func TestLikeOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}, {Name: "prop3"}}
	inputCQL := "prop1 LIKE 'foo%' AND prop2 LIKE 'bar_' OR prop3 LIKE '%abc'"
	expectedSQLGeoPackage := "((\"prop1\" LIKE :cql_bcde AND \"prop2\" LIKE :cql_fghi) OR \"prop3\" LIKE :cql_jklm)"
	expectedSQLPostgres := "((\"prop1\" LIKE @cql_bcde AND \"prop2\" LIKE @cql_fghi) OR \"prop3\" LIKE @cql_jklm)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "foo%", "cql_fghi": "bar_", "cql_jklm": "%abc"}, actual.Params)
		})
	}
}

func TestNotLikeOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}, {Name: "prop3"}}
	inputCQL := "prop1 NOT LIKE 'foo%' AND prop2 LIKE 'bar_' OR prop3 LIKE '%abc'"
	expectedSQLGeoPackage := "((\"prop1\" NOT LIKE :cql_bcde AND \"prop2\" LIKE :cql_fghi) OR \"prop3\" LIKE :cql_jklm)"
	expectedSQLPostgres := "((\"prop1\" NOT LIKE @cql_bcde AND \"prop2\" LIKE @cql_fghi) OR \"prop3\" LIKE @cql_jklm)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "foo%", "cql_fghi": "bar_", "cql_jklm": "%abc"}, actual.Params)
		})
	}
}

func TestOperatorShouldWorkRegardlessOfCasing(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}, {Name: "prop3"}}
	inputCQL := "prop1 lIkE 'foo%' AND prop2 LiKE 'bar_' OR prop3 like '%abc'"
	expectedSQLGeoPackage := "((\"prop1\" LIKE :cql_bcde AND \"prop2\" LIKE :cql_fghi) OR \"prop3\" LIKE :cql_jklm)"
	expectedSQLPostgres := "((\"prop1\" LIKE @cql_bcde AND \"prop2\" LIKE @cql_fghi) OR \"prop3\" LIKE @cql_jklm)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "foo%", "cql_fghi": "bar_", "cql_jklm": "%abc"}, actual.Params)
		})
	}
}

func TestLikeOperatorNotEnabled(t *testing.T) {
	// given
	cqlConfig := cqlConfigAllEnabled
	cqlConfig.EnableAdvancedComparisonOperators = false

	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}, {Name: "prop3"}}
	inputCQL := "prop1 LIKE 'foo%' AND prop2 LIKE 'bar_' OR prop3 LIKE '%abc'"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			}

			// then
			assert.ErrorContains(t, err, "advanced comparison operators (LIKE, BETWEEN, IN, IS NULL) are not enabled for this collection")
		})
	}
}

func TestCaseInsensitiveOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}}
	inputCQL := "CASEI(prop1) = CASEI('Foo')"
	expectedSQLGeoPackage := "\"prop1\" COLLATE NOCASE = :cql_bcde COLLATE NOCASE"
	expectedSQLPostgres := "\"prop1\" COLLATE NOCASE = @cql_bcde COLLATE NOCASE"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}

			// then
			assert.Equal(t, map[string]any{"cql_bcde": "Foo"}, actual.Params)
		})
	}
}

func TestCaseInsensitiveOperatorNotEnabled(t *testing.T) {
	// given
	cqlConfig := cqlConfigAllEnabled
	cqlConfig.EnableCaseInsensitiveComparison = false

	queryables := []domain.Field{{Name: "prop1"}}
	inputCQL := "CASEI(prop1) = CASEI('Foo')"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			}

			// then
			assert.ErrorContains(t, err, "case-insensitive comparison (CASEI) is not enabled for this collection")
		})
	}
}

func TestAccentInsensitiveOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}}
	inputCQL := "ACCENTI(prop1) = ACCENTI('fóo') OR ACCENTI(prop1) = ACCENTI('débárquér')"
	expectedSQLGeoPackage := "(\"prop1\" COLLATE NOACCENT = :cql_bcde COLLATE NOACCENT OR \"prop1\" COLLATE NOACCENT = :cql_fghi COLLATE NOACCENT)"
	expectedSQLPostgres := "(\"prop1\" COLLATE NOACCENT = @cql_bcde COLLATE NOACCENT OR \"prop1\" COLLATE NOACCENT = @cql_fghi COLLATE NOACCENT)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "fóo", "cql_fghi": "débárquér"}, actual.Params)
		})
	}
}

func TestAccentInsensitiveOperatorNotEnabled(t *testing.T) {
	// given
	cqlConfig := cqlConfigAllEnabled
	cqlConfig.EnableAccentInsensitiveComparison = false

	queryables := []domain.Field{{Name: "prop1"}}
	inputCQL := "ACCENTI(prop1) = ACCENTI('fóo') OR ACCENTI(prop1) = ACCENTI('débárquér')"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			}

			// then
			assert.ErrorContains(t, err, "accent-insensitive comparison (ACCENTI) is not enabled for this collection")
		})
	}
}

func TestNestedCaseAndAccentInsensitiveOperators(t *testing.T) {
	tests := []struct {
		name        string
		inputCQL    string
		expectedSQL string
	}{
		{
			name:        "CASEI around ACCENTI",
			inputCQL:    "CASEI(ACCENTI(prop1)) = CASEI(ACCENTI('Fóo'))",
			expectedSQL: "\"prop1\" COLLATE NOACCENT_NOCASE = :cql_bcde COLLATE NOACCENT_NOCASE",
		},
		{
			name:        "ACCENTI around CASEI",
			inputCQL:    "ACCENTI(CASEI(prop1)) = ACCENTI(CASEI('Fóo'))",
			expectedSQL: "\"prop1\" COLLATE NOACCENT_NOCASE = :cql_bcde COLLATE NOACCENT_NOCASE",
		},
		{
			name:        "Mixed up",
			inputCQL:    "CASEI(ACCENTI(prop1)) = ACCENTI(CASEI('Fóo'))",
			expectedSQL: "\"prop1\" COLLATE NOACCENT_NOCASE = :cql_bcde COLLATE NOACCENT_NOCASE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			queryables := []domain.Field{{Name: "prop1"}}

			for _, datasource := range datasources {
				t.Run(datasource, func(t *testing.T) {
					var actual *SQLResult
					var err error
					switch datasource {
					case gpkg:
						// when
						actual, err = ParseToSQL(tt.inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

						// then
						require.NoError(t, err)
						assertValidSQLiteQuery(t, actual)
						assert.Equal(t, tt.expectedSQL, actual.SQL)
					case postgresql:
						expectedSQLPostgres := strings.ReplaceAll(tt.expectedSQL, ":", "@")

						// when
						actual, err = ParseToSQL(tt.inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

						// then
						require.NoError(t, err)
						assert.Equal(t, expectedSQLPostgres, actual.SQL)
					}
					assert.Equal(t, map[string]any{"cql_bcde": "Fóo"}, actual.Params)
				})
			}
		})
	}
}

func TestCaseAndAccentInsensitiveOperatorWithLike(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}}
	inputCQL := "CASEI(prop1) LIKE CASEI('Foo%') AND ACCENTI(CASEI(prop1)) LIKE ACCENTI(CASEI('Fóo%'))"
	expectedSQLGeoPackage := "(\"prop1\" COLLATE NOCASE LIKE :cql_bcde COLLATE NOCASE AND \"prop1\" COLLATE NOACCENT_NOCASE LIKE :cql_fghi COLLATE NOACCENT_NOCASE)"
	expectedSQLPostgres := "(\"prop1\" ILIKE @cql_bcde AND unaccent(\"prop1\") ILIKE unaccent(@cql_fghi))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "Foo%", "cql_fghi": "Fóo%"}, actual.Params)
		})
	}
}

func TestLikeOperatorFailOnMissingWildcard(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}}
	inputCQL := "prop1 LIKE 'foo'"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			}

			// then
			assert.ErrorContains(t, err, "LIKE pattern is missing wildcard symbol. "+
				"Either percentage '%' to match multiple characters or underscore '_' to match a "+
				"single character can be used as a wildcard symbol. For example: LIKE 'foo%'.")
		})
	}
}

func TestBetweenOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 BETWEEN 4 AND 6 AND prop2 = 'bar'"
	expectedSQLGeoPackage := "(\"prop1\" BETWEEN :cql_bcde AND :cql_fghi AND \"prop2\" = :cql_jklm)"
	expectedSQLPostgres := "(\"prop1\" BETWEEN @cql_bcde AND @cql_fghi AND \"prop2\" = @cql_jklm)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": int64(4), "cql_fghi": int64(6), "cql_jklm": "bar"}, actual.Params)
		})
	}
}

func TestNotBetweenOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 NOT BETWEEN 4 AND 6 AND prop2 = 'bar'"
	expectedSQLGeoPackage := "(\"prop1\" NOT BETWEEN :cql_bcde AND :cql_fghi AND \"prop2\" = :cql_jklm)"
	expectedSQLPostgres := "(\"prop1\" NOT BETWEEN @cql_bcde AND @cql_fghi AND \"prop2\" = @cql_jklm)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": int64(4), "cql_fghi": int64(6), "cql_jklm": "bar"}, actual.Params)
		})
	}
}

func TestBetweenOperatorNotEnabled(t *testing.T) {
	// given
	cqlConfig := cqlConfigAllEnabled
	cqlConfig.EnableAdvancedComparisonOperators = false

	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 NOT BETWEEN 4 AND 6 AND prop2 = 'bar'"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			}

			// then
			assert.ErrorContains(t, err, "advanced comparison operators (LIKE, BETWEEN, IN, IS NULL) are not enabled for this collection")
		})
	}
}

func TestInListOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 IN ('foo', 'bar', 'baz') AND prop2 = 'baz'"
	expectedSQLGeoPackage := "(\"prop1\" IN (:cql_bcde, :cql_fghi, :cql_jklm) AND \"prop2\" = :cql_nopq)"
	expectedSQLPostgres := "(\"prop1\" IN (@cql_bcde, @cql_fghi, @cql_jklm) AND \"prop2\" = @cql_nopq)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "foo", "cql_fghi": "bar", "cql_jklm": "baz", "cql_nopq": "baz"}, actual.Params)
		})
	}
}

func TestNotInListOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 NOT IN ('foo', 'bar', 'baz') AND prop2 = 'baz'"
	expectedSQLGeoPackage := "(\"prop1\" NOT IN (:cql_bcde, :cql_fghi, :cql_jklm) AND \"prop2\" = :cql_nopq)"
	expectedSQLPostgres := "(\"prop1\" NOT IN (@cql_bcde, @cql_fghi, @cql_jklm) AND \"prop2\" = @cql_nopq)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "foo", "cql_fghi": "bar", "cql_jklm": "baz", "cql_nopq": "baz"}, actual.Params)
		})
	}
}

func TestInOperatorNotEnabled(t *testing.T) {
	// given
	cqlConfig := cqlConfigAllEnabled
	cqlConfig.EnableAdvancedComparisonOperators = false

	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 IN ('foo', 'bar', 'baz') AND prop2 = 'baz'"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			}

			// then
			assert.ErrorContains(t, err, "advanced comparison operators (LIKE, BETWEEN, IN, IS NULL) are not enabled for this collection")
		})
	}
}

func TestIsNullOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 IS NULL AND prop2 = 'baz'"
	expectedSQLGeoPackage := "(\"prop1\" IS NULL AND \"prop2\" = :cql_bcde)"
	expectedSQLPostgres := "(\"prop1\" IS NULL AND \"prop2\" = @cql_bcde)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "baz"}, actual.Params)
		})
	}
}

func TestIsNotNullOperator(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 IS NOT NULL AND prop2 = 'baz'"
	expectedSQLGeoPackage := "(\"prop1\" IS NOT NULL AND \"prop2\" = :cql_bcde)"
	expectedSQLPostgres := "(\"prop1\" IS NOT NULL AND \"prop2\" = @cql_bcde)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "baz"}, actual.Params)
		})
	}
}

func TestIsNullOperatorNotEnabled(t *testing.T) {
	// given
	cqlConfig := cqlConfigAllEnabled
	cqlConfig.EnableAdvancedComparisonOperators = false

	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 IS NULL AND prop2 = 'baz'"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			}

			// then
			assert.ErrorContains(t, err, "advanced comparison operators (LIKE, BETWEEN, IN, IS NULL) are not enabled for this collection")
		})
	}
}

func TestFailOnInvalidInListQuery(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop2"}}
	inputCQL := "prop1 IN ('foo', 'bar' 'baz')"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			}

			// then
			assert.ErrorContains(t, err, "syntax error at column 23: extraneous input ''baz'' expecting {')', ','}")
		})
	}
}

func TestSpatialQueryWithPoint(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, POINT(4.897 52.377))"
	expectedSQLGeoPackage := "ST_Intersects(CastAutomagic(\"geom\"), SwapCoords(ST_GeomFromText(:cql_bcde, 4326)))"
	expectedSQLPostgres := "ST_Intersects(\"geom\", ST_FlipCoordinates(ST_Transform(ST_GeomFromText(@cql_bcde, 4326), ST_SRID(\"geom\"))))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "POINT(4.897 52.377)"}, actual.Params)
		})
	}
}

func TestSpatialQueryWithPoint3D(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, POINTZ(4.897 52.377 10.0))"
	expectedSQLGeoPackage := "ST_Intersects(CastAutomagic(\"geom\"), SwapCoords(ST_GeomFromText(:cql_bcde, 4326)))"
	expectedSQLPostgres := "ST_Intersects(\"geom\", ST_FlipCoordinates(ST_Transform(ST_GeomFromText(@cql_bcde, 4326), ST_SRID(\"geom\"))))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "POINTZ(4.897 52.377 10.0)"}, actual.Params)
		})
	}
}

func TestSpatialQueryWithLinestring(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, LINESTRING(0.0 0.0, 1.0 1.0, 2.0 0.0))"
	expectedSQLGeoPackage := "ST_Intersects(CastAutomagic(\"geom\"), SwapCoords(ST_GeomFromText(:cql_bcde, 4326)))"
	expectedSQLPostgres := "ST_Intersects(\"geom\", ST_FlipCoordinates(ST_Transform(ST_GeomFromText(@cql_bcde, 4326), ST_SRID(\"geom\"))))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "LINESTRING(0.0 0.0, 1.0 1.0, 2.0 0.0)"}, actual.Params)
		})
	}
}

func TestSpatialQueryWithPolygon(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, POLYGON((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0)))"
	expectedSQLGeoPackage := "ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 28992))"
	expectedSQLPostgres := "ST_Intersects(\"geom\", ST_Transform(ST_GeomFromText(@cql_bcde, 28992), ST_SRID(\"geom\")))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 28992, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 28992, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "POLYGON((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0))"}, actual.Params)
		})
	}
}

func TestSpatialQueryWithPolygonWithHole(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, POLYGON((0.0 0.0, 10.0 0.0, 10.0 10.0, 0.0 10.0, 0.0 0.0),(2.0 2.0, 8.0 2.0, 8.0 8.0, 2.0 8.0, 2.0 2.0)))"
	expectedSQLGeoPackage := "ST_Intersects(CastAutomagic(\"geom\"), SwapCoords(ST_GeomFromText(:cql_bcde, 4326)))"
	expectedSQLPostgres := "ST_Intersects(\"geom\", ST_FlipCoordinates(ST_Transform(ST_GeomFromText(@cql_bcde, 4326), ST_SRID(\"geom\"))))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "POLYGON((0.0 0.0, 10.0 0.0, 10.0 10.0, 0.0 10.0, 0.0 0.0), (2.0 2.0, 8.0 2.0, 8.0 8.0, 2.0 8.0, 2.0 2.0))"}, actual.Params)
		})
	}
}

func TestSpatialQueryWithMultiPoint(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, MULTIPOINT(0.0 0.0, 1.0 1.0))"
	expectedSQLGeoPackage := "ST_Intersects(CastAutomagic(\"geom\"), SwapCoords(ST_GeomFromText(:cql_bcde, 4326)))"
	expectedSQLPostgres := "ST_Intersects(\"geom\", ST_FlipCoordinates(ST_Transform(ST_GeomFromText(@cql_bcde, 4326), ST_SRID(\"geom\"))))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "MULTIPOINT(0.0 0.0, 1.0 1.0)"}, actual.Params)
		})
	}
}

func TestSpatialQueryWithMultiLinestring(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, MULTILINESTRING((0.0 0.0, 1.0 1.0),(2.0 2.0, 3.0 3.0)))"
	expectedSQLGeoPackage := "ST_Intersects(CastAutomagic(\"geom\"), SwapCoords(ST_GeomFromText(:cql_bcde, 4326)))"
	expectedSQLPostgres := "ST_Intersects(\"geom\", ST_FlipCoordinates(ST_Transform(ST_GeomFromText(@cql_bcde, 4326), ST_SRID(\"geom\"))))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "MULTILINESTRING((0.0 0.0, 1.0 1.0), (2.0 2.0, 3.0 3.0))"}, actual.Params)
		})
	}
}

func TestSpatialQueryWithMultiPolygon(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, MULTIPOLYGON(((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0)),((2.0 2.0, 3.0 2.0, 3.0 3.0, 2.0 3.0, 2.0 2.0))))"
	expectedSQLGeoPackage := "ST_Intersects(CastAutomagic(\"geom\"), SwapCoords(ST_GeomFromText(:cql_bcde, 4326)))"
	expectedSQLPostgres := "ST_Intersects(\"geom\", ST_FlipCoordinates(ST_Transform(ST_GeomFromText(@cql_bcde, 4326), ST_SRID(\"geom\"))))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "MULTIPOLYGON(((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0)), ((2.0 2.0, 3.0 2.0, 3.0 3.0, 2.0 3.0, 2.0 2.0)))"}, actual.Params)
		})
	}
}

func TestSpatialQueryWithMultiPolygonFailsWithOnlyBasicSpatialFilteringEnabled(t *testing.T) {
	// given
	cqlConfig := cqlConfigAllEnabled
	cqlConfig.EnableBasicSpatialFunctionsPlus = false
	cqlConfig.EnableSpatialFunctions = false

	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, MULTIPOLYGON(((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0)),((2.0 2.0, 3.0 2.0, 3.0 3.0, 2.0 3.0, 2.0 2.0))))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfig))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfig))
			}

			// then
			assert.ErrorContains(t, err, "geometry type 'MULTIPOLYGON' is not allowed, only POINT and BBOX are allowed with basic spatial filtering")
		})
	}
}

func TestSpatialQueryWithGeometryCollection(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, GEOMETRYCOLLECTION(POINT(0.0 0.0),LINESTRING(0.0 0.0, 1.0 1.0)))"
	expectedSQLGeoPackage := "ST_Intersects(CastAutomagic(\"geom\"), SwapCoords(ST_GeomFromText(:cql_bcde, 4326)))"
	expectedSQLPostgres := "ST_Intersects(\"geom\", ST_FlipCoordinates(ST_Transform(ST_GeomFromText(@cql_bcde, 4326), ST_SRID(\"geom\"))))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "GEOMETRYCOLLECTION(POINT(0.0 0.0), LINESTRING(0.0 0.0, 1.0 1.0))"}, actual.Params)
		})
	}
}

func TestSpatialQueryWithBbox(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, BBOX(10.0, 20.1, 30.0, 40.0))"
	expectedSQLGeoPackage := "ST_Intersects(CastAutomagic(\"geom\"), SwapCoords(BuildMbr(:cql_bcde, :cql_fghi, :cql_jklm, :cql_nopq, 4326)))"
	expectedSQLPostgres := "ST_Intersects(\"geom\", ST_FlipCoordinates(ST_Transform(ST_MakeEnvelope(@cql_bcde, @cql_fghi, @cql_jklm, @cql_nopq, 4326), ST_SRID(\"geom\"))))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": 10.0, "cql_fghi": 20.1, "cql_jklm": 30.0, "cql_nopq": 40.0}, actual.Params)
		})
	}
}

func TestSpatialQueryFailsOnInvalidBbox(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, BBOX(10.0 20.1 30.0 40.0))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))
			}

			// then
			assert.ErrorContains(t, err, "missing east bound coordinate (maxx) in bounding box")
		})
	}
}

func TestSpatialQueryWithBboxWithOnlyBasicSpatialFunctionsEnabled(t *testing.T) {
	// given
	cqlConfig := cqlConfigAllEnabled
	cqlConfig.EnableBasicSpatialFunctionsPlus = false
	cqlConfig.EnableSpatialFunctions = false

	queryables := []domain.Field{{Name: "geometry", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, BBOX(10.0, 20.1, 30.0, 40.0))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfig))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfig))
			}

			// then
			require.NoError(t, err)
		})
	}
}

func TestSpatialQueryFailsOnInvalidBboxWithText(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, BBOX(10.0, 'bla'))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))
			}

			// then
			assert.ErrorContains(t, err, "'bla' is not a valid numeric type")
		})
	}
}

func TestSpatialQueryFailsOnNonGeometryProperty(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geom, BBOX(10.0, 20.1, 30.0, 40.0))" // should be geometry instead of geom

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))
			}

			// then
			assert.ErrorContains(t, err, "spatial filtering is only supported on property 'geometry'")
		})
	}
}

func TestSpatialQueryFailsOnUndefinedGeometry(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geometry", IsPrimaryGeometry: false}} // not marked as primary geometry
	inputCQL := "S_INTERSECTS(geometry, BBOX(10.0, 20.1, 30.0, 40.0))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))
			}

			// then
			assert.ErrorContains(t, err, "spatial filtering is not supported for this collection since there is no geometry field defined")
		})
	}
}

func TestSpatialQueryUsesRtree(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, BBOX(10.0, 20.1, 30.0, 40.0))"

	// when
	actual, err := ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

	// then
	require.NoError(t, err)
	normalizedRtreeClause := strings.Join(strings.Fields(strings.Map(removeNewlinesAndTabs, actual.RtreeSQL)), " ")
	assert.Equal(t, "AND EXISTS (SELECT 1 FROM rtree_%[1]s_%[2]s r WHERE r.minx <= ST_MaxX(SwapCoords(BuildMbr(:cql_bcde, :cql_fghi, :cql_jklm, :cql_nopq, 4326))) AND r.maxx >= ST_MinX(SwapCoords(BuildMbr(:cql_bcde, :cql_fghi, :cql_jklm, :cql_nopq, 4326))) AND r.miny <= ST_MaxY(SwapCoords(BuildMbr(:cql_bcde, :cql_fghi, :cql_jklm, :cql_nopq, 4326))) AND r.maxy >= ST_MinY(SwapCoords(BuildMbr(:cql_bcde, :cql_fghi, :cql_jklm, :cql_nopq, 4326))))", normalizedRtreeClause)
}

func TestSpatialQueryWithGeometryAndBooleanFilter(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "geom", IsPrimaryGeometry: true}}
	inputCQL := "prop1 = 'foo' AND S_INTERSECTS(geometry, POINT(4.897 52.377))"
	expectedSQLGeoPackage := "(\"prop1\" = :cql_bcde AND ST_Intersects(CastAutomagic(\"geom\"), SwapCoords(ST_GeomFromText(:cql_fghi, 4326))))"
	expectedSQLPostgres := "(\"prop1\" = @cql_bcde AND ST_Intersects(\"geom\", ST_FlipCoordinates(ST_Transform(ST_GeomFromText(@cql_fghi, 4326), ST_SRID(\"geom\")))))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "foo", "cql_fghi": "POINT(4.897 52.377)"}, actual.Params)
		})
	}
}

func TestSpatialQueryWithAllSpatialFunctions(t *testing.T) {
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}

	for cqlFunc, sqlFunc := range spatialFunctions {
		t.Run(cqlFunc, func(t *testing.T) {
			for _, datasource := range datasources {

				/// test for each spatial function, per datasource
				t.Run(datasource, func(t *testing.T) {
					inputCQL := cqlFunc + "(geometry, POINT(4.897 52.377))"
					var actual *SQLResult
					var err error
					switch datasource {
					case gpkg:
						// given
						expectedSQL := sqlFunc + "(CastAutomagic(\"geom\"), SwapCoords(ST_GeomFromText(:cql_bcde, 4326)))"

						// when
						actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

						// then
						require.NoError(t, err)
						assert.Equal(t, expectedSQL, actual.SQL)
					case postgresql:
						// given
						expectedSQL := sqlFunc + "(\"geom\", ST_FlipCoordinates(ST_Transform(ST_GeomFromText(@cql_bcde, 4326), ST_SRID(\"geom\"))))"

						// when
						actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderYX, geospatial.Features, cqlConfigAllEnabled))

						// then
						require.NoError(t, err)
						assert.Equal(t, expectedSQL, actual.SQL)
					}
					assert.Equal(t, map[string]any{"cql_bcde": "POINT(4.897 52.377)"}, actual.Params)
				})
			}
		})
	}
}

func TestSpatialQueryForAllWellKnownTexts(t *testing.T) {
	queryables := []domain.Field{{Name: "geom", IsPrimaryGeometry: true}}
	wkts := []string{
		"POINT(10 20)",
		"POINT (10 20)",
		"POINT Z(10 20 5)",
		"POINT Z (10 20 5)",
		"LINESTRING(0 0, 10 10)",
		"LINESTRING (0 0, 10 10)",
		"LINESTRING Z(0 0 0, 10 10 5)",
		"LINESTRING Z (0 0 0, 10 10 5)",
		"POLYGON((0 0, 10 0, 10 10, 0 10, 0 0))",
		"POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
		"POLYGON((0 0, 10 0, 10 10, 0 10, 0 0), (3 3, 7 3, 7 7, 3 7, 3 3))",
		"POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0), (3 3, 7 3, 7 7, 3 7, 3 3))",
		"POLYGON Z((0 0 0, 10 0 1, 10 10 2, 0 10 1, 0 0 0))",
		"POLYGON Z ((0 0 0, 10 0 1, 10 10 2, 0 10 1, 0 0 0))",
		"MULTIPOINT(10 20, 30 40)",
		"MULTIPOINT (10 20, 30 40)",
		"MULTIPOINT((10 20), (30 40))",
		"MULTIPOINT ((10 20), (30 40))",
		"MULTIPOINT Z (10 20 5, 30 40 7)",
		"MULTIPOINT Z(10 20 5, 30 40 7)",
		"MULTIPOINT Z ((10 20 5), (30 40 7))",
		"MULTIPOINT Z((10 20 5), (30 40 7))",
		"MULTILINESTRING((0 0, 10 10), (20 20, 30 10))",
		"MULTILINESTRING ((0 0, 10 10), (20 20, 30 10))",
		"MULTILINESTRING Z((0 0 0, 10 10 5), (20 20 1, 30 10 2))",
		"MULTILINESTRING Z ((0 0 0, 10 10 5), (20 20 1, 30 10 2))",
		"MULTIPOLYGON(((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 20, 30 20, 30 30, 20 30, 20 20)))",
		"MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 20, 30 20, 30 30, 20 30, 20 20)))",
		"MULTIPOLYGON(((0 0, 10 0, 10 10, 0 10, 0 0), (3 3, 7 3, 7 7, 3 7, 3 3)))",
		"MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0), (3 3, 7 3, 7 7, 3 7, 3 3)))",
		"MULTIPOLYGON Z(((0 0 0, 10 0 1, 10 10 2, 0 10 1, 0 0 0)))",
		"MULTIPOLYGON Z (((0 0 0, 10 0 1, 10 10 2, 0 10 1, 0 0 0)))",
		"GEOMETRYCOLLECTION(POINT(10 20), LINESTRING(0 0, 10 10), POLYGON((0 0, 10 0, 10 10, 0 10, 0 0)))",
		"GEOMETRYCOLLECTION (POINT (10 20), LINESTRING (0 0, 10 10), POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)))",
		"GEOMETRYCOLLECTION Z(POINT(10 20 5), LINESTRING(0 0 0, 10 10 5))",
		"GEOMETRYCOLLECTION Z (POINT (10 20 5), LINESTRING (0 0 0, 10 10 5))",
	}

	for _, wkt := range wkts {
		t.Run(wkt, func(t *testing.T) {
			for _, datasource := range datasources {
				t.Run(datasource, func(t *testing.T) {
					inputCQL := "S_INTERSECTS(geometry, " + wkt + ")"
					var actual *SQLResult
					var err error
					switch datasource {
					case gpkg:
						// given
						expectedSQL := "ST_Intersects(CastAutomagic(\"geom\"), ST_GeomFromText(:cql_bcde, 4326))"

						// when
						actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

						// then
						require.NoError(t, err)
						assert.Equal(t, expectedSQL, actual.SQL)
					case postgresql:
						// given
						expectedSQL := "ST_Intersects(\"geom\", ST_Transform(ST_GeomFromText(@cql_bcde, 4326), ST_SRID(\"geom\")))"

						// when
						actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 4326, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

						// then
						require.NoError(t, err)
						assert.Equal(t, expectedSQL, actual.SQL)
					}
					assert.Equal(t, map[string]any{"cql_bcde": wkt}, actual.Params)
				})
			}
		})
	}
}

func TestBasicSpatialOperatorNotEnabled(t *testing.T) {
	// given
	cqlConfig := cqlConfigAllEnabled
	cqlConfig.EnableBasicSpatialFunctions = false
	cqlConfig.EnableBasicSpatialFunctionsPlus = false
	cqlConfig.EnableSpatialFunctions = false

	queryables := []domain.Field{{Name: "geometry", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, GEOMETRYCOLLECTION(POINT(0.0 0.0),LINESTRING(0.0 0.0, 1.0 1.0)))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			}

			// then
			assert.ErrorContains(t, err, "spatial operators are not enabled for this collection")
		})
	}
}

func TestBasicSpatialPlusOperatorNotEnabled(t *testing.T) {
	// given
	cqlConfig := cqlConfigAllEnabled
	cqlConfig.EnableBasicSpatialFunctions = true
	cqlConfig.EnableBasicSpatialFunctionsPlus = false
	cqlConfig.EnableSpatialFunctions = false

	queryables := []domain.Field{{Name: "geometry", IsPrimaryGeometry: true}}
	inputCQL := "S_INTERSECTS(geometry, GEOMETRYCOLLECTION(POINT(0.0 0.0),LINESTRING(0.0 0.0, 1.0 1.0)))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			}

			// then
			assert.ErrorContains(t, err, "geometry type 'GEOMETRYCOLLECTION' is not allowed, only POINT and BBOX are allowed with basic spatial filtering")
		})
	}
}

func TestAllSpatialFunctionsNotEnabled(t *testing.T) {
	// given
	cqlConfig := cqlConfigAllEnabled
	cqlConfig.EnableBasicSpatialFunctions = true
	cqlConfig.EnableBasicSpatialFunctionsPlus = true
	cqlConfig.EnableSpatialFunctions = false

	queryables := []domain.Field{{Name: "geometry", IsPrimaryGeometry: true}}
	inputCQL := "S_OVERLAPS(geometry, GEOMETRYCOLLECTION(POINT(0.0 0.0),LINESTRING(0.0 0.0, 1.0 1.0)))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			}

			// then
			assert.ErrorContains(t, err, "spatial operator 'S_OVERLAPS' is not enabled for this collection, only S_INTERSECTS is allowed")
		})
	}
}

func TestTemporalAfterWithDate(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_AFTER(prop5, DATE('2015-01-01'))"
	expectedSQLGeoPackage := "\"prop5\" > :cql_bcde"
	expectedSQLPostgres := "\"prop5\" > @cql_bcde"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2015-01-01"}, actual.Params)
		})
	}
}

func TestTemporalAfterWithTimestamp(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop6"}}
	inputCQL := "T_AFTER(prop6, TIMESTAMP('2020-01-01T00:00:00Z'))"
	expectedSQLGeoPackage := "\"prop6\" > :cql_bcde"
	expectedSQLPostgres := "\"prop6\" > @cql_bcde"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01T00:00:00Z"}, actual.Params)
		})
	}
}

func TestTemporalAfterWithInterval(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop6"}}
	inputCQL := "T_AFTER(prop6, INTERVAL('..', '2020-01-01T00:00:00Z'))"
	expectedSQLGeoPackage := "\"prop6\" > :cql_bcde"
	expectedSQLPostgres := "\"prop6\" > @cql_bcde"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01T00:00:00Z"}, actual.Params)
		})
	}
}

func TestTemporalAfterIntervalToInterval(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop9"}, {Name: "prop10"}}
	inputCQL := "T_AFTER(INTERVAL(prop9, prop10), INTERVAL('2020-01-01', '2025-12-31'))"
	expectedSQLGeoPackage := "\"prop9\" > :cql_fghi"
	expectedSQLPostgres := "\"prop9\" > @cql_fghi"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01", "cql_fghi": "2025-12-31"}, actual.Params)
		})
	}
}

func TestTemporalBeforeWithDate(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_BEFORE(prop5, DATE('2027-01-01'))"
	expectedSQLGeoPackage := "\"prop5\" < :cql_bcde"
	expectedSQLPostgres := "\"prop5\" < @cql_bcde"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2027-01-01"}, actual.Params)
		})
	}
}

func TestTemporalBeforeWithTimestamp(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_BEFORE(prop5, TIMESTAMP('2020-01-01T00:00:00Z'))"
	expectedSQLGeoPackage := "\"prop5\" < :cql_bcde"
	expectedSQLPostgres := "\"prop5\" < @cql_bcde"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01T00:00:00Z"}, actual.Params)
		})
	}
}

func TestTemporalBeforeWithInterval(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_BEFORE(prop5, INTERVAL('2020-01-01T00:00:00Z', '..'))"
	expectedSQLGeoPackage := "\"prop5\" < :cql_bcde"
	expectedSQLPostgres := "\"prop5\" < @cql_bcde"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01T00:00:00Z"}, actual.Params)
		})
	}
}

func TestTemporalEqualsWithDate(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_EQUALS(prop5, DATE('2026-02-12'))"
	expectedSQLGeoPackage := "\"prop5\" = :cql_bcde"
	expectedSQLPostgres := "\"prop5\" = @cql_bcde"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2026-02-12"}, actual.Params)
		})
	}
}

func TestTemporalEqualsWithTimestamp(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_EQUALS(prop5, TIMESTAMP('2020-01-01T00:00:00Z'))"
	expectedSQLGeoPackage := "\"prop5\" = :cql_bcde"
	expectedSQLPostgres := "\"prop5\" = @cql_bcde"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01T00:00:00Z"}, actual.Params)
		})
	}
}

func TestTemporalEqualsWithInterval(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_EQUALS(prop5, INTERVAL('2020-01-01T00:00:00Z', '2030-01-01T00:00:00Z'))"
	expectedSQLGeoPackage := "(\"prop5\" = :cql_bcde AND \"prop5\" = :cql_fghi)"
	expectedSQLPostgres := "(\"prop5\" = @cql_bcde AND \"prop5\" = @cql_fghi)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01T00:00:00Z", "cql_fghi": "2030-01-01T00:00:00Z"}, actual.Params)
		})
	}
}

func TestTemporalEqualsIntervalToInterval(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop9"}, {Name: "prop10"}}
	inputCQL := "T_EQUALS(INTERVAL(prop9, prop10), INTERVAL('2026-02-13', '2026-02-28'))"
	expectedSQLGeoPackage := "(\"prop9\" = :cql_bcde AND \"prop10\" = :cql_fghi)"
	expectedSQLPostgres := "(\"prop9\" = @cql_bcde AND \"prop10\" = @cql_fghi)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2026-02-13", "cql_fghi": "2026-02-28"}, actual.Params)
		})
	}
}

func TestTemporalIntersectsWithDate(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_INTERSECTS(prop5, DATE('2026-02-12'))"
	expectedSQLGeoPackage := "(\"prop5\" <= :cql_bcde AND \"prop5\" >= :cql_bcde)"
	expectedSQLPostgres := "(\"prop5\" <= @cql_bcde AND \"prop5\" >= @cql_bcde)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2026-02-12"}, actual.Params)
		})
	}
}

func TestTemporalIntersectsWithTimestamp(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_INTERSECTS(prop5, TIMESTAMP('2020-01-01T00:00:00Z'))"
	expectedSQLGeoPackage := "(\"prop5\" <= :cql_bcde AND \"prop5\" >= :cql_bcde)"
	expectedSQLPostgres := "(\"prop5\" <= @cql_bcde AND \"prop5\" >= @cql_bcde)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01T00:00:00Z"}, actual.Params)
		})
	}
}

func TestTemporalIntersectsWithIntervalDate(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_INTERSECTS(prop5, INTERVAL('2020-01-01', '2030-01-01'))"
	expectedSQLGeoPackage := "(\"prop5\" <= :cql_fghi AND \"prop5\" >= :cql_bcde)"
	expectedSQLPostgres := "(\"prop5\" <= @cql_fghi AND \"prop5\" >= @cql_bcde)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01", "cql_fghi": "2030-01-01"}, actual.Params)
		})
	}
}

func TestTemporalIntersectsWithIntervalTimestamp(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop6"}}
	inputCQL := "T_INTERSECTS(prop6, INTERVAL('2017-06-10T07:30:00Z', '2017-06-11T10:30:00Z'))"
	expectedSQLGeoPackage := "(\"prop6\" <= :cql_fghi AND \"prop6\" >= :cql_bcde)"
	expectedSQLPostgres := "(\"prop6\" <= @cql_fghi AND \"prop6\" >= @cql_bcde)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2017-06-10T07:30:00Z", "cql_fghi": "2017-06-11T10:30:00Z"}, actual.Params)
		})
	}
}

func TestTemporalDisjointWithDate(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_DISJOINT(prop5, DATE('2026-02-12'))"
	expectedSQLGeoPackage := "(\"prop5\" < :cql_bcde OR \"prop5\" > :cql_bcde)"
	expectedSQLPostgres := "(\"prop5\" < @cql_bcde OR \"prop5\" > @cql_bcde)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2026-02-12"}, actual.Params)
		})
	}
}

func TestTemporalDisjointWithTimestamp(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_DISJOINT(prop5, TIMESTAMP('2020-01-01T00:00:00Z'))"
	expectedSQLGeoPackage := "(\"prop5\" < :cql_bcde OR \"prop5\" > :cql_bcde)"
	expectedSQLPostgres := "(\"prop5\" < @cql_bcde OR \"prop5\" > @cql_bcde)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01T00:00:00Z"}, actual.Params)
		})
	}
}

func TestTemporalDisjointWithIntervalDate(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_DISJOINT(prop5, INTERVAL('2020-01-01', '2030-01-01'))"
	expectedSQLGeoPackage := "(\"prop5\" < :cql_bcde OR \"prop5\" > :cql_fghi)"
	expectedSQLPostgres := "(\"prop5\" < @cql_bcde OR \"prop5\" > @cql_fghi)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01", "cql_fghi": "2030-01-01"}, actual.Params)
		})
	}
}

func TestTemporalIntersectsIntervalToInterval(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop9"}, {Name: "prop10"}}
	inputCQL := "T_INTERSECTS(INTERVAL(prop9, prop10), INTERVAL('2026-01-01', '2026-12-31'))"
	expectedSQLGeoPackage := "(\"prop9\" <= :cql_fghi AND \"prop10\" >= :cql_bcde)"
	expectedSQLPostgres := "(\"prop9\" <= @cql_fghi AND \"prop10\" >= @cql_bcde)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2026-01-01", "cql_fghi": "2026-12-31"}, actual.Params)
		})
	}
}

func TestTemporalIntervalOperators(t *testing.T) {
	tests := []struct {
		name        string
		inputCQL    string
		expectedSQL string
		params      map[string]any
	}{
		{
			name:        "T_CONTAINS",
			inputCQL:    "T_CONTAINS(INTERVAL(prop9, prop10), DATE('2026-02-20'))",
			expectedSQL: "(\"prop9\" < :cql_bcde AND \"prop10\" > :cql_bcde)",
			params:      map[string]any{"cql_bcde": "2026-02-20"},
		},
		{
			name:        "T_DURING",
			inputCQL:    "T_DURING(prop5, INTERVAL('2020-01-01', '2030-01-01'))",
			expectedSQL: "(\"prop5\" > :cql_bcde AND \"prop5\" < :cql_fghi)",
			params:      map[string]any{"cql_bcde": "2020-01-01", "cql_fghi": "2030-01-01"},
		},
		{
			name:        "T_FINISHEDBY",
			inputCQL:    "T_FINISHEDBY(INTERVAL(prop9, prop10), INTERVAL('2026-01-01', '2026-02-28'))",
			expectedSQL: "(\"prop9\" < :cql_bcde AND \"prop10\" = :cql_fghi)",
			params:      map[string]any{"cql_bcde": "2026-01-01", "cql_fghi": "2026-02-28"},
		},
		{
			name:        "T_FINISHES",
			inputCQL:    "T_FINISHES(INTERVAL(prop9, prop10), INTERVAL('2026-01-01', '2026-02-28'))",
			expectedSQL: "(\"prop9\" > :cql_bcde AND \"prop10\" = :cql_fghi)",
			params:      map[string]any{"cql_bcde": "2026-01-01", "cql_fghi": "2026-02-28"},
		},
		{
			name:        "T_MEETS",
			inputCQL:    "T_MEETS(INTERVAL(prop9, prop10), INTERVAL('2026-03-01', '2026-12-31'))",
			expectedSQL: "\"prop10\" = :cql_bcde",
			params:      map[string]any{"cql_bcde": "2026-03-01", "cql_fghi": "2026-12-31"},
		},
		{
			name:        "T_METBY",
			inputCQL:    "T_METBY(INTERVAL(prop9, prop10), INTERVAL('2026-01-01', '2026-02-13'))",
			expectedSQL: "\"prop9\" = :cql_fghi",
			params:      map[string]any{"cql_bcde": "2026-01-01", "cql_fghi": "2026-02-13"},
		},
		{
			name:        "T_OVERLAPS",
			inputCQL:    "T_OVERLAPS(INTERVAL(prop9, prop10), INTERVAL('2026-02-20', '2026-12-31'))",
			expectedSQL: "(\"prop9\" < :cql_bcde AND \"prop10\" > :cql_bcde AND \"prop10\" < :cql_fghi)",
			params:      map[string]any{"cql_bcde": "2026-02-20", "cql_fghi": "2026-12-31"},
		},
		{
			name:        "T_OVERLAPPEDBY",
			inputCQL:    "T_OVERLAPPEDBY(INTERVAL(prop9, prop10), INTERVAL('2026-01-01', '2026-02-20'))",
			expectedSQL: "(\"prop9\" > :cql_bcde AND \"prop9\" < :cql_fghi AND \"prop10\" > :cql_fghi)",
			params:      map[string]any{"cql_bcde": "2026-01-01", "cql_fghi": "2026-02-20"},
		},
		{
			name:        "T_STARTEDBY",
			inputCQL:    "T_STARTEDBY(INTERVAL(prop9, prop10), INTERVAL('2026-02-13', '2026-02-20'))",
			expectedSQL: "(\"prop9\" = :cql_bcde AND \"prop10\" > :cql_fghi)",
			params:      map[string]any{"cql_bcde": "2026-02-13", "cql_fghi": "2026-02-20"},
		},
		{
			name:        "T_STARTS",
			inputCQL:    "T_STARTS(INTERVAL(prop9, prop10), INTERVAL('2026-02-13', '2026-12-31'))",
			expectedSQL: "(\"prop9\" = :cql_bcde AND \"prop10\" < :cql_fghi)",
			params:      map[string]any{"cql_bcde": "2026-02-13", "cql_fghi": "2026-12-31"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			queryables := []domain.Field{{Name: "prop5"}, {Name: "prop9"}, {Name: "prop10"}}

			for _, datasource := range datasources {
				t.Run(datasource, func(t *testing.T) {
					var actual *SQLResult
					var err error
					switch datasource {
					case gpkg:
						// when
						actual, err = ParseToSQL(tt.inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

						// then
						require.NoError(t, err)
						assertValidSQLiteQuery(t, actual)
						assert.Equal(t, tt.expectedSQL, actual.SQL)
					case postgresql:
						expectedSQLPostgres := strings.ReplaceAll(tt.expectedSQL, ":", "@")

						// when
						actual, err = ParseToSQL(tt.inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

						// then
						require.NoError(t, err)
						assert.Equal(t, expectedSQLPostgres, actual.SQL)
					}
					assert.Equal(t, tt.params, actual.Params)
				})
			}
		})
	}
}

func TestTemporalIntervalOperatorsFailOnInstants(t *testing.T) {
	tests := []struct {
		name        string
		inputCQL    string
		expectedSQL string
		params      map[string]any
	}{
		{
			name:     "T_CONTAINS",
			inputCQL: "T_CONTAINS(prop9, DATE('2026-02-20'))",
		},
		{
			name:     "T_DURING",
			inputCQL: "T_DURING(prop5, DATE('2026-02-20'))",
		},
		{
			name:     "T_FINISHEDBY",
			inputCQL: "T_FINISHEDBY(prop5, DATE('2026-02-20'))",
		},
		{
			name:     "T_FINISHES",
			inputCQL: "T_FINISHES(prop5, DATE('2026-02-20'))",
		},
		{
			name:     "T_MEETS",
			inputCQL: "T_MEETS(prop5, DATE('2026-02-20'))",
		},
		{
			name:     "T_METBY",
			inputCQL: "T_METBY(prop5, DATE('2026-02-20'))",
		},
		{
			name:     "T_OVERLAPS",
			inputCQL: "T_OVERLAPS(prop5, DATE('2026-02-20'))",
		},
		{
			name:     "T_OVERLAPPEDBY",
			inputCQL: "T_OVERLAPPEDBY(prop5, DATE('2026-02-20'))",
		},
		{
			name:     "T_STARTEDBY",
			inputCQL: "T_STARTEDBY(prop5, DATE('2026-02-20'))",
		},
		{
			name:     "T_STARTS",
			inputCQL: "T_STARTS(prop5, DATE('2026-02-20'))",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			queryables := []domain.Field{{Name: "prop5"}, {Name: "prop9"}, {Name: "prop10"}}

			for _, datasource := range datasources {
				t.Run(datasource, func(t *testing.T) {
					// when
					var err error
					switch datasource {
					case gpkg:
						_, err = ParseToSQL(tt.inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
					case postgresql:
						_, err = ParseToSQL(tt.inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
					}

					// then
					assert.ErrorContains(t, err, "only allows intervals, not instants (timestamp/date)")
				})
			}
		})
	}
}

// Unbounded intervals (..) are supported, but each operator has
// limitations on which part of the interval may be unbounded
func TestTemporalOperatorsFailOnInvalidUnboundedIntervals(t *testing.T) {
	tests := []struct {
		name        string
		inputCQL    string
		expectedSQL string
		params      map[string]any
	}{
		{
			name:     "T_AFTER",
			inputCQL: "T_AFTER(prop1, INTERVAL('2026-02-13', '..'))",
		},
		{
			name:     "T_BEFORE",
			inputCQL: "T_BEFORE(prop1, INTERVAL('..', '2026-02-13'))",
		},
		{
			name:     "T_EQUALS",
			inputCQL: "T_EQUALS(prop1, INTERVAL('2026-02-13', '..'))",
		},
		{
			name:     "T_DISJOINT",
			inputCQL: "T_DISJOINT(prop1, INTERVAL('..', '..'))",
		},
		{
			name:     "T_INTERSECTS",
			inputCQL: "T_INTERSECTS(prop1, INTERVAL('..', '..'))",
		},
		{
			name:     "T_CONTAINS",
			inputCQL: "T_CONTAINS(prop1, INTERVAL('2026-02-13', '..'))",
		},
		{
			name:     "T_DURING",
			inputCQL: "T_DURING(prop1, INTERVAL('2026-02-13', '..'))",
		},
		{
			name:     "T_FINISHEDBY",
			inputCQL: "T_FINISHEDBY(prop1, INTERVAL('2026-02-13', '..'))",
		},
		{
			name:     "T_FINISHES",
			inputCQL: "T_FINISHES(prop1, INTERVAL('2026-02-13', '..'))",
		},
		{
			name:     "T_MEETS",
			inputCQL: "T_MEETS(prop1, INTERVAL('..', '2026-02-13'))",
		},
		{
			name:     "T_METBY",
			inputCQL: "T_METBY(prop1, INTERVAL('2026-02-13', '..'))",
		},
		{
			name:     "T_OVERLAPS",
			inputCQL: "T_OVERLAPS(prop1, INTERVAL('2026-02-13', '..'))",
		},
		{
			name:     "T_OVERLAPPEDBY",
			inputCQL: "T_OVERLAPPEDBY(prop1, INTERVAL('2026-02-13', '..'))",
		},
		{
			name:     "T_STARTEDBY",
			inputCQL: "T_STARTEDBY(prop1, INTERVAL('..', '2026-02-13'))",
		},
		{
			name:     "T_STARTS",
			inputCQL: "T_STARTS(prop1, INTERVAL('..', '2026-02-13'))",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			queryables := []domain.Field{{Name: "prop1"}}

			for _, datasource := range datasources {
				t.Run(datasource, func(t *testing.T) {
					// when
					var err error
					switch datasource {
					case gpkg:
						_, err = ParseToSQL(tt.inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
					case postgresql:
						_, err = ParseToSQL(tt.inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
					}

					// then
					assert.ErrorContains(t, err, "requires a second parameter, can't be unbounded")
				})
			}
		})
	}
}

func TestTemporalUnboundedIntervalAtBegin(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_INTERSECTS(prop5, INTERVAL('..', '2020-01-01'))"
	expectedSQLGeoPackage := "\"prop5\" <= :cql_bcde"
	expectedSQLPostgres := "\"prop5\" <= @cql_bcde"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01"}, actual.Params)
		})
	}
}

func TestTemporalUnboundedIntervalAtEnd(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop5"}}
	inputCQL := "T_INTERSECTS(prop5, INTERVAL('2020-01-01', '..'))"
	expectedSQLGeoPackage := "\"prop5\" >= :cql_bcde"
	expectedSQLPostgres := "\"prop5\" >= @cql_bcde"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": "2020-01-01"}, actual.Params)
		})
	}
}

func TestFailOnTemporalLiteralAsFirstArgument(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "starts_at"}, {Name: "ends_at"}}
	inputCQL := "T_DISJOINT(INTERVAL('..', '2005-01-10T01:01:01.393216Z'), INTERVAL(starts_at, ends_at))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			}

			// then
			assert.ErrorContains(t, err, "the first interval should reference a property, not be an unbounded interval")
		})
	}
}

func TestTemporalAndBooleanQuery(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop5"}}
	inputCQL := "prop1 = 10 AND T_AFTER(prop5, DATE('2015-01-01'))"
	expectedSQLGeoPackage := "(cast (\"prop1\" as numeric) = :cql_bcde AND \"prop5\" > :cql_fghi)"
	expectedSQLPostgres := "(cast (\"prop1\" as numeric) = @cql_bcde AND \"prop5\" > @cql_fghi)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			var actual *SQLResult
			var err error
			switch datasource {
			case gpkg:
				// when
				actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assertValidSQLiteQuery(t, actual)
				assert.Equal(t, expectedSQLGeoPackage, actual.SQL)
			case postgresql:
				// when
				actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))

				// then
				require.NoError(t, err)
				assert.Equal(t, expectedSQLPostgres, actual.SQL)
			}
			assert.Equal(t, map[string]any{"cql_bcde": int64(10), "cql_fghi": "2015-01-01"}, actual.Params)
		})
	}
}

func TestTemporalOperatorsNotEnabled(t *testing.T) {
	// given
	cqlConfig := cqlConfigAllEnabled
	cqlConfig.EnableTemporalFunctions = false

	queryables := []domain.Field{{Name: "prop1"}, {Name: "prop5"}}
	inputCQL := "prop1 = 10 AND T_AFTER(prop5, DATE('2015-01-01'))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfig))
			}

			// then
			assert.ErrorContains(t, err, "temporal operators are not enabled for this collection")
		})
	}
}

func TestFailOnSpatialFilteringONonGeoCollection(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "geometry"}, {Name: "prop10"}}
	inputCQL := "S_OVERLAPS(geometry, GEOMETRYCOLLECTION(POINT(0.0 0.0),LINESTRING(0.0 0.0, 1.0 1.0)))"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, 0, geospatial.Attributes, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, 0, geospatial.Attributes, cqlConfigAllEnabled))
			}

			// then
			assert.ErrorContains(t, err, "spatial filtering using 'S_OVERLAPS' is not allowed for this collection since "+
				"it does not contain geospatial items (features), only non-geospatial items (attributes)")
		})
	}
}

func TestFailOnNonSupportedCustomFunctions(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop9"}, {Name: "prop10"}}
	inputCQL := "COOL_FUNCTION(prop9)"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			}

			// then
			assert.ErrorContains(t, err, "function COOL_FUNCTION is unsupported")
		})
	}
}

func TestFailOnNonSupportedArrayOperators(t *testing.T) {
	// given
	queryables := []domain.Field{{Name: "prop9"}, {Name: "prop10"}}
	inputCQL := "A_CONTAINS(prop9, ('foo', 'bar')"

	for _, datasource := range datasources {
		t.Run(datasource, func(t *testing.T) {
			// when
			var err error
			switch datasource {
			case gpkg:
				_, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			case postgresql:
				_, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
			}

			// then
			assert.ErrorContains(t, err, "array operators are not supported")
		})
	}
}

// Test CQL examples provided by OGC.
// See https://github.com/opengeospatial/ogcapi-features/tree/64ac2d892b877b711a4570336cb9d42e2afb4ef8/cql2/standard/schema/examples/text
func TestCQLExamplesProvidedByOGC(t *testing.T) {
	const ext = ".txt"

	ogcExamples := path.Join(pwd, "testdata", "ogc")
	testCases, err := os.ReadDir(ogcExamples)
	require.NoError(t, err)

	for _, datasource := range datasources {
		expectedSuffix := "_expected_" + datasource + ext
		expectedErrSuffix := "_expected_error_" + datasource + ext

		var excludeFilter string
		switch datasource {
		case gpkg:
			excludeFilter = postgresql + ext
		case postgresql:
			excludeFilter = gpkg + ext
		}

		for _, testCase := range testCases {
			if testCase.IsDir() ||
				strings.Contains(testCase.Name(), excludeFilter) ||
				strings.Contains(testCase.Name(), expectedSuffix) ||
				strings.Contains(testCase.Name(), expectedErrSuffix) {
				continue
			}

			// actual test starts here!
			t.Run(testCase.Name(), func(t *testing.T) {
				// given
				queryables := []domain.Field{{Name: "*"}, {Name: "geometry", IsPrimaryGeometry: true}}

				example, err := os.ReadFile(path.Join(ogcExamples, testCase.Name()))
				require.NoError(t, err)

				expectedFile := path.Join(ogcExamples, strings.TrimSuffix(testCase.Name(), ext)+expectedSuffix)
				expectedErrFile := path.Join(ogcExamples, strings.TrimSuffix(testCase.Name(), ext)+expectedErrSuffix)

				inputCQL := strings.Map(removeNewlinesAndTabs, strings.TrimSpace(string(example)))
				require.NotEmpty(t, inputCQL)
				log.Printf("Parsing CQL: %s", inputCQL)

				if strings.HasPrefix(testCase.Name(), "SKIP_") {
					t.Skipf("Skipping %s, since this example is not (yet) supported by our CQL implementation", testCase.Name())
				}

				var expectedSQL, expectedErr []byte
				expectedSQL, err = os.ReadFile(expectedFile)
				if os.IsNotExist(err) {
					// no exception file found, assume error is expected
					expectedErr, err = os.ReadFile(expectedErrFile)
					require.NoError(t, err, "file with expected error not found")
				}

				// when
				var actual *SQLResult
				switch datasource {
				case gpkg:
					actual, err = ParseToSQL(inputCQL, NewGeoPackageListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
				case postgresql:
					actual, err = ParseToSQL(inputCQL, NewPostgresListener(&util.MockRandomizer{}, queryables, 0, domain.AxisOrderXY, geospatial.Features, cqlConfigAllEnabled))
				}

				// then
				switch {
				case len(expectedSQL) > 0:
					require.NoError(t, err)
					require.NotNil(t, actual)
					assert.Equal(t, string(expectedSQL), actual.SQL)
					if datasource == gpkg {
						assertValidSQLiteQuery(t, actual)
					}
				case len(expectedErr) > 0:
					require.Error(t, err)
					assert.Contains(t, err.Error(), string(expectedErr))
				default:
					require.Fail(t, "expected either an expected SQL result or an expected error, but neither was found")
				}
			})
		}
	}
}

func removeNewlinesAndTabs(r rune) rune {
	if r == '\n' || r == '\r' || r == '\t' {
		return -1
	}
	return r
}

func assertValidSQLiteQuery(t *testing.T, result *SQLResult) {
	t.Helper()

	geopackage.LoadDriver()

	require.NotNil(t, result)

	dbPath := pwd + "/testdata/cql.gpkg"
	db, err := sqlx.Open(geopackage.SqliteDriverName, dbPath)
	require.NoError(t, err)
	defer db.Close()

	query := "select * from cql where " + result.SQL
	rows, err := db.NamedQuery(query, result.Params) //nolint:sqlclosecheck
	if err != nil {
		require.FailNow(t, "Failed to execute query", err)
	}
	defer rows.Close()

	require.NoError(t, err)
	for rows.Next() {
		_ = rows.Scan()
	}
	require.NoError(t, rows.Err())
}
