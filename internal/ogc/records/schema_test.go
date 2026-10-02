package records

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenRecordsDatabaseGeoPackagePaths(t *testing.T) {
	t.Chdir(t.TempDir())
	file := "records #?.gpkg"
	absoluteFile, err := filepath.Abs(file)
	require.NoError(t, err)
	fileURL := url.URL{Scheme: "file", Path: filepath.ToSlash(absoluteFile)}
	setupDB, err := sql.Open("sqlite3", fileURL.String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = setupDB.Close() })
	_, err = setupDB.Exec("CREATE TABLE records (id TEXT)")
	require.NoError(t, err)
	require.NoError(t, setupDB.Close())

	for _, path := range []string{file, "./" + file, absoluteFile} {
		t.Run(path, func(t *testing.T) {
			db, driver, err := openRecordsDatabase(&config.Datasource{GeoPackage: &config.GeoPackage{
				Local: &config.GeoPackageLocal{File: path},
			}})
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			require.Equal(t, "sqlite", driver)
			var count int
			require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM records").Scan(&count))
			require.Zero(t, count)
			_, err = db.Exec("INSERT INTO records (id) VALUES ('test')")
			require.ErrorContains(t, err, "readonly")
		})
	}
}

func TestGeoPackageSchemaExecutes(t *testing.T) {
	file := filepath.Join(t.TempDir(), "records.gpkg")
	setupDB, err := sql.Open("sqlite3", file)
	require.NoError(t, err)
	core, err := os.ReadFile("schema/geopackage-core.sql")
	require.NoError(t, err)
	_, err = setupDB.Exec(string(core))
	require.NoError(t, err)
	schema, err := os.ReadFile("schema/geopackage.sql")
	require.NoError(t, err)
	_, err = setupDB.Exec(string(schema))
	require.NoError(t, err)
	seed, err := os.ReadFile("schema/example-data.sql")
	require.NoError(t, err)
	_, err = setupDB.Exec(string(seed))
	require.NoError(t, err)
	require.NoError(t, setupDB.Close())
	db, driver, err := openRecordsDatabase(&config.Datasource{GeoPackage: &config.GeoPackage{
		Local: &config.GeoPackageLocal{File: file},
	}})
	require.NoError(t, err)
	require.Equal(t, "sqlite", driver)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	var datasetTableType string
	require.NoError(t, db.QueryRow("SELECT data_type FROM gpkg_contents WHERE table_name = 'dcat_dataset'").Scan(&datasetTableType))
	require.Equal(t, "attributes", datasetTableType)
	enableCQL := true
	serverConfig := &config.Config{
		BaseURL: config.URL{URL: &url.URL{Scheme: "http", Host: "example.org"}},
		OgcAPI: config.OgcAPI{Records: &config.OgcAPIRecords{
			Datasource: &config.Datasource{GeoPackage: &config.GeoPackage{Local: &config.GeoPackageLocal{File: file}}},
			Collections: config.RecordsCollections{{
				ID: "geoparquet-example",
				Filters: config.FeatureFilters{
					Properties: []config.Queryable{{Name: "title"}},
					CQL:        config.CQL{Enable: &enableCQL},
				},
			},
			}},
		},
	}
	recordsAPI, err := NewRecords(&engine.Engine{Config: serverConfig})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, recordsAPI.Close()) })
	require.True(t, recordsAPI.HasCollection("geoparquet-example"))
	require.Len(t, serverConfig.OgcAPI.Records.Collections, 1)
	require.Len(t, serverConfig.OgcAPI.Records.Collections[0].Records, 1)
	record := serverConfig.OgcAPI.Records.Collections[0].Records[0]
	assert.Equal(t, "geoparquet-example-dataset", record.ID)
	title, err := record.Properties["title"].Any()
	require.NoError(t, err)
	assert.Equal(t, "GeoParquet example dataset", title)
	assert.Equal(t, []string{"https://stac-extensions.github.io/table/v1.3.0/schema.json"}, record.STACExtensions)
	assert.Equal(t, "https://raw.githubusercontent.com/opengeospatial/geoparquet/main/examples/example.parquet", record.Assets["geoparquet"].Href)
	matchedIDs, err := cqlMatchedRecordIDs(t.Context(), recordsAPI.database, recordsAPI.databaseDriver,
		serverConfig.OgcAPI.Records.Collections[0], "title = 'GeoParquet example dataset'")
	require.NoError(t, err)
	assert.Equal(t, map[string]bool{"geoparquet-example-dataset": true}, matchedIDs)
	router := chi.NewRouter()
	router.Get(pluralItemsPath, recordsAPI.Items())
	request := httptest.NewRequest(http.MethodGet,
		"http://example.org/collections/geoparquet-example/items?filter=title%20%3D%20%27GeoParquet%20example%20dataset%27&filter-lang=cql2-text", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	features := response["features"].([]any)
	require.Len(t, features, 1)
	assert.Equal(t, "geoparquet-example-dataset", features[0].(map[string]any)["id"])

	queryablesRouter := chi.NewRouter()
	queryablesRouter.Get(queryablesPath, recordsAPI.Queryables())
	queryablesRecorder := httptest.NewRecorder()
	queryablesRequest := httptest.NewRequest(http.MethodGet, "http://example.org/collections/geoparquet-example/queryables", nil)
	queryablesRouter.ServeHTTP(queryablesRecorder, queryablesRequest)
	require.Equal(t, http.StatusOK, queryablesRecorder.Code)
	var queryablesResponse map[string]any
	require.NoError(t, json.Unmarshal(queryablesRecorder.Body.Bytes(), &queryablesResponse))
	queryableProperties := queryablesResponse["properties"].(map[string]any)
	assert.Contains(t, queryableProperties, "id")
	assert.Contains(t, queryableProperties, "title")
}
