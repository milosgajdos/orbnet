package sqlite

import (
	"flag"
	"path/filepath"
	"testing"
)

var dump = flag.Bool("dump", false, "save work data")

// MustOpenDB returns a new, open DB. Fatal on error.
func MustOpenDB(tb testing.TB) *DB {
	tb.Helper()

	// Write to an in-memory database by default.
	// If the -dump flag is set, generate a temp file for the database.
	dsn := MemoryDSN
	if *dump {
		dir := tb.TempDir()
		dsn = filepath.Join(dir, "db")
		println("DUMP=" + dsn)
	}

	db, err := NewDB(dsn)
	if err != nil {
		tb.Fatal(err)
	}
	return db
}

// MustCloseDB closes the DB. Fatal on error.
func MustCloseDB(tb testing.TB, db *DB) {
	tb.Helper()
	if err := db.Close(); err != nil {
		tb.Fatal(err)
	}
}
