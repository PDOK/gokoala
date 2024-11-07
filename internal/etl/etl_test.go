package etl

import (
	"context"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/PDOK/gomagpie/config"
	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var pwd string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	pwd = path.Dir(filename)
}

func TestCreateSearchIndex(t *testing.T) {
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

	// when/then
	err = CreateSearchIndex(dbConn)
	assert.NoError(t, err)
	err = insertTestData(ctx, dbConn)
	assert.NoError(t, err)
}

func TestImportGeoPackage(t *testing.T) {
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

	cfg, err := config.NewConfig(pwd + "/testdata/config.yaml")
	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, cfg)

	// when/then
	err = CreateSearchIndex(dbConn)
	assert.NoError(t, err)
	err = ImportFile(cfg, pwd+"/testdata/addresses-crs84.gpkg", config.FeatureTable{Name: "addresses", FID: "fid", Geom: "geom"}, 1000, "", "", dbConn)
	assert.NoError(t, err)
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
				HostFilePath:      "../../tests/testdata/sql/init-db.sql",
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base("../../testdata/init-db.sql"),
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

func insertTestData(ctx context.Context, conn string) error {
	db, err := pgx.Connect(ctx, conn)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer db.Close(ctx)

	// Create required partitions for testData
	//nolint:misspell
	partitions := `
	create table search_index_addres partition of search_index
		for values in ('adres');
		-- partition by list(collection_version);
	create table search_index_weg partition of search_index
		for values in ('weg');
		-- partition by list(collection_version);
    `

	_, err = db.Exec(ctx, partitions)
	if err != nil {
		log.Printf("Error creating partitions: %v\n", err)
	}

	testData := `
	insert into search_index(feature_id, collection_id, collection_version, display_name, suggest, geometry_type, bbox)
	values
	  ('408f5e13', 'adres', 1, 'Daendelsweg 4A, 7315AJ Apeldoorn', 'Daendelsweg 4A, 7315AJ Apeldoorn', 'POINT'     , 'POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))'),
	  ('408f5e13', 'adres', 1, 'Daendelsweg 4A, 7315AJ Apeldoorn', 'Daendelsweg 4A, 7315AJ Apeldoorn', 'POINT'     , 'POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))'),
	  ('408f5e13', 'adres', 1, 'Daendelsweg 4A, 7315AJ Apeldoorn', 'Daendelsweg 4A, 7315AJ Apeldoorn', 'POINT'     , 'POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))'),
	  ('408f5e13', 'adres', 1, 'Daendelsweg 4A, 7315AJ Apeldoorn', 'Daendelsweg 4A, 7315AJ Apeldoorn', 'POINT'     , 'POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))'),
	  ('1e99b620', 'weg'  , 2, 'Daendelsweg, 7315AJ Apeldoorn'   , 'Daendelsweg 4A, 7315AJ Apeldoorn', 'LINESTRING', 'POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))'),
	  ('1e99b620', 'weg'  , 2, 'Daendelsweg, 7315AJ Apeldoorn'   , 'Daendelsweg 4A, 7315AJ Apeldoorn', 'LINESTRING', 'POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))');
    `

	_, err = db.Exec(ctx, testData)
	if err != nil {
		log.Printf("Error creating testData: %v\n", err)
	}
	return err
}
