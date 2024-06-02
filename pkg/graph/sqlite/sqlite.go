package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	// sqlite blank import
	_ "github.com/mattn/go-sqlite3"
)

const (
	// MemoryDSN is the in-memory data source name.
	MemoryDSN = "sqlite://:memory:"
	// Migrations is the file path glob for migrations.
	Migrations = "schema/*.sql"
	// Scehem is required sqlite scheme
	Scheme = "sqlite"
	// DefaultLabel is memory graph default label.
	DefaultLabel = "SqliteGraph"
)

//go:embed schema/*.sql
var migrationFS embed.FS

type DB struct {
	db     *sql.DB
	ctx    context.Context // background context
	cancel func()          // cancel background context

	// Datasource name.
	DSN string
}

// NewDB returns a new instance of DB associated with the given datasource name.
func NewDB(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("missing DSN")
	}

	dsn, err := parseDSN(dsn)
	if err != nil {
		return nil, err
	}

	if dsn != MemoryDSN {
		if err := os.MkdirAll(filepath.Dir(dsn), 0700); err != nil {
			return nil, err
		}
	}

	s := &DB{
		DSN: dsn,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())

	if s.db, err = sql.Open("sqlite3", dsn); err != nil {
		return nil, err
	}

	if _, err := s.db.Exec(`PRAGMA journal_mode = wal;`); err != nil {
		return nil, fmt.Errorf("enable wal: %w", err)
	}

	if _, err := s.db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return nil, fmt.Errorf("foreign keys pragma: %w", err)
	}

	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

// migrate database with the attached migration files.
func (s *DB) migrate() error {
	names, err := fs.Glob(migrationFS, Migrations)
	if err != nil {
		return err
	}
	sort.Strings(names)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback()

	for _, name := range names {
		buf, err := fs.ReadFile(migrationFS, name)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(buf)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Close closes the database connection.
func (s *DB) Close() error {
	// Cancel background context.
	s.cancel()

	// Close database.
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
