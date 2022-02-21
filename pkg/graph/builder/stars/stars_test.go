package stars

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"

	"github.com/google/go-github/v34/github"
)

const (
	testPath = "testdata/sample.json"
)

func MustGraph(t *testing.T) *memory.Graph {
	g, err := memory.NewGraph()
	if err != nil {
		t.Fatalf("failed to create memory graph: %v", err)
	}
	return g
}

func MustBuilder(t *testing.T, g graph.Adder) *Stars {
	s, err := NewBuilder(g)
	if err != nil {
		t.Fatalf("failed to create a stars builder: %v", err)
	}
	return s
}

func TestBuildGraph(t *testing.T) {
	g := MustGraph(t)
	b := MustBuilder(t, g)

	data, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("failed to read test data from %s: %v", testPath, err)
	}

	var repos []*github.StarredRepository
	if err := json.Unmarshal(data, &repos); err != nil {
		t.Fatalf("failed to unmarshal GitHub repos: %v", err)
	}

	reposChan := make(chan interface{})
	errChan := make(chan error)
	go func() {
		errChan <- b.Build(context.Background(), reposChan)
	}()

	reposChan <- repos
	close(reposChan)

	if err := <-errChan; err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}

	// NOTE(milosgajdos): these two tests are silly, but there is no easy way
	// to tell from the sample data what nodes will have what ID assigned.

	if to := g.To(0); to.Len() < 0 {
		t.Errorf("invalid number of nodes: %d", to.Len())
	}

	g.HasEdgeFromTo(0, 1)
}
