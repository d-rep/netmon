//go:build !cgo

package storage

// Warning: datetime format from this driver is "2021-01-02 16:39:17.123456789 +0000 UTC" (and not ISO-8601)
// which can't then be parsed back out using sqlite native functions like strftime, datetime, etc.
// So we have to cast dates before saving them, or they won't be parseable in SQLite (necessary for our summary view).
import (
	_ "modernc.org/sqlite"
)

const driverName = "sqlite"
