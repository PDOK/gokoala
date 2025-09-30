//go:build darwin

package geopackage

import (
	"log"

	"github.com/PDOK/gokoala/config"
)

// Dummy implementation to make compilation on macOS work. We don't support cloud-backed
// sqlite/geopackages on macOS since the LLVM linker on macOS doesn't support the
// '--allow-multiple-definition' flag. This flag is required since both the 'mattn' sqlite
// driver and 'go-cloud-sqlite-vfs' contain a copy of the sqlite C-code, which causes
// duplicate symbols (aka multiple definitions).
func newCloudBackedGeoPackage(_ *config.GeoPackageCloud) geoPackageBackend {
	log.Fatalf("Cloud backed GeoPackage isn't supported on darwin/macos")

	return nil
}
