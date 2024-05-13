//go:build windows

package geopackage

import (
	"log"

	"github.com/PDOK/gokoala/config"
)

// Dummy implementation to make compilation on window work.
func newCloudBackedGeoPackage(_ *config.GeoPackageCloud) geoPackageBackend {
	log.Fatalf("Cloud backed GeoPackage isn't supported on windows")
	return nil
}
