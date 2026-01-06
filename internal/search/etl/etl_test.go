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

	"github.com/PDOK/gokoala/internal/search/etl/config"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/text/language"
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
	dbConn := makeDbConnection(dbPort)

	// when/then
	err = CreateSearchIndex(dbConn, "search_index", 28992, language.Dutch)
	require.NoError(t, err)
	err = insertTestData(ctx, dbConn)
	require.NoError(t, err)
}

func TestCreateSearchIndexIdempotent(t *testing.T) {
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
	dbConn := makeDbConnection(dbPort)

	// when/then
	err = CreateSearchIndex(dbConn, "search_index", 28992, language.English)
	require.NoError(t, err)
	err = CreateSearchIndex(dbConn, "search_index", 28992, language.English) // second time, should not fail
	require.NoError(t, err)
}

func makeDbConnection(dbPort nat.Port) string {
	return fmt.Sprintf("postgres://postgres:postgres@127.0.0.1:%d/%s?sslmode=disable", dbPort.Int(), "test_db")
}

func TestGetVersion(t *testing.T) {

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
	dbConn := makeDbConnection(dbPort)

	// when/then
	err = CreateSearchIndex(dbConn, "search_index", 28992, language.English)
	require.NoError(t, err)

	cfg, err := config.NewConfig(pwd + "/testdata/config.yaml")
	if err != nil {
		t.Error(err)
	}
	require.NotNil(t, cfg)
	collection := cfg.CollectionByID("addresses")
	require.NotNil(t, collection)
	collectionVersion := uuid.NewString()
	err = ImportFile(*collection, "search_index", collectionVersion,
		pwd+"/testdata/addresses-crs84.gpkg", 1000, true, dbConn)
	require.NoError(t, err)

	version, err := GetRevision(dbConn, "addresses", "search_index")
	require.NoError(t, err)
	assert.NotEmpty(t, version)
	assert.Equal(t, collectionVersion, version)
}

func TestImportGeoPackage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	tests := []struct {
		name  string
		where string
		count int
	}{
		{
			name:  "import everything",
			where: "",
			count: 66060, // 33030*2
		},
		{
			name:  "with where clause",
			where: "fid <= 2",
			count: 2 * 2, // * 2 because 2 suggest templates
		},
	}
	for _, tt := range tests {
		ctx := context.Background()

		// given
		dbPort, postgisContainer, err := setupPostgis(ctx, t)
		if err != nil {
			t.Error(err)
		}
		dbConn := makeDbConnection(dbPort)

		cfg, err := config.NewConfig(pwd + "/testdata/config.yaml")
		if err != nil {
			t.Error(err)
		}
		require.NotNil(t, cfg)
		for i := range cfg.Collections {
			cfg.Collections[i].Filter = tt.where
		}
		collection := cfg.CollectionByID("addresses")
		require.NotNil(t, collection)

		// when/then
		err = CreateSearchIndex(dbConn, "search_index", 4326, language.English)
		require.NoError(t, err)

		collectionVersion := uuid.NewString()
		err = ImportFile(*collection, "search_index", collectionVersion, pwd+"/testdata/addresses-crs84.gpkg",
			1000, true, dbConn)
		require.NoError(t, err)

		// check nr of records
		db, err := pgx.Connect(ctx, dbConn)
		require.NoError(t, err)
		var count int
		err = db.QueryRow(ctx, "select count(*) from search_index").Scan(&count)
		db.Close(ctx)
		require.NoError(t, err)
		assert.Equal(t, tt.count, count)

		terminateContainer(ctx, t, postgisContainer)
	}
}

func TestImportGeoPackageMultipleTimes(t *testing.T) {
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
	dbConn := makeDbConnection(dbPort)

	cfg, err := config.NewConfig(pwd + "/testdata/config.yaml")
	if err != nil {
		t.Error(err)
	}
	require.NotNil(t, cfg)
	collection := cfg.CollectionByID("addresses")
	require.NotNil(t, collection)

	// when: create index
	err = CreateSearchIndex(dbConn, "search_index", 4326, language.English)
	require.NoError(t, err)

	// when: first import (should create new table)
	collectionVersion := uuid.NewString()
	err = ImportFile(*collection, "search_index", collectionVersion, pwd+"/testdata/addresses-crs84.gpkg",
		1000, true, dbConn)
	require.NoError(t, err)

	db, err := pgx.Connect(ctx, dbConn)
	require.NoError(t, err)
	defer db.Close(ctx)
	var tableCount, indexCountFirst, indexCountSecond int

	// then: check table is filled
	err = db.QueryRow(ctx, "select count(*) from search_index_addresses_alpha").Scan(&tableCount)
	require.NoError(t, err)
	assert.NotZero(t, tableCount)

	// then: check index is filled
	err = db.QueryRow(ctx, "select count(*) from search_index").Scan(&indexCountFirst)
	require.NoError(t, err)
	assert.NotZero(t, indexCountFirst)

	// then: check metadata table is updated
	version, err := GetRevision(dbConn, "addresses", "search_index")
	require.NoError(t, err)
	assert.NotEmpty(t, version)
	assert.Equal(t, collectionVersion, version)

	// when: second import (should create a new table and switch partitions)
	collectionVersion = uuid.NewString()
	err = ImportFile(*collection, "search_index", collectionVersion, pwd+"/testdata/addresses-crs84.gpkg",
		1000, true, dbConn)
	require.NoError(t, err)

	// then: check table is filled
	err = db.QueryRow(ctx, "select count(*) from search_index_addresses_beta").Scan(&tableCount)
	require.NoError(t, err)
	assert.NotZero(t, tableCount)

	// then: check index is filled, should be the same as before
	err = db.QueryRow(ctx, "select count(*) from search_index").Scan(&indexCountSecond)
	require.NoError(t, err)
	assert.NotZero(t, indexCountSecond)
	assert.Equal(t, indexCountFirst, indexCountSecond)

	// then: check metadata table is updated
	version, err = GetRevision(dbConn, "addresses", "search_index")
	require.NoError(t, err)
	assert.NotEmpty(t, version)
	assert.Equal(t, collectionVersion, version)

	// when: third import (should fill an existing table and switch partitions)
	collectionVersion = uuid.NewString()
	err = ImportFile(*collection, "search_index", collectionVersion, pwd+"/testdata/addresses-crs84.gpkg",
		1000, true, dbConn)
	require.NoError(t, err)

	// then: check table is filled
	err = db.QueryRow(ctx, "select count(*) from search_index_addresses_alpha").Scan(&tableCount)
	require.NoError(t, err)
	assert.NotZero(t, tableCount)

	// then: check metadata table is updated
	version, err = GetRevision(dbConn, "addresses", "search_index")
	require.NoError(t, err)
	assert.NotEmpty(t, version)
	assert.Equal(t, collectionVersion, version)
}

