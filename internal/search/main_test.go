package search

import (
	"context"
	"fmt"
	"log"
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
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

func TestSuggest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	ctx := context.Background()

	// given
	dbPort, postgisContainer, err := setupPostgis(ctx, t)
	if err != nil {
		t.Error(err)
	}
	defer terminateContainer(ctx, t, postgisContainer)

	dbConn := fmt.Sprintf("postgres://postgres:postgres@127.0.0.1:%d/%s?sslmode=disable", dbPort.Int(), "test_db")

	err = etl.CreateSearchIndex(dbConn, "search_index")
	assert.NoError(t, err)

	cfg, err := config.NewConfig("internal/etl/testdata/config.yaml")
	assert.NoError(t, err)
	table := config.FeatureTable{Name: "addresses", FID: "fid", Geom: "geom"}
	err = etl.ImportFile(cfg, "search_index", "internal/etl/testdata/addresses-crs84.gpkg", table, 1000, dbConn)
	assert.NoError(t, err)

	// when/then
	e, err := engine.NewEngine("internal/etl/testdata/config.yaml", false, false)
	assert.NoError(t, err)
	searchEndpoint := NewSearch(e, dbConn, "search_index")
	searchEndpoint.Suggest()
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
