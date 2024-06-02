package sqlite

import (
	"context"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"
)

func MustLoader(tb testing.TB, db *DB) *Loader {
	l, err := NewLoader(db)
	if err != nil {
		tb.Fatal(err)
	}
	return l
}

func TestLoader_Load(t *testing.T) {
	db := MustOpenDB(t)
	defer db.Close()
	l := MustLoader(t, db)
	s := MustSyncer(t, db)

	ctx := context.Background()

	// Create a test graph
	g, err := memory.NewGraph()
	if err != nil {
		t.Fatalf("failed to create new graph: %v", err)
	}

	// Create two nodes
	nodeOpts := []memory.Option{
		memory.WithUID("node1"),
		memory.WithLabel("Node 1"),
		memory.WithAttrs(map[string]interface{}{"key": "value"}),
	}

	node1, err := memory.NewNode(1, nodeOpts...)
	if err != nil {
		t.Fatalf("failed to create new node: %v", err)
	}
	g.AddNode(node1)

	nodeOpts = []memory.Option{
		memory.WithUID("node2"),
		memory.WithLabel("Node 2"),
		memory.WithAttrs(map[string]interface{}{"key": "value"}),
	}

	node2, err := memory.NewNode(2, nodeOpts...)
	if err != nil {
		t.Fatalf("failed to create new node: %v", err)
	}
	g.AddNode(node2)

	// Create an edge between the two nodes
	edgeOpts := []memory.Option{
		memory.WithUID("edge1"),
		memory.WithLabel("Edge 1"),
		memory.WithWeight(1.0),
		memory.WithAttrs(map[string]interface{}{"key2": "value2"}),
	}

	edge, err := memory.NewEdge(node1, node2, edgeOpts...)
	if err != nil {
		t.Fatalf("failed to create new edge: %v", err)
	}
	g.SetWeightedEdge(edge)

	// Sync the graph to the database
	if err := s.Sync(ctx, g); err != nil {
		t.Fatalf("failed to sync graph: %v", err)
	}

	// Load the graph from the database
	loadedGraph, err := l.Load(ctx, g.UID())
	if err != nil {
		t.Fatalf("failed to load graph: %v", err)
	}

	// Verify the graph
	if loadedGraph.UID() != g.UID() {
		t.Errorf("expected graph UID %v, got %v", g.UID(), loadedGraph.UID())
	}
	if loadedGraph.Label() != g.Label() {
		t.Errorf("expected graph label %v, got %v", g.Label(), loadedGraph.Label())
	}
	if len(loadedGraph.Attrs()) != len(g.Attrs()) {
		t.Errorf("expected graph attributes %v, got %v", g.Attrs(), loadedGraph.Attrs())
	}

	// Verify the nodes
	nodes := loadedGraph.Nodes()
	nodeMap := make(map[string]graph.Node)
	for nodes.Next() {
		n := nodes.Node().(graph.Node)
		nodeMap[n.UID()] = n
	}

	// Verify node1
	if n, ok := nodeMap[node1.UID()]; !ok {
		t.Errorf("node1 not found in loaded graph")
	} else {
		if n.UID() != node1.UID() {
			t.Errorf("expected node1 UID %v, got %v", node1.UID(), n.UID())
		}
		if n.Label() != node1.Label() {
			t.Errorf("expected node1 label %v, got %v", node1.Label(), n.Label())
		}
		if len(n.Attrs()) != len(node1.Attrs()) {
			t.Errorf("expected node1 attributes %v, got %v", node1.Attrs(), n.Attrs())
		}
	}

	// Verify node2
	if n, ok := nodeMap[node2.UID()]; !ok {
		t.Errorf("node2 not found in loaded graph")
	} else {
		if n.UID() != node2.UID() {
			t.Errorf("expected node2 UID %v, got %v", node2.UID(), n.UID())
		}
		if n.Label() != node2.Label() {
			t.Errorf("expected node2 label %v, got %v", node2.Label(), n.Label())
		}
		if len(n.Attrs()) != len(node2.Attrs()) {
			t.Errorf("expected node2 attributes %v, got %v", node2.Attrs(), n.Attrs())
		}
	}

	// Verify the edges
	edges := loadedGraph.Edges()
	edgeMap := make(map[string]graph.Edge)
	for edges.Next() {
		e := edges.Edge().(graph.Edge)
		edgeMap[e.UID()] = e
	}

	// Verify edge
	if e, ok := edgeMap[edge.UID()]; !ok {
		t.Errorf("edge not found in loaded graph")
	} else {
		if e.UID() != edge.UID() {
			t.Errorf("expected edge UID %v, got %v", edge.UID(), e.UID())
		}
		if e.Label() != edge.Label() {
			t.Errorf("expected edge label %v, got %v", edge.Label(), e.Label())
		}
		if e.Weight() != edge.Weight() {
			t.Errorf("expected edge weight %v, got %v", edge.Weight(), e.Weight())
		}
		if len(e.Attrs()) != len(edge.Attrs()) {
			t.Errorf("expected edge attributes %v, got %v", edge.Attrs(), e.Attrs())
		}
		fromUID, edgeFromUID := e.From().(graph.Node).UID(), edge.From().(graph.Node).UID()
		if fromUID != edgeFromUID {
			t.Errorf("expected edge from node UID %v, got %v", edgeFromUID, fromUID)
		}
		toUID, edgeToUID := e.To().(graph.Node).UID(), edge.To().(graph.Node).UID()
		if toUID != edgeToUID {
			t.Errorf("expected edge to node UID %v, got %v", edgeToUID, toUID)
		}
	}
}
