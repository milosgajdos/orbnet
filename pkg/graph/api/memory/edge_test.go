package memory

import (
	"context"
	"reflect"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"
)

func MustAddNodeEdgeService(t *testing.T, es *EdgeService, uid string, opts ...memory.Option) *memory.Node {
	es.db.Lock()
	defer es.db.Unlock()

	g, ok := es.db.db[uid]
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

func MustEdgeServiceWithGraph(t *testing.T, dsn, uid string) *EdgeService {
	es := MustEdgeService(t, dsn, uid)

	es.db.Lock()
	defer es.db.Unlock()

	g := MustGraph(t, memory.WithUID(uid))
	if _, ok := es.db.db[uid]; ok {
		t.Fatalf("graph %s already exists", uid)
	}
	es.db.db[uid] = g
	return es
}

func MustEdgeService(t *testing.T, dsn, uid string) *EdgeService {
	db := MustOpenDB(t, dsn)
	db.Lock()
	defer db.Unlock()

	es, err := NewEdgeService(db)
	if err != nil {
		t.Fatal(err)
	}
	return es
}

func TestCreateEdge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		testGraphUID := "testUID"

		es := MustEdgeServiceWithGraph(t, MemoryDSN, testGraphUID)
		n := MustAddNodeEdgeService(t, es, testGraphUID)
		n2 := MustAddNodeEdgeService(t, es, testGraphUID)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := es.CreateEdge(context.TODO(), testGraphUID, e); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrInvalidEdge", func(t *testing.T) {
		testGraphUID := "randUID"
		es := MustEdgeServiceWithGraph(t, MemoryDSN, testGraphUID)

		e := &api.Edge{Label: "testEdge"}

		if err := es.CreateEdge(context.TODO(), testGraphUID, e); api.ErrorCode(err) != api.EINVALID {
			t.Fatalf("expected error: %s, got: %s", api.EINVALID, api.ErrorCode(err))
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		testGraphUID := "randUID"
		es := MustEdgeService(t, MemoryDSN, testGraphUID)

		e := &api.Edge{Label: "testEdge"}

		if err := es.CreateEdge(context.TODO(), testGraphUID, e); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})
}

func TestFindEdgeByUID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		testGraphUID := "testUID"

		es := MustEdgeServiceWithGraph(t, MemoryDSN, testGraphUID)
		n := MustAddNodeEdgeService(t, es, testGraphUID)
		n2 := MustAddNodeEdgeService(t, es, testGraphUID)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := es.CreateEdge(context.TODO(), testGraphUID, e); err != nil {
			t.Fatal(err)
		}

		edge, err := es.FindEdgeByUID(context.TODO(), testGraphUID, e.UID)
		if err != nil {
			t.Fatal(err)
		}

		if e.UID != edge.UID {
			t.Fatalf("Expected UID: %s, got: %s", e.UID, edge.UID)
		}
	})

	t.Run("ErrNotFound", func(t *testing.T) {
		testGraphUID := "testUID"

		es := MustEdgeServiceWithGraph(t, MemoryDSN, testGraphUID)

		if _, err := es.FindEdgeByUID(context.TODO(), testGraphUID, "foo"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		testGraphUID := "randUID"
		es := MustEdgeService(t, MemoryDSN, testGraphUID)

		if _, err := es.FindEdgeByUID(context.TODO(), testGraphUID, "foo"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})
}

func TestUpdateEdgeBetween(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		testGraphUID := "testUID"
		es := MustEdgeServiceWithGraph(t, MemoryDSN, testGraphUID)
		n := MustAddNodeEdgeService(t, es, testGraphUID)
		n2 := MustAddNodeEdgeService(t, es, testGraphUID)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := es.CreateEdge(context.TODO(), testGraphUID, e); err != nil {
			t.Fatal(err)
		}

		weight := 100.0
		label := "testLabel"
		attrs := map[string]interface{}{
			"foo": 10,
			"bar": "bar0",
		}

		update := api.EdgeUpdate{Weight: &weight, Label: &label, Attrs: attrs}

		e2, err := es.UpdateEdgeBetween(context.TODO(), testGraphUID, e.Source, e.Target, update)
		if err != nil {
			t.Fatal(err)
		}

		if w := e2.Weight; w != weight {
			t.Errorf("expected weight: %f, got: %f", weight, w)
		}

		if l := e2.Label; l != label {
			t.Errorf("expected label: %s, got: %s", label, l)
		}

		a := e2.Attrs
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

	t.Run("ErrEdgeNotFound", func(t *testing.T) {
		testGraphUID := "testUID"
		es := MustEdgeServiceWithGraph(t, MemoryDSN, testGraphUID)
		n := MustAddNodeEdgeService(t, es, testGraphUID)
		n2 := MustAddNodeEdgeService(t, es, testGraphUID)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := es.CreateEdge(context.TODO(), testGraphUID, e); err != nil {
			t.Fatal(err)
		}

		if _, err := es.UpdateEdgeBetween(context.TODO(), testGraphUID, n.ID(), -200, api.EdgeUpdate{}); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}

		if _, err := es.UpdateEdgeBetween(context.TODO(), testGraphUID, -100, n.ID(), api.EdgeUpdate{}); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		testGraphUID := "randUID"
		es := MustEdgeService(t, MemoryDSN, testGraphUID)

		if _, err := es.UpdateEdgeBetween(context.TODO(), testGraphUID, 1, 2, api.EdgeUpdate{}); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})
}

func TestDeleteEdge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		testGraphUID := "testUID"

		es := MustEdgeServiceWithGraph(t, MemoryDSN, testGraphUID)
		n := MustAddNodeEdgeService(t, es, testGraphUID)
		n2 := MustAddNodeEdgeService(t, es, testGraphUID)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := es.CreateEdge(context.TODO(), testGraphUID, e); err != nil {
			t.Fatal(err)
		}

		if err := es.DeleteEdge(context.TODO(), testGraphUID, e.UID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		testGraphUID := "randUID"
		es := MustEdgeService(t, MemoryDSN, testGraphUID)

		if err := es.DeleteEdge(context.TODO(), testGraphUID, "foo"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})
}

func TestDeleteEdgeBetween(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		testGraphUID := "testUID"

		es := MustEdgeServiceWithGraph(t, MemoryDSN, testGraphUID)
		n := MustAddNodeEdgeService(t, es, testGraphUID)
		n2 := MustAddNodeEdgeService(t, es, testGraphUID)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := es.CreateEdge(context.TODO(), testGraphUID, e); err != nil {
			t.Fatal(err)
		}

		if err := es.DeleteEdgeBetween(context.TODO(), testGraphUID, e.Source, e.Target); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		testGraphUID := "randUID"
		es := MustEdgeService(t, MemoryDSN, testGraphUID)

		if err := es.DeleteEdgeBetween(context.TODO(), testGraphUID, 1, 2); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})
}
