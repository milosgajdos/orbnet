package sqlite

import (
	"context"
	"reflect"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

// MustCreateEdge creates an edge in the database. Fatal on error.
func MustCreateEdge(ctx context.Context, tb testing.TB, db *DB, graphUID string, edge *api.Edge) {
	tb.Helper()
	es, err := NewEdgeService(db)
	if err != nil {
		tb.Fatal(err)
	}
	if err = es.CreateEdge(ctx, graphUID, edge); err != nil {
		tb.Fatal(err)
	}
}

func MustEdgeService(t *testing.T, db *DB) *EdgeService {
	es, err := NewEdgeService(db)
	if err != nil {
		t.Fatal(err)
	}
	return es
}

func TestEdgeService_CreateEdge(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		es := MustEdgeService(t, db)

		graphUID := "graph1"
		MustCreateGraph(context.Background(), t, db, &api.Graph{UID: graphUID, Label: StringPtr("Graph1")})

		// Create nodes to be used in edges
		node1 := &api.Node{UID: "node1", Label: StringPtr("Node1")}
		node2 := &api.Node{UID: "node2", Label: StringPtr("Node2")}
		MustCreateNode(context.Background(), t, db, graphUID, node1)
		MustCreateNode(context.Background(), t, db, graphUID, node2)

		edge := &api.Edge{
			UID:    "edge1",
			Source: node1.UID,
			Target: node2.UID,
			Label:  "Edge1",
			Weight: 1.0,
			Attrs: map[string]interface{}{
				"foo": "bar",
			},
		}

		// Create new edge & verify it was created
		if err := es.CreateEdge(context.Background(), graphUID, edge); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrUIDRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		es := MustEdgeService(t, db)

		graphUID := "graph1"
		MustCreateGraph(context.Background(), t, db, &api.Graph{UID: graphUID, Label: StringPtr("Graph1")})

		// Create nodes to be used in edges
		node1 := &api.Node{UID: "node1", Label: StringPtr("Node1")}
		node2 := &api.Node{UID: "node2", Label: StringPtr("Node2")}
		MustCreateNode(context.Background(), t, db, graphUID, node1)
		MustCreateNode(context.Background(), t, db, graphUID, node2)

		if err := es.CreateEdge(context.Background(), graphUID, &api.Edge{}); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestEdgeService_FindEdge(t *testing.T) {
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		es := MustEdgeService(t, db)

		if _, err := es.FindEdgeByUID(context.Background(), "graph1", "garbage"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestEdgeService_FindEdges(t *testing.T) {
	t.Run("Source", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		es := MustEdgeService(t, db)

		ctx := context.Background()
		graphUID := "graph1"
		MustCreateGraph(ctx, t, db, &api.Graph{UID: graphUID, Label: StringPtr("Graph1")})

		// Create nodes to be used in edges
		node1 := &api.Node{UID: "node1", Label: StringPtr("Node1")}
		node2 := &api.Node{UID: "node2", Label: StringPtr("Node2")}
		node3 := &api.Node{UID: "node3", Label: StringPtr("Node3")}
		MustCreateNode(ctx, t, db, graphUID, node1)
		MustCreateNode(ctx, t, db, graphUID, node2)
		MustCreateNode(ctx, t, db, graphUID, node3)

		// Create edges
		edge1 := &api.Edge{UID: "edge1", Source: node1.UID, Target: node2.UID, Label: "Edge1"}
		edge2 := &api.Edge{UID: "edge2", Source: node1.UID, Target: node3.UID, Label: "Edge2"}
		MustCreateEdge(ctx, t, db, graphUID, edge1)
		MustCreateEdge(ctx, t, db, graphUID, edge2)

		source := node1.UID
		ex, n, err := es.FindEdges(ctx, graphUID, api.EdgeFilter{Source: &source})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(ex), 2; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		}
		if got, want := n, 2; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})

	t.Run("Target", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		es := MustEdgeService(t, db)

		ctx := context.Background()
		graphUID := "graph1"
		MustCreateGraph(ctx, t, db, &api.Graph{UID: graphUID, Label: StringPtr("Graph1")})

		// Create nodes to be used in edges
		node1 := &api.Node{UID: "node1", Label: StringPtr("Node1")}
		node2 := &api.Node{UID: "node2", Label: StringPtr("Node2")}
		node3 := &api.Node{UID: "node3", Label: StringPtr("Node3")}
		MustCreateNode(ctx, t, db, graphUID, node1)
		MustCreateNode(ctx, t, db, graphUID, node2)
		MustCreateNode(ctx, t, db, graphUID, node3)

		// Create edges
		edge1 := &api.Edge{UID: "edge1", Source: node1.UID, Target: node2.UID, Label: "Edge1"}
		edge2 := &api.Edge{UID: "edge2", Source: node3.UID, Target: node2.UID, Label: "Edge2"}
		MustCreateEdge(ctx, t, db, graphUID, edge1)
		MustCreateEdge(ctx, t, db, graphUID, edge2)

		target := node2.UID
		ex, n, err := es.FindEdges(ctx, graphUID, api.EdgeFilter{Target: &target})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(ex), 2; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		}
		if got, want := n, 2; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})
}

