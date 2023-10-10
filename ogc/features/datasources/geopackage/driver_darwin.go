//go:build darwin

package geopackage

import (
	"database/sql"

	// Don't use the 'mattn' sqlite driver on macOS. Since the LLVM linker on macOS
	// doesn't support the '--allow-multiple-definition' flag. This flag is required
	// since both the 'mattn' driver and 'go-cloud-sqlite-vfs' contain a copy of
	// the sqlite C-code, which causes duplicate symbols (aka multiple definitions).
	//
	// As an alternative we use the following pure Go sqlite driver. But we favor the
	// actual cgo driver for production use on Linux (in a Docker container).
	"modernc.org/sqlite"
)

// register 'modernc' sqlite driver under same name as 'mattn' sqlite driver
func init() {
	sql.Register(sqliteDriverName, &sqlite.Driver{})
}
