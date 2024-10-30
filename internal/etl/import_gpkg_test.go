package etl

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"testing"

	"github.com/PDOK/gomagpie/config"
	"github.com/stretchr/testify/assert"
)

var pwd string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	pwd = path.Dir(filename)
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
	err = ImportGeoPackage(cfg, pwd+"/testdata/addresses-crs84.gpkg", "addresses", 1000, "", "", dbConn)
	assert.NoError(t, err)
}
