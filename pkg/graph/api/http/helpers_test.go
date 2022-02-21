package http

import (
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/api/memory"
)

const (
	testDir = "testdata"
)

func MustServer(t *testing.T) *Server {
	s, err := NewServer()
	if err != nil {
		t.Fatalf("failed to created new server: %v", err)
	}
	return s
}

func MustOpenDB(t *testing.T, dsn string) *memory.DB {
	db, err := memory.NewDB(dsn)
	if err != nil {
		t.Fatalf("failed creating new DB: %v", err)
	}
	if err := db.Open(); err != nil {
		t.Fatalf("failed opening DB: %v", err)
	}
	return db
}
