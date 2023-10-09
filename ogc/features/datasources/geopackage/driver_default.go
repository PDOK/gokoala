//go:build !darwin

package geopackage

import (
	// Use the 'mattn' driver as the default sqlite driver. This driver
	// requires the use of cgo, but we already depend on cgo due to
	// the use of 'go-cloud-sqlite-vfs'.
	_ "github.com/mattn/go-sqlite3"
)
