package sqlite

import (
	"context"
	"reflect"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

func MustNodeService(t *testing.T, db *DB) *NodeService {
	ns, err := NewNodeService(db)
	if err != nil {
		t.Fatal(err)
	}
	return ns
}

func MustCreateNode(ctx context.Context, tb testing.TB, db *DB, guid string, node *api.Node) {
	tb.Helper()
	gs, err := NewNodeService(db)
	if err != nil {
		tb.Fatal(err)
	}
	if err = gs.CreateNode(ctx, guid, node); err != nil {
		tb.Fatal(err)
	}
}

func TestNodeService_CreateNode(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		ns := MustNodeService(t, db)

		g := &api.Graph{
			UID:   "graph1cn",
			Label: StringPtr("Graph1"),
			Attrs: map[string]interface{}{
				"foo": "bar",
				"car": 10,
			},
		}

		ctx := context.Background()

		MustCreateGraph(ctx, t, db, g)

		nodeUID, noodeLabel := "node1cn", "Node1"
		n := &api.Node{
			UID:   nodeUID,
			Label: StringPtr(noodeLabel),
			Attrs: map[string]interface{}{
				"foo": "bar",
				"num": 10,
			},
		}

		if err := ns.CreateNode(ctx, g.UID, n); err != nil {
			t.Fatal(err)
		}
		if got, want := n.UID, nodeUID; got != want {
			t.Fatalf("UID=%v, want %v", got, want)
		}
		if n.CreatedAt.IsZero() {
			t.Fatal("expected created at")
		}
		if n.UpdatedAt.IsZero() {
			t.Fatal("expected updated at")
		}

		got, err := ns.FindNodeByUID(context.Background(), g.UID, nodeUID)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := *got.Label, *n.Label; got != want {
			t.Fatalf("label=%v, want %v", got, want)
		}
	})

	t.Run("ErrUIDRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		ns := MustNodeService(t, db)

		if err := ns.CreateNode(context.Background(), "foo", &api.Node{}); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestNodeService_FindNode(t *testing.T) {
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		ns := MustNodeService(t, db)

		if _, err := ns.FindNodeByUID(context.Background(), "graph1", "garbage"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestNodeService_FindNodes(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)
	ns := MustNodeService(t, db)
	ctx := context.Background()

	// Create a graph with nodes and edges
	graphUID := "graph1fnUID"
	MustCreateGraph(ctx, t, db, &api.Graph{UID: graphUID, Label: StringPtr("Graph1")})

	// Create nodes
	nodes := []*api.Node{
		{UID: "node1fn", Label: StringPtr("Node1fn")},
		{UID: "node2fn", Label: StringPtr("Node2fn")},
		{UID: "node3fn", Label: StringPtr("Node3fn")},
		{UID: "node4fn", Label: StringPtr("Node4fn")},
	}
	for _, node := range nodes {
		MustCreateNode(ctx, t, db, graphUID, node)
	}

	// Create edges
	_, err := db.db.ExecContext(ctx, `
		INSERT INTO edges (uid, graph, source, target, label)
		VALUES ('edge12', ?, 'node1fn', 'node2fn', 'Edge12'),
		       ('edge13', ?, 'node1fn', 'node3fn', 'Edge13'),
		       ('edge23', ?, 'node2fn', 'node3fn', 'Edge23'),
		       ('edge42', ?, 'node4fn', 'node2fn', 'Edge42'),
		       ('edge34', ?, 'node3fn', 'node4fn', 'Edge34')
	`, graphUID, graphUID, graphUID, graphUID, graphUID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("ID", func(t *testing.T) {
		nodeID, nodeUID := nodes[0].ID, nodes[0].UID
		nx, n, err := ns.FindNodes(ctx, graphUID, api.NodeFilter{ID: &nodeID})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(nx), 1; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		}
		if got, want := nx[0].UID, nodeUID; got != want {
			t.Fatalf("UID=%v, want %v", got, want)
		}
		if got, want := n, 1; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})

	t.Run("UID", func(t *testing.T) {
		uid := nodes[1].UID
		nx, n, err := ns.FindNodes(ctx, graphUID, api.NodeFilter{UID: &uid})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(nx), 1; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		}
		if got, want := nx[0].UID, uid; got != want {
			t.Fatalf("UID=%v, want %v", got, want)
		}
		if got, want := n, 1; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})

	t.Run("To", func(t *testing.T) {
		// node2fn node
		to := nodes[1].ID
		nx, n, err := ns.FindNodes(ctx, graphUID, api.NodeFilter{To: &to})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(nx), 2; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		}
		expNodes := map[string]struct{}{
			"node1fn": {},
			"node4fn": {},
		}
		for _, n := range nx {
			if _, ok := expNodes[n.UID]; !ok {
				t.Fatalf("exp node=%v", n.UID)
			}
		}
		if got, want := n, 2; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})

	t.Run("From", func(t *testing.T) {
		from := nodes[0].ID
		nx, n, err := ns.FindNodes(ctx, graphUID, api.NodeFilter{From: &from})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(nx), 2; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		}
		expNodes := map[string]struct{}{
			"node2fn": {},
			"node3fn": {},
		}
		for _, n := range nx {
			if _, ok := expNodes[n.UID]; !ok {
				t.Fatalf("exp node=%v", n.UID)
			}
		}
		if got, want := n, 2; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})

	t.Run("Label", func(t *testing.T) {
		label := "Node3fn"
		nx, n, err := ns.FindNodes(ctx, graphUID, api.NodeFilter{Label: &label})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(nx), 1; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		}
		if got, want := nx[0].UID, "node3fn"; got != want {
			t.Fatalf("UID=%v, want %v", got, want)
		}
		if got, want := n, 1; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})
}

