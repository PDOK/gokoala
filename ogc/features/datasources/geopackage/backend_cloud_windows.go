//go:build windows

package geopackage

import (
	"log"

	"github.com/PDOK/gokoala/engine"
)

func newCloudBackedGeoPackage(_ *engine.GeoPackageCloud) geoPackageBackend {
	log.Fatalf("Cloud backed GeoPackage isn't supported on windows")
	return nil
}
