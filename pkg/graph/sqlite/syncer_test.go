package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"
)

func MustSyncer(tb testing.TB, db *DB) *Syncer {
	s, err := NewSyncer(db)
	if err != nil {
		tb.Fatal(err)
	}
	return s
}

func TestSyncer_Sync(t *testing.T) {
	db := MustOpenDB(t)
	defer db.Close()
	s := MustSyncer(t, db)

	ctx := context.Background()

	// Create a test graph
	g, err := memory.NewGraph()
	if err != nil {
		t.Fatalf("failed to create new graph: %v", err)
	}

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

	// Sync the graph
	if err := s.Sync(ctx, g); err != nil {
		t.Fatalf("failed to sync graph: %v", err)
	}

	// Verify the graph was inserted correctly
	var graphUID, label, attrs string
	var createdAt, updatedAt time.Time
	err = db.db.QueryRowContext(ctx, `
		SELECT uid, label, attrs, created_at, updated_at
		FROM graphs WHERE uid = ?
	`, g.UID()).Scan(&graphUID, &label, &attrs, (*NullTime)(&createdAt), (*NullTime)(&updatedAt))
	if err != nil {
		t.Fatalf("failed to query graph: %v", err)
	}

	if graphUID != g.UID() {
		t.Errorf("expected graph UID %v, got %v", g.UID(), graphUID)
	}
	if label != g.Label() {
		t.Errorf("expected graph label %v, got %v", g.Label(), label)
	}

	graphAttrs, err := AttrsFromString(attrs)
	if err != nil {
		t.Fatal(err)
	}
	if len(graphAttrs) != len(g.Attrs()) {
		t.Errorf("expected graph attributes %v, got %v", g.Attrs(), graphAttrs)
	}

	// Verify the first node was inserted correctly
	var node1UID, node1Label, node1Attrs string
	err = db.db.QueryRowContext(ctx, `
		SELECT uid, label, attrs
		FROM nodes WHERE uid = ?
	`, node1.UID()).Scan(&node1UID, &node1Label, &node1Attrs)
	if err != nil {
		t.Fatalf("failed to query node1: %v", err)
	}

	if node1UID != node1.UID() {
		t.Errorf("expected node1 UID %v, got %v", node1.UID(), node1UID)
	}
	if node1Label != node1.Label() {
		t.Errorf("expected node1 label %v, got %v", node1.Label(), node1Label)
	}

	node1Attributes, err := AttrsFromString(node1Attrs)
	if err != nil {
		t.Fatal(err)
	}
	if len(node1Attributes) != len(node1.Attrs()) {
		t.Errorf("expected node1 attributes %v, got %v", node1.Attrs(), node1Attributes)
	}

	// Verify the second node was inserted correctly
	var node2UID, node2Label, node2Attrs string
	err = db.db.QueryRowContext(ctx, `
		SELECT uid, label, attrs
		FROM nodes WHERE uid = ?
	`, node2.UID()).Scan(&node2UID, &node2Label, &node2Attrs)
	if err != nil {
		t.Fatalf("failed to query node2: %v", err)
	}

	if node2UID != node2.UID() {
		t.Errorf("expected node2 UID %v, got %v", node2.UID(), node2UID)
	}
	if node2Label != node2.Label() {
		t.Errorf("expected node2 label %v, got %v", node2.Label(), node2Label)
	}

	node2Attributes, err := AttrsFromString(node2Attrs)
	if err != nil {
		t.Fatal(err)
	}
	if len(node2Attributes) != len(node2.Attrs()) {
		t.Errorf("expected node2 attributes %v, got %v", node2.Attrs(), node2Attributes)
	}

	// Verify the edge was inserted correctly
	var edgeUID, sourceUID, targetUID, edgeLabel, edgeAttrs string
	var edgeWeight float64
	err = db.db.QueryRowContext(ctx, `
		SELECT uid, source, target, label, weight, attrs
		FROM edges WHERE uid = ?
	`, edge.UID()).Scan(&edgeUID, &sourceUID, &targetUID, &edgeLabel, &edgeWeight, &edgeAttrs)
	if err != nil {
		t.Fatalf("failed to query edge: %v", err)
	}

	if edgeUID != edge.UID() {
		t.Errorf("expected edge UID %v, got %v", edge.UID(), edgeUID)
	}

	if fromUID := edge.From().(graph.Node).UID(); fromUID != sourceUID {
		t.Errorf("expected source UID %v, got %v", fromUID, sourceUID)
	}
	if toUID := edge.To().(graph.Node).UID(); toUID != targetUID {
		t.Errorf("expected target UID %v, got %v", toUID, targetUID)
	}
	if edgeLabel != edge.Label() {
		t.Errorf("expected edge label %v, got %v", edge.Label(), edgeLabel)
	}
	if edgeWeight != edge.Weight() {
		t.Errorf("expected edge weight %v, got %v", edge.Weight(), edgeWeight)
	}

	edgeAttributes, err := AttrsFromString(edgeAttrs)
	if err != nil {
		t.Fatal(err)
	}
	if len(edgeAttributes) != len(edge.Attrs()) {
		t.Errorf("expected edge attributes %v, got %v", edge.Attrs(), edgeAttributes)
	}
}
