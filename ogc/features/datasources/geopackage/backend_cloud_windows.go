//go:build windows

package geopackage

import (
	"log"

	"github.com/PDOK/gokoala/config"
)

func newCloudBackedGeoPackage(_ *config.GeoPackageCloud) geoPackageBackend {
	log.Fatalf("Cloud backed GeoPackage isn't supported on windows")
	return nil
}
