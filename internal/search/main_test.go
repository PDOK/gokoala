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

	"github.com/PDOK/gomagpie/config"
	"github.com/PDOK/gomagpie/internal/engine"
	"github.com/PDOK/gomagpie/internal/etl"
	"github.com/docker/go-connections/nat"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/text/language"
)

const testSearchIndex = "search_index"
const configFile = "internal/search/testdata/config.yaml"

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
	defer terminateContainer(ctx, t, postgisContainer)

	dbConn := fmt.Sprintf("postgres://postgres:postgres@127.0.0.1:%d/%s?sslmode=disable", dbPort.Int(), "test_db")

	// given available engine
	eng, err := engine.NewEngine(configFile, false, false)
	assert.NoError(t, err)

	// given search endpoint
	searchEndpoint, err := NewSearch(eng, dbConn, testSearchIndex, "internal/search/testdata/rewrites.csv", "internal/search/testdata/synonyms.csv")
	assert.NoError(t, err)

	// given empty search index
	err = etl.CreateSearchIndex(dbConn, testSearchIndex, language.Dutch)
	assert.NoError(t, err)

	// given imported geopackage (creates two collections in search_index with identical data)
	err = importAddressesGpkg("addresses", dbConn)
	assert.NoError(t, err)
	err = importAddressesGpkg("buildings", dbConn)
	assert.NoError(t, err)

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
			name: "Search: 'Den' for a single collection in WGS84 (default)",
			fields: fields{
				url: "http://localhost:8080/search?q=Den&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-search-den-single-collection-wgs84.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search: 'Den' for a single collection in RD",
			fields: fields{
				url: "http://localhost:8080/search?q=Den&addresses[version]=1&addresses[relevance]=0.8&limit=10&f=json&crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992",
			},
			want: want{
				body:       "internal/search/testdata/expected-search-den-single-collection-rd.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search: 'Den' in another collection in WGS84",
			fields: fields{
				url: "http://localhost:8080/search?q=Den&buildings[version]=1&limit=10&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-search-den-building-collection-wgs84.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search: 'Den' in multiple collections: with one non-existing collection, so same output as single collection) in RD",
			fields: fields{
				url: "http://localhost:8080/search?q=Den&addresses[version]=1&addresses[relevance]=0.8&foo[version]=2&foo[relevance]=0.8&limit=10&f=json&crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992",
			},
			want: want{
				body:       "internal/search/testdata/expected-search-den-single-collection-rd.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Search: 'Den' in multiple collections: collection addresses + collection buildings, but addresses with non-existing version",
			fields: fields{
				url: "http://localhost:8080/search?q=Den&addresses[version]=2&buildings[version]=1&limit=20&f=json",
			},
			want: want{
				body:       "internal/search/testdata/expected-search-den-multiple-collection-single-output-wgs84.json", // only expect building results since addresses version doesn't exist.
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
			assert.NoError(t, err)
			handler.ServeHTTP(rr, req)

			// then
			assert.Equal(t, tt.want.statusCode, rr.Code)

			log.Printf("============ ACTUAL:\n %s", rr.Body.String())
			expectedBody, err := os.ReadFile(tt.want.body)
			if err != nil {
				assert.NoError(t, err)
			}
			assert.JSONEq(t, string(expectedBody), rr.Body.String())
		})
	}
}

func importAddressesGpkg(collectionName string, dbConn string) error {
	conf, err := config.NewConfig(configFile)
	if err != nil {
		return err
	}
	collection := config.CollectionByID(conf, collectionName)
	table := config.FeatureTable{Name: "addresses", FID: "fid", Geom: "geom"}
	return etl.ImportFile(*collection, testSearchIndex,
		"internal/etl/testdata/addresses-crs84.gpkg", table, 5000, dbConn)
}

func setupPostgis(ctx context.Context, t *testing.T) (nat.Port, testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image: "docker.io/postgis/postgis:16-3.5", // use debian, not alpine (proj issues between environments)
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
				HostFilePath:      "tests/testdata/sql/init-db.sql",
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base("testdata/init-db.sql"),
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

	log.Println("Giving postgres a few extra seconds to fully start")
	time.Sleep(2 * time.Second)

	return port, container, err
}

func terminateContainer(ctx context.Context, t *testing.T, container testcontainers.Container) {
	if err := container.Terminate(ctx); err != nil {
		t.Fatalf("Failed to terminate container: %s", err.Error())
	}
}

func createMockServer() (*httptest.ResponseRecorder, *httptest.Server) {
	rr := httptest.NewRecorder()
	l, err := net.Listen("tcp", "localhost:9095")
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engine.SafeWrite(w.Write, []byte(r.URL.String()))
	}))
	err = ts.Listener.Close()
	if err != nil {
		log.Fatal(err)
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
