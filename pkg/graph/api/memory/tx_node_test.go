package memory

import (
	"context"
	"reflect"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

func TestTxCreateNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := &api.Node{Label: "testNode", Attrs: map[string]interface{}{"foo": 1}}

		if err := tx.CreateNode(ctx, uid, n); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		n := &api.Node{Label: "testNode"}

		if err := tx.CreateNode(ctx, "foo", n); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})
}

func TestTxFindNodeByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := &api.Node{Label: "testNode"}

		if err := tx.CreateNode(ctx, uid, n); err != nil {
			t.Fatal(err)
		}

		n2, err := tx.FindNodeByID(ctx, uid, n.ID)
		if err != nil {
			t.Fatal(err)
		}

		if n2.ID() != n.ID {
			t.Errorf("expected graph with ID: %d, got: %d", n.ID, n2.ID())
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := &api.Node{Label: "testNode"}

		if err := tx.CreateNode(ctx, uid, n); err != nil {
			t.Fatal(err)
		}

		if _, err := tx.FindNodeByID(ctx, uid, 3000); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})
}

func TestTxFindNodeByiUID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := &api.Node{Label: "testNode"}

		if err := tx.CreateNode(ctx, uid, n); err != nil {
			t.Fatal(err)
		}

		n2, err := tx.FindNodeByUID(ctx, uid, n.UID)
		if err != nil {
			t.Fatal(err)
		}

		if n2.UID() != n.UID {
			t.Errorf("expected graph with UID: %s, got: %s", n.UID, n2.UID())
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := &api.Node{Label: "testNode"}

		if err := tx.CreateNode(ctx, uid, n); err != nil {
			t.Fatal(err)
		}

		if _, err := tx.FindNodeByUID(ctx, uid, "sdfdf"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})
}

func TestTxFindNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// NOTE(milosgajdos): these come from testdata/sample.json
	graphUID := "cc099040-9dab-4f3d-848e-3046912aa281"

	testID := int64(1)
	testLabel := "Repo"
	toTestLabel := "Topic"
	fromTestLabel := "Owner"

	randID := int64(20)
	randLabel := "randLabel"

	testCases := []struct {
		name       string
		uid        string
		filter     api.NodeFilter
		expRes     int
		expMatches int
		expErr     bool
	}{
		// TODO: add tests for UID filter
		{"GraphNotFound", "fooUID", api.NodeFilter{}, 0, 0, true},
		{"EmptyFilter", graphUID, api.NodeFilter{}, 6, 6, false},
		{"IDNoMatch", graphUID, api.NodeFilter{ID: &randID}, 0, 0, false},
		{"IDMatch", graphUID, api.NodeFilter{ID: &testID}, 1, 1, false},
		{"IDMatch_LabelMatch", graphUID, api.NodeFilter{ID: &testID, Label: &testLabel}, 1, 1, false},
		{"IDMatch_LabelNoMatch", graphUID, api.NodeFilter{ID: &testID, Label: &randLabel}, 0, 0, false},
		{"ToIDMatch", graphUID, api.NodeFilter{To: &testID}, 2, 2, false},
		{"ToIDMatch_LabelMatch", graphUID, api.NodeFilter{To: &testID, Label: &toTestLabel}, 2, 2, false},
		{"ToIDMatch_LabelNoMatch", graphUID, api.NodeFilter{To: &testID, Label: &randLabel}, 0, 0, false},
		{"FromIDMatch", graphUID, api.NodeFilter{From: &testID}, 2, 2, false},
		{"FromIDMatch_LabelMatch", graphUID, api.NodeFilter{From: &testID, Label: &fromTestLabel}, 2, 2, false},
		{"FromIDMatch_LabelNoMatch", graphUID, api.NodeFilter{From: &testID, Label: &randLabel}, 0, 0, false},
		{"LabelOnly", graphUID, api.NodeFilter{Label: &testLabel}, 2, 2, false},
		{"LabelOnly_WithOffsetLimit", graphUID, api.NodeFilter{Label: &testLabel, Offset: 1, Limit: 1}, 1, 2, false},
		{"LabelOnly_WithNegOffset", graphUID, api.NodeFilter{Label: &testLabel, Offset: -1}, 2, 2, false},
		{"LabelOnly_WithLargeOffset", graphUID, api.NodeFilter{Label: &testLabel, Offset: 100}, 0, 2, false},
	}

	tx := MustOpenTx(t, context.TODO(), testDir)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gx, n, err := tx.FindNodes(context.TODO(), tc.uid, tc.filter)
			if !tc.expErr && err != nil {
				t.Fatal(err)
			}

			if tc.expRes != len(gx) {
				t.Errorf("expected results: %d, got: %d", tc.expRes, len(gx))
			}

			if n != tc.expMatches {
				t.Errorf("expected matches: %d, got: %d", tc.expMatches, n)
			}
		})
	}
}

func TestTxUpdateNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("GraphNotFound", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		if _, err := tx.UpdateNode(ctx, "randomUID", 10, api.NodeUpdate{}); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})

	t.Run("NodeNotFound", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := &api.Node{Label: "testNode"}

		if err := tx.CreateNode(ctx, uid, n); err != nil {
			t.Fatal(err)
		}

		if _, err := tx.UpdateNode(ctx, uid, 1000, api.NodeUpdate{}); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})

	t.Run("LabelAttrs", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := &api.Node{Label: "testNode"}

		if err := tx.CreateNode(ctx, uid, n); err != nil {
			t.Fatal(err)
		}

		label := "testLabel"
		attrs := map[string]interface{}{
			"foo": 10,
			"bar": "bar0",
		}

		g2, err := tx.UpdateNode(ctx, uid, n.ID, api.NodeUpdate{Label: &label, Attrs: attrs})
		if err != nil {
			t.Fatal(err)
		}

		if l := g2.Label(); l != label {
			t.Errorf("expected label: %s, got: %s", label, l)
		}

		a := g2.Attrs()
		for k, v := range attrs {
			val, ok := a[k]
			if !ok {
				t.Errorf("could not find any values for key: %s", k)
			}

			if !reflect.DeepEqual(val, v) {
				t.Errorf("expected value: %v, for key: %s, got: %v", v, k, val)
			}
		}
	})

	t.Run("NoUpdate", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := &api.Node{Label: "testNode"}

		if err := tx.CreateNode(ctx, uid, n); err != nil {
			t.Fatal(err)
		}

		n2, err := tx.UpdateNode(ctx, uid, n.ID, api.NodeUpdate{})
		if err != nil {
			t.Fatal(err)
		}

		if n.Label != n2.Label() {
			t.Errorf("expected label: %s, got: %s", n.Label, n2.Label())
		}
	})
}

func TestTxDeleteNodeByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := &api.Node{Label: "testNode"}

		if err := tx.CreateNode(ctx, uid, n); err != nil {
			t.Fatal(err)
		}

		if err := tx.DeleteNodeByID(ctx, uid, n.ID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GraphNotFound", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		if err := tx.DeleteNodeByID(ctx, "randomUID", 10); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})
}

func TestTxDeleteNodeByUID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := &api.Node{Label: "testNode"}

		if err := tx.CreateNode(ctx, uid, n); err != nil {
			t.Fatal(err)
		}

		if err := tx.DeleteNodeByUID(ctx, uid, n.UID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("NodeNotFound", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		if err := tx.DeleteNodeByUID(ctx, g.UID(), "randomUID"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})

	t.Run("GraphNotFound", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		if err := tx.DeleteNodeByUID(ctx, "randomUID", "doesntMatter"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})
}
