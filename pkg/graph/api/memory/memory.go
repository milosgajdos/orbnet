package memory

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"

	"github.com/milosgajdos/orbnet/pkg/graph/api/memory/marshal/json"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"
)

var (
	// ErrDBClosed is returned if attempting a transaction on closed DB.
	ErrDBClosed = errors.New("ErrDBClosed")
)

const (
	// DefaultLabel is the default graph label.
	DefaultLabel = "MemoryGraph"
	// MemoryDSN is the in-memory data source name.
	MemoryDSN = ":memory:"
)

// DB is an in-memory graph store.
type DB struct {
	// Datasource name.
	DSN string
	// Closed flag for DB operations.
	Closed bool
	// db stores memory db
	db map[string]*memory.Graph
	*sync.RWMutex
}

// Tx wraps DB to provide graph CRUD operations.
type Tx struct {
	db *DB
}

// NewDB returns a new instance of DB associated with the given datasource name.
// NOTE: dsn is either ":memory:" or a path to a directory that stores graphs.
func NewDB(dsn string) (*DB, error) {
	return &DB{
		DSN:     dsn,
		Closed:  true,
		RWMutex: &sync.RWMutex{},
	}, nil
}

// Open opens the database connection.
func (db *DB) Open() (err error) {
	db.Lock()
	defer db.Unlock()

	if !db.Closed {
		return nil
	}

	// Ensure a DSN is set before attempting to open the database.
	if db.DSN == "" {
		return fmt.Errorf("dsn required")
	}

	if db.DSN != MemoryDSN {
		if db.db, err = openFromFS(os.DirFS(db.DSN)); err != nil {
			return err
		}
		db.Closed = false
		return nil
	}

	db.db = make(map[string]*memory.Graph)
	db.Closed = false

	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	db.RLock()
	defer db.RUnlock()

	if db.Closed {
		return nil
	}

	db.Closed = true
	return nil
}

// Do creates a DB operation.
// nolint:revive
func (db *DB) BeginTx(ctx context.Context) (*Tx, error) {
	if db.Closed {
		return nil, ErrDBClosed
	}

	return &Tx{
		db: db,
	}, nil
}

// openFromFS opens DB and loads all data stored on the given fs.
//
// TODO(milosgajods): allow to specify graph.Unmarshaler for unmarshaling graph stored in path.
// At the moment we use the only available unmarshaler: json.Unmarshaler.
func openFromFS(sys fs.FS) (map[string]*memory.Graph, error) {
	graphs := make(map[string]*memory.Graph)

	if err := fs.WalkDir(sys, ".", func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if e.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(sys, path)
		if err != nil {
			return err
		}

		// NOTE: DefaultLabel and attributes are overridden during deserialization.
		g, err := memory.NewGraph()
		if err != nil {
			return err
		}

		if err := json.Unmarshal(data, g); err != nil {
			return err
		}

		graphs[g.UID()] = g

		return nil
	}); err != nil {
		return nil, err
	}

	return graphs, nil
}
