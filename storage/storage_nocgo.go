//go:build !cgo

package storage

import (
	_ "modernc.org/sqlite"
)

const driverName = "sqlite"
