package search

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc/features"
	fd "github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/PDOK/gokoala/internal/search/etl"
	etlconfig "github.com/PDOK/gokoala/internal/search/etl/config"
	"github.com/docker/go-connections/nat"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/text/language"
)

const (
	postgresPortEnv = "DB_PORT"

	testSearchIndex  = "search_index"
	etlConfigFile    = "internal/search/testdata/config_etl.yaml"
	searchConfigFile = "internal/search/testdata/config_search.yaml"
)

func init() {
	// change working dir to root
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	ctx := context.Background()

	// given available postgres
	dbPort, postgisContainer, err := setupPostgis(ctx, t)
	if err != nil {
		t.Error(err)
	}
	t.Cleanup(func() {
		terminateContainer(ctx, t, postgisContainer)
	})

	dbConn := fmt.Sprintf("postgres://postgres:postgres@127.0.0.1:%d/%s?sslmode=disable", dbPort.Int(), "test_db")

	// given empty search index
	err = etl.CreateSearchIndex(dbConn, testSearchIndex, fd.WGS84SRIDPostgis, language.Dutch)
	require.NoError(t, err)

	// given imported geopackage
	err = importGpkg("addresses", dbConn) // in CRS84
	require.NoError(t, err)
	err = importGpkg("buildings", dbConn) // in EPSG:4326
	require.NoError(t, err)

	// given available engine
	newEngine, err := engine.NewEngine(searchConfigFile, "", "", false, false)
	require.NoError(t, err)

	// use fixed decimal limit in coordinates and UTC timezone across all tests for
	// stable output between different data sources (postgres, geopackage, etc)
	newEngine.Config.OgcAPI.FeaturesSearch.MaxDecimals = 5
	newEngine.Config.OgcAPI.FeaturesSearch.ForceUTC = true

	datasources := features.CreateDatasources(newEngine.Config.OgcAPI.FeaturesSearch.OgcAPIFeatures, newEngine.RegisterShutdownHook)
	// axisOrderBySRID := features.DetermineAxisOrder(datasources)

	// given search endpoint
	searchEndpoint, err := NewSearch(newEngine, datasources, nil,
		"internal/search/testdata/rewrites.csv",
		"internal/search/testdata/synonyms.csv")
	require.NoError(t, err)

	// run test cases
	type fields struct {
		url string
	}
	type want struct {
		body       string
		statusCode int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "Fail on search with boolean operators",
			fields: fields{
				url: "http://localhost:8080/search?q=!foo&addresses[version]=1",
			},
			want: want{
				body:       "internal/search/testdata/expected-boolean-operators.json",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Fail on search without collection parameter(s)",
			fields: fields{
				url: "http://localhost:8080/search?q=Oudeschild&limit=50",
			},
			want: want{
				body:       "internal/search/testdata/expected-search-no-collection.json",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Fail on search with collection without version (first variant)",
			fields: fields{
				url: "http://localhost:8080/search?q=Oudeschild&addresses",
			},
			want: want{
				body:       "internal/search/testdata/expected-search-no-version-1.json",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Fail on search with collection without version (second variant)",
			fields: fields{
				url: "http://localhost:8080/search?q=Oudeschild&addresses=1",
			},
			want: want{
				body:       "internal/search/testdata/expected-search-no-version-2.json",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Fail on search with collection without version (third variant)",
			fields: fields{
				url: "http://localhost:8080/search?q=Oudeschild&addresses[foo]=1",
			},
			want: want{
				body:       "internal/search/testdata/expected-search-no-version-3.json",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Complex search term with synonyms and rewrites, should not result in error",
			fields: fields{
				url: "http://localhost:8080/search?q=goev straat 1 in Den Haag niet in Friesland&addresses[version]=1&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-complex-search-term.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search matches multiple suggests, the suggest which equals the display name should be the first result",
			fields: fields{
				url: "http://localhost:8080/search?q=Achtertune 1794BL Oosterend&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-display-name-first-result.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search exact match before should be ranked before wildcard match",
			fields: fields{
				url: "http://localhost:8080/search?q=Holland Den Burg&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-exact-match.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Short results should rank above longer results (for example housenr 1 should rank before 1A)",
			fields: fields{
				url: "http://localhost:8080/search?q=Akenbuurt 1&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-short-before-long.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search for house numbers, should rank in logical order",
			fields: fields{
				url: "http://localhost:8080/search?q=Amaliaweg&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-housenumber-ranking-1.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search for house numbers, should rank in logical order - second test",
			fields: fields{
				url: "http://localhost:8080/search?q=Abbewaal&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-housenumber-ranking-2.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search for house numbers, should rank in logical order - third test",
			fields: fields{
				url: "http://localhost:8080/search?q=Amstel Amsterdam&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-housenumber-ranking-3.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search for house numbers, should rank in logical order - fourth test",
			fields: fields{
				url: "http://localhost:8080/search?q=Amstel 4 Amsterdam&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-housenumber-ranking-4.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search short streetname",
			fields: fields{
				url: "http://localhost:8080/search?q=A Ottoland&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-short-streetname.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search synonym with space",
			fields: fields{
				url: "http://localhost:8080/search?q=Spui Den Haag&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-synonym-with-space.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search synonym with space - second test",
			fields: fields{
				url: "http://localhost:8080/search?q=Spui 's-Gravenhage&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-synonym-with-space.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search streetname with dots",
			fields: fields{
				url: "http://localhost:8080/search?q=A.B.C straat&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-streetname-with-dots.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search streetname with number (not housenumber)",
			fields: fields{
				url: "http://localhost:8080/search?q=1944&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-streetname-with-number.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search long street",
			fields: fields{
				url: "http://localhost:8080/search?q=Ir. Mr. Dr. van Waterschoot van der Grachtstraat&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-long-street.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search frisian street - with frisian input",
			fields: fields{
				url: "http://localhost:8080/search?q=Brânbuorren&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-frisian-street.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search frisian street - with dutch input",
			fields: fields{
				url: "http://localhost:8080/search?q=Branbuorren&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-frisian-street.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search building with polygon output",
			fields: fields{
				url: "http://localhost:8080/search?q=Molwerk&buildings[version]=1&buildings[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-polygon.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search in two collections, with matches in both collections",
			fields: fields{
				url: "http://localhost:8080/search?q=Achter&addresses[version]=1&addresses[relevance]=0.8&buildings[version]=1&buildings[relevance]=0.8&limit=50&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-two-collections.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search in one collections (while another collection also has a match but that one shouldn't appear in the results)",
			fields: fields{
				url: "http://localhost:8080/search?q=Achter&buildings[version]=1&buildings[relevance]=0.8&limit=50&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-one-collection.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search and get output in RD",
			fields: fields{
				url: "http://localhost:8080/search?q=Acht&addresses[version]=1&limit=50&f=json&crs=http://www.opengis.net/def/crs/EPSG/0/28992",
			},
			want: want{
				body:       "internal/search/testdata/expected-rd.json",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given mock time
			now = func() time.Time { return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) }
			engine.Now = now

			// given available server
			rr, ts := createMockServer()
			defer ts.Close()

			// when
			handler := searchEndpoint.Search()
			req, err := createRequest(tt.fields.url)
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)

			// then
			assert.Equal(t, tt.want.statusCode, rr.Code)

			log.Printf("============ ACTUAL:\n %s", rr.Body.String())
			expectedBody, err := os.ReadFile(tt.want.body)
			if err != nil {
				require.NoError(t, err)
			}
			assert.JSONEq(t, string(expectedBody), rr.Body.String())
		})
	}
}

func importGpkg(collectionName string, dbConn string) error {
	conf, err := etlconfig.NewConfig(etlConfigFile)
	if err != nil {
		return err
	}
	collection := conf.CollectionByID(collectionName)
	if collection == nil {
		return fmt.Errorf("collection %s not found in config", collectionName)
	}
	collectionVersion := uuid.NewString()
	return etl.ImportFile(*collection, testSearchIndex, collectionVersion, "internal/search/testdata/fake-addresses-crs84.gpkg", 5000, false, dbConn)
}

func setupPostgis(ctx context.Context, t *testing.T) (nat.Port, testcontainers.Container, error) {
	t.Helper()
	req := testcontainers.ContainerRequest{
		Image: "docker.io/imresamu/postgis:16-3.5-bookworm", // use debian, not alpine (proj issues between environments). Also use multi-arch image (AMD64 and ARM).
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		},
		ExposedPorts: []string{"5432/tcp"},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
		WaitingFor:   wait.ForLog("PostgreSQL init process complete; ready for start up."),
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      "internal/search/etl/testdata/init-db.sql",
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base("/testdata/init-db.sql"),
				FileMode:          0755,
			},
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Error(err)
	}
	if port.Int() == 0 {
		t.Error("port is 0")
	}
	if err = os.Setenv(postgresPortEnv, port.Port()); err != nil {
		log.Fatal("failed to set env var", err)
	}

	log.Println("Giving postgres a few extra seconds to fully start")
	time.Sleep(2 * time.Second)
	log.Printf("Postgres running at port %s", port.Port())

	return port, container, err
}

func terminateContainer(ctx context.Context, t *testing.T, container testcontainers.Container) {
	t.Helper()
	if err := container.Terminate(ctx); err != nil {
		t.Fatalf("Failed to terminate container: %s", err.Error())
	}
	if err := os.Unsetenv(postgresPortEnv); err != nil {
		log.Fatal("failed to unset env var", err)
	}
}

func createMockServer() (*httptest.ResponseRecorder, *httptest.Server) {
	rr := httptest.NewRecorder()
	l, err := net.Listen("tcp", "localhost:") // random port
	if err != nil {
		log.Fatal("failed to start mocks erver", err)
	}
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engine.SafeWrite(w.Write, []byte(r.URL.String()))
	}))
	if err = ts.Listener.Close(); err != nil {
		log.Fatal("failed to close mocks erver", err)
	}
	ts.Listener = l
	ts.Start()

	return rr, ts
}

func createRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if req == nil || err != nil {
		return req, err
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
	return req, err
}
