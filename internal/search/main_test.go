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
)

const testSearchIndex = "search_index"

func init() {
	// change working dir to root
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestSuggest(t *testing.T) {
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
	eng, err := engine.NewEngine("internal/etl/testdata/config.yaml", false, false)
	assert.NoError(t, err)

	// given search endpoint
	searchEndpoint := NewSearch(eng, dbConn, testSearchIndex)

	// given empty search index
	err = etl.CreateSearchIndex(dbConn, testSearchIndex)
	assert.NoError(t, err)

	// given imported geopackage
	conf, err := config.NewConfig("internal/etl/testdata/config.yaml")
	assert.NoError(t, err)
	collection := config.CollectionByID(conf, "addresses")
	table := config.FeatureTable{Name: "addresses", FID: "fid", Geom: "geom"}
	err = etl.ImportFile(*collection, testSearchIndex, "internal/etl/testdata/addresses-crs84.gpkg", table, 1000, dbConn)
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
			name: "Suggest: Oudeschild",
			fields: fields{
				url: "http://localhost:8080/search/suggest?q=\"Oudeschild\"&limit=50",
			},
			want: want{
				body:       "internal/search/testdata/expected-search-oudeschild.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Suggest: Den ",
			fields: fields{
				url: "http://localhost:8080/search/suggest?q=\"Den\"&limit=50",
			},
			want: want{
				body:       "internal/search/testdata/expected-search-den.json",
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Suggest: Den. With deepCopy params",
			fields: fields{
				url: "http://localhost:8080/search/suggest?q=\"Den\"&weg[version]=2&weg[relevance]=0.8&adres[version]=1&adres[relevance]=1&limit=10&f=json&crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992",
			},
			want: want{
				body:       "internal/search/testdata/expected-search-den-deepcopy.json",
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock time
			now = func() time.Time { return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) }

			// given available server
			rr, ts := createMockServer()
			defer ts.Close()

			// when
			handler := searchEndpoint.Suggest()
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

func setupPostgis(ctx context.Context, t *testing.T) (nat.Port, testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image: "docker.io/postgis/postgis:16-3.5-alpine",
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