func TestEdgeService_UpdateEdge(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		es := MustEdgeService(t, db)

		ctx := context.Background()
		graphUID := "graph1"
		MustCreateGraph(ctx, t, db, &api.Graph{UID: graphUID, Label: StringPtr("Graph1")})

		// Create nodes to be used in edges
		node1 := &api.Node{UID: "node1", Label: StringPtr("Node1")}
		node2 := &api.Node{UID: "node2", Label: StringPtr("Node2")}
		MustCreateNode(ctx, t, db, graphUID, node1)
		MustCreateNode(ctx, t, db, graphUID, node2)

		// Create edge
		edge := &api.Edge{
			UID:    "edge1",
			Source: node1.UID,
			Target: node2.UID,
			Label:  "Edge1",
			Weight: 1.0,
			Attrs:  map[string]interface{}{"foo": "bar"},
		}
		MustCreateEdge(ctx, t, db, graphUID, edge)

		// Update edge
		newLabel := "Edge1Updated"
		newWeight := 2.0
		newAttrs := map[string]interface{}{"foo": "baz"}
		ue, err := es.UpdateEdgeBetween(ctx, graphUID, node1.UID, node2.UID,
			api.EdgeUpdate{
				Label:  &newLabel,
				Weight: &newWeight,
				Attrs:  newAttrs,
			})
		if err != nil {
			t.Fatal(err)
		}

		if ue.Label != newLabel {
			t.Fatalf("label=%v, want=%v", ue.Label, newLabel)
		}
		if ue.Weight != newWeight {
			t.Fatalf("weight=%v, want=%v", ue.Weight, newWeight)
		}
		if !reflect.DeepEqual(ue.Attrs, newAttrs) {
			t.Fatalf("expected: %#v, got: %#v", newAttrs, ue.Attrs)
		}

		// Verify the update
		got, err := es.FindEdgeByUID(ctx, graphUID, edge.UID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Label != newLabel {
			t.Fatalf("label=%v, want=%v", got.Label, newLabel)
		}
		if got.Weight != newWeight {
			t.Fatalf("weight=%v, want=%v", got.Weight, newWeight)
		}
		if !reflect.DeepEqual(got.Attrs, newAttrs) {
			t.Fatalf("expected: %#v, got: %#v", newAttrs, got.Attrs)
		}
	})

	t.Run("NonExistent", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		es := MustEdgeService(t, db)

		_, err := es.UpdateEdgeBetween(context.Background(), "graph1", "source1", "target1",
			api.EdgeUpdate{Label: StringPtr("foo")})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestEdgeService_DeleteEdge(t *testing.T) {
	t.Run("ByUID_OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		es := MustEdgeService(t, db)

		ctx := context.Background()
		graphUID := "graph1"
		MustCreateGraph(ctx, t, db, &api.Graph{UID: graphUID, Label: StringPtr("Graph1")})

		// Create nodes to be used in edges
		node1 := &api.Node{UID: "node1", Label: StringPtr("Node1")}
		node2 := &api.Node{UID: "node2", Label: StringPtr("Node2")}
		MustCreateNode(ctx, t, db, graphUID, node1)
		MustCreateNode(ctx, t, db, graphUID, node2)

		// Create edge
		edge := &api.Edge{UID: "edge1", Source: node1.UID, Target: node2.UID, Label: "Edge1"}
		MustCreateEdge(ctx, t, db, graphUID, edge)

		// Delete edge
		if err := es.DeleteEdge(ctx, graphUID, edge.UID); err != nil {
			t.Fatal(err)
		}

		// Verify deletion
		if _, err := es.FindEdgeByUID(context.Background(), graphUID, edge.UID); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("BySourceTarget_OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		es := MustEdgeService(t, db)

		ctx := context.Background()
		graphUID := "graph1"
		MustCreateGraph(ctx, t, db, &api.Graph{UID: graphUID, Label: StringPtr("Graph1")})

		// Create nodes to be used in edges
		node1 := &api.Node{UID: "node1", Label: StringPtr("Node1")}
		node2 := &api.Node{UID: "node2", Label: StringPtr("Node2")}
		MustCreateNode(ctx, t, db, graphUID, node1)
		MustCreateNode(ctx, t, db, graphUID, node2)

		// Create edge
		edge := &api.Edge{UID: "edge1", Source: node1.UID, Target: node2.UID, Label: "Edge1"}
		MustCreateEdge(ctx, t, db, graphUID, edge)

		// Delete edge by source and target
		if err := es.DeleteEdgeBetween(ctx, graphUID, node1.UID, node2.UID); err != nil {
			t.Fatal(err)
		}

		// Verify deletion
		if _, err := es.FindEdgeByUID(context.Background(), graphUID, edge.UID); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}