func TestNodeService_UpdateNode(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		ns := MustNodeService(t, db)

		ctx := context.Background()
		graphUID := "graph1un"
		attrs := map[string]interface{}{
			"foo": "bar",
			"num": 5,
		}
		nodeUID := "node1un"
		MustCreateGraph(ctx, t, db, &api.Graph{UID: graphUID, Label: StringPtr("Graph1un")})
		MustCreateNode(ctx, t, db, graphUID, &api.Node{UID: nodeUID, Label: StringPtr("Node1un"), Attrs: attrs})

		newLabel := "Node1Updated"
		newAttrs := map[string]interface{}{
			"foo": "baz",
		}

		un, err := ns.UpdateNode(ctx, graphUID, 1, api.NodeUpdate{
			Label: &newLabel,
			Attrs: newAttrs,
		})
		if err != nil {
			t.Fatal(err)
		}

		if *un.Label != newLabel {
			t.Fatalf("label=%v, want=%v", *un.Label, newLabel)
		}
		expAttrs := map[string]interface{}{
			"foo": "baz",
			"num": int64(5),
		}
		if !reflect.DeepEqual(expAttrs, un.Attrs) {
			t.Fatalf("expected: %#v, got: %#v", expAttrs, un.Attrs)
		}

		got, err := ns.FindNodeByUID(context.Background(), graphUID, nodeUID)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := *got.Label, *un.Label; got != want {
			t.Fatalf("label=%v, want %v", got, want)
		}
		if !reflect.DeepEqual(expAttrs, got.Attrs) {
			t.Fatalf("expected: %#v, got: %#v", expAttrs, got.Attrs)
		}
	})

	t.Run("NonExistent", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		ns := MustNodeService(t, db)

		_, err := ns.UpdateNode(context.Background(), "graph1", 999, api.NodeUpdate{Label: StringPtr("NonExistent")})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestNodeService_DeleteNode(t *testing.T) {
	t.Run("ByUID_OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		ns := MustNodeService(t, db)

		ctx := context.Background()
		graphUID, nodeUID := "graph1dnUID", "node1dnUID"
		MustCreateGraph(ctx, t, db, &api.Graph{UID: graphUID, Label: StringPtr("Graph1")})
		MustCreateNode(ctx, t, db, graphUID, &api.Node{UID: nodeUID, Label: StringPtr("Node1")})

		if err := ns.DeleteNodeByUID(ctx, graphUID, nodeUID); err != nil {
			t.Fatal(err)
		}

		if _, err := ns.FindNodeByUID(context.Background(), graphUID, nodeUID); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("ByID_OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		ns := MustNodeService(t, db)

		ctx := context.Background()
		graphUID, nodeUID := "graph1dnID", "node1dnID"
		MustCreateGraph(ctx, t, db, &api.Graph{UID: graphUID, Label: StringPtr("Graph1")})
		MustCreateNode(ctx, t, db, graphUID, &api.Node{UID: nodeUID, Label: StringPtr("Node1")})

		node, err := ns.FindNodeByUID(ctx, graphUID, nodeUID)
		if err != nil {
			t.Fatal(err)
		}

		if err := ns.DeleteNodeByID(ctx, graphUID, node.ID); err != nil {
			t.Fatal(err)
		}

		if _, err := ns.FindNodeByID(context.Background(), graphUID, node.ID); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}
