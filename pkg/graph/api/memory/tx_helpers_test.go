package memory

import (
	"context"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/memory"
)

func MustNode(t *testing.T, id int64, opts ...memory.Option) *memory.Node {
	n, err := memory.NewNode(id, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func MustGraph(t *testing.T, opts ...memory.Option) *memory.Graph {
	g, err := memory.NewGraph(opts...)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func MustDB(t *testing.T, dsn string) *DB {
	db, err := NewDB(dsn)
	if err != nil {
		t.Fatalf("failed creating new DB: %v", err)
	}
	return db
}

func MustOpenDB(t *testing.T, dsn string) *DB {
	db := MustDB(t, dsn)
	if err := db.Open(); err != nil {
		t.Fatalf("failed opening DB: %v", err)
	}
	return db
}

// nolint:revive
func MustOpenTx(t *testing.T, ctx context.Context, dsn string) *Tx {
	db := MustOpenDB(t, dsn)
	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

// nolint:revive
func MustAddGraph(t *testing.T, ctx context.Context, tx *Tx, g *memory.Graph) {
	tx.db.Lock()
	defer tx.db.Unlock()

	if g == nil {
		t.Fatal("can not add nil graph")
	}
	tx.db.db[g.UID()] = g
}

// nolint:revive
func MustAddNode(t *testing.T, ctx context.Context, tx *Tx, uid string, opts ...memory.Option) *memory.Node {
	tx.db.Lock()
	defer tx.db.Unlock()

	g, ok := tx.db.db[uid]
	if !ok {
		t.Fatalf("graph %s not found", uid)
	}

	node := g.NewNode()
	n, err := memory.NewNode(node.ID(), opts...)
	if err != nil {
		t.Fatalf("failed creating new node: %v", err)
	}
	g.AddNode(n)

	return n
}
