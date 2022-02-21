package memory

import (
	"context"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"
)

func MustNodeServiceWithGraph(t *testing.T, dsn, uid string) *NodeService {
	db := MustOpenDB(t, dsn)
	g := MustGraph(t, memory.WithUID(uid))

	db.Lock()
	defer db.Unlock()

	if _, ok := db.db[uid]; ok {
		t.Fatalf("graph %s already exists", uid)
	}
	db.db[uid] = g

	ns, err := NewNodeService(db)
	if err != nil {
		t.Fatal(err)
	}
	return ns
}

func MustNodeService(t *testing.T, dsn string) *NodeService {
	db := MustOpenDB(t, dsn)
	db.Lock()
	defer db.Unlock()

	ns, err := NewNodeService(db)
	if err != nil {
		t.Fatal(err)
	}
	return ns
}

func TestCreateNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		testGraphUID := "testUID"

		ns := MustNodeServiceWithGraph(t, MemoryDSN, testGraphUID)

		n := &api.Node{Label: "testNode", Attrs: map[string]interface{}{"foo": 1}}

		if err := ns.CreateNode(context.TODO(), testGraphUID, n); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		ns := MustNodeService(t, MemoryDSN)

		n := &api.Node{Label: "testNode"}

		if err := ns.CreateNode(context.TODO(), "rangraphuid", n); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})
}

func TestFindNodeByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		testGraphUID := "testUID"

		ns := MustNodeServiceWithGraph(t, MemoryDSN, testGraphUID)

		n := &api.Node{Label: "testNode"}

		if err := ns.CreateNode(context.TODO(), testGraphUID, n); err != nil {
			t.Fatal(err)
		}

		node, err := ns.FindNodeByID(context.TODO(), testGraphUID, n.ID)
		if err != nil {
			t.Fatal(err)
		}

		if n.ID != node.ID {
			t.Fatalf("Expected ID: %d, got: %d", n.ID, node.ID)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		testGraphUID := "testUID"

		ns := MustNodeServiceWithGraph(t, MemoryDSN, testGraphUID)

		if _, err := ns.FindNodeByID(context.TODO(), testGraphUID, -1000); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		ns := MustNodeService(t, MemoryDSN)

		if _, err := ns.FindNodeByID(context.TODO(), "randgraphuid", -1000); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})
}

func TestFindNodes(t *testing.T) {
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
		{"LabelOnly_WithNegOffset", graphUID, api.NodeFilter{Label: &testLabel, Offset: -1, Limit: 10}, 2, 2, false},
		{"LabelOnly_WithLargeOffset", graphUID, api.NodeFilter{Label: &testLabel, Offset: 100, Limit: 10}, 0, 2, false},
	}

	ns := MustNodeService(t, testDir)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gx, n, err := ns.FindNodes(context.TODO(), graphUID, tc.filter)
			if !tc.expErr && err != nil {
				t.Fatal(err)
			}

			if n != tc.expMatches {
				t.Errorf("expected graphs: %d, got: %d", tc.expMatches, n)
			}

			if tc.expRes != len(gx) {
				t.Errorf("expected results: %d, got: %d", tc.expRes, len(gx))
			}
		})
	}

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		ns := MustNodeService(t, MemoryDSN)

		if err := ns.DeleteNodeByID(context.TODO(), "randuid", 300); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})
}

func TestUpdateNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		testGraphUID := "testUID"

		ns := MustNodeServiceWithGraph(t, MemoryDSN, testGraphUID)

		n := &api.Node{Label: "testNode", Attrs: map[string]interface{}{"foo": 1}}

		if err := ns.CreateNode(context.TODO(), testGraphUID, n); err != nil {
			t.Fatal(err)
		}

		if err := ns.CreateNode(context.TODO(), testGraphUID, n); err != nil {
			t.Fatal(err)
		}

		newLabel := "NewLabel"
		fooKey, fooVal := "foo", "fooVal"
		update := api.NodeUpdate{
			Label: &newLabel,
			Attrs: map[string]interface{}{
				fooKey: fooVal,
			},
		}

		node, err := ns.UpdateNode(context.TODO(), testGraphUID, n.ID, update)
		if err != nil {
			t.Fatal(err)
		}

		if node.Label != newLabel {
			t.Fatalf("expected label: %s, got: %s", newLabel, node.Label)
		}

		if val := node.Attrs[fooKey]; val != fooVal {
			t.Fatalf("expected value for key %s: %v, got: %v", fooKey, fooVal, val)
		}
	})

	t.Run("ErrNodeNotFound", func(t *testing.T) {
		testGraphUID := "testUID"

		ns := MustNodeServiceWithGraph(t, MemoryDSN, testGraphUID)

		newLabel := "NewLabel"
		update := api.NodeUpdate{
			Label: &newLabel,
		}

		if _, err := ns.UpdateNode(context.TODO(), testGraphUID, 3000, update); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		ns := MustNodeService(t, MemoryDSN)

		if _, err := ns.UpdateNode(context.TODO(), "randuid", 3000, api.NodeUpdate{}); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})
}

func TestDeleteNodeByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		testGraphUID := "testUID"

		ns := MustNodeServiceWithGraph(t, MemoryDSN, testGraphUID)

		n := &api.Node{Label: "testNode", Attrs: map[string]interface{}{"foo": 1}}

		if err := ns.CreateNode(context.TODO(), testGraphUID, n); err != nil {
			t.Fatal(err)
		}

		if err := ns.DeleteNodeByID(context.TODO(), testGraphUID, n.ID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		ns := MustNodeService(t, MemoryDSN)

		if err := ns.DeleteNodeByID(context.TODO(), "randUID", 300); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})
}

func TestDeleteNodeByUID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		testGraphUID := "testUID"
		ns := MustNodeServiceWithGraph(t, MemoryDSN, testGraphUID)

		n := &api.Node{Label: "testNode", Attrs: map[string]interface{}{"foo": 1}}

		if err := ns.CreateNode(context.TODO(), testGraphUID, n); err != nil {
			t.Fatal(err)
		}

		if err := ns.DeleteNodeByUID(context.TODO(), testGraphUID, n.UID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("NodeNotFound", func(t *testing.T) {
		testGraphUID := "testUID"
		ns := MustNodeServiceWithGraph(t, MemoryDSN, testGraphUID)

		if err := ns.DeleteNodeByUID(context.TODO(), testGraphUID, "doesntMatter"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})

	t.Run("GraphNotFound", func(t *testing.T) {
		ns := MustNodeService(t, MemoryDSN)

		if err := ns.DeleteNodeByUID(context.TODO(), "randUID", "doesntMatter"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})
}