func TestImportGeoPackageNoDuplicates(t *testing.T) {
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
	dbConn := makeDbConnection(dbPort)

	cfg, err := config.NewConfig(pwd + "/testdata/config.yaml")
	if err != nil {
		t.Error(err)
	}
	require.NotNil(t, cfg)
	collection := cfg.CollectionByID("addresses")
	require.NotNil(t, collection)

	// when: create index
	err = CreateSearchIndex(dbConn, "search_index", 4326, language.English)
	require.NoError(t, err)

	// when: first import (should create new table)
	collectionVersion := uuid.NewString()
	err = ImportFile(*collection, "search_index", collectionVersion, pwd+"/testdata/addresses-crs84.gpkg",
		1000, true, dbConn)
	require.NoError(t, err)

	// when: second import (should create a new table and switch partitions)
	collectionVersion = uuid.NewString()
	err = ImportFile(*collection, "search_index", collectionVersion, pwd+"/testdata/addresses-crs84.gpkg",
		1000, true, dbConn)
	require.NoError(t, err)

	db, err := pgx.Connect(ctx, dbConn)
	require.NoError(t, err)
	defer db.Close(ctx)

	// when: attach first partition again
	_, err = db.Exec(ctx, "alter table search_index attach partition search_index_addresses_alpha for values in ('addresses')")

	// then: should fail
	require.ErrorContains(t, err, "ERROR: partition \"search_index_addresses_alpha\" would overlap partition \"search_index_addresses_beta\"")
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
				HostFilePath:      "./testdata/init-db.sql",
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base("/testdata/init-db.sql"),
				FileMode:          0755,
			},
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil || container == nil {
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
	t.Helper()
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

	var partitions = [3]string{"addresses", "roads"}

	for i := range partitions {
		partition := `create table if not exists search_index_` + partitions[i] + ` partition of search_index for values in ('` + partitions[i] + `');`
		_, err = db.Exec(ctx, partition)
		if err != nil {
			log.Printf("Error creating partition: %s Error: %v\n", partitions[i], err)
		}
	}

	testData := `
	insert into search_index(feature_id, collection_id, collection_version, display_name, suggest, geometry_type, bbox)
	values
	  ('408f5e13', 'addresses', 1, 'Daendelsweg 4A, 7315AJ Apeldoorn', 'Daendelsweg 4A, 7315AJ Apeldoorn', 'POINT'     , 'POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))'),
	  ('408f5e13', 'addresses', 1, 'Daendelsweg 4A, 7315AJ Apeldoorn', 'Daendelsweg 4A, 7315AJ Apeldoorn', 'POINT'     , 'POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))'),
	  ('408f5e13', 'addresses', 1, 'Daendelsweg 4A, 7315AJ Apeldoorn', 'Daendelsweg 4A, 7315AJ Apeldoorn', 'POINT'     , 'POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))'),
	  ('408f5e13', 'addresses', 1, 'Daendelsweg 4A, 7315AJ Apeldoorn', 'Daendelsweg 4A, 7315AJ Apeldoorn', 'POINT'     , 'POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))'),
	  ('1e99b620', 'roads'  , 2, 'Daendelsweg, 7315AJ Apeldoorn'   , 'Daendelsweg 4A, 7315AJ Apeldoorn', 'LINESTRING', 'POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))'),
	  ('1e99b620', 'roads'  , 2, 'Daendelsweg, 7315AJ Apeldoorn'   , 'Daendelsweg 4A, 7315AJ Apeldoorn', 'LINESTRING', 'POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))');
    `

	_, err = db.Exec(ctx, testData)
	if err != nil {
		log.Printf("Error creating testData: %v\n", err)
	}
	return err
}
