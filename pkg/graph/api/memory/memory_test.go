package memory

import (
	"context"
	"errors"

	"testing"
)

const (
	testDir = "testdata"
)

func TestNewDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		if _, err := NewDB("foo"); err != nil {
			t.Fatalf("failed creating new DB: %v", err)
		}
	})
}

func TestOpenDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("memory", func(t *testing.T) {
		db := MustDB(t, MemoryDSN)
		if err := db.Open(); err != nil {
			t.Fatalf("failed to open memory DB: %v", err)
		}
	})

	t.Run("FS", func(t *testing.T) {
		db := MustDB(t, testDir)
		if err := db.Open(); err != nil {
			t.Fatalf("failed to open FS DB: %v", err)
		}
	})

	t.Run("EmptyDSN", func(t *testing.T) {
		db := MustDB(t, "")
		if err := db.Open(); err == nil {
			t.Fatalf("expected error, got: %v", err)
		}
	})
}

func TestCloseDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("open", func(t *testing.T) {
		db := MustOpenDB(t, MemoryDSN)
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}
	})

	t.Run("closed", func(t *testing.T) {
		db := MustOpenDB(t, MemoryDSN)
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}
	})
}

func TestBeginTx(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("open", func(t *testing.T) {
		db := MustOpenDB(t, MemoryDSN)
		if _, err := db.BeginTx(context.Background()); err != nil {
			t.Fatalf("failed starting a transaction: %v", err)
		}
	})

	t.Run("closed", func(t *testing.T) {
		db := MustOpenDB(t, MemoryDSN)
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close DB: %v", err)
		}
		if _, err := db.BeginTx(context.Background()); !errors.Is(err, ErrDBClosed) {
			t.Fatalf("expected error: %v, got: %v", ErrDBClosed, err)
		}
	})
}
