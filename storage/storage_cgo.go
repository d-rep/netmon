//go:build cgo

package storage

import (
	_ "github.com/mattn/go-sqlite3"
)

const driverName = "sqlite3"
