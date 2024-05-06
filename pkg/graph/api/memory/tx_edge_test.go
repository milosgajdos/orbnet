package memory

import (
	"context"
	"reflect"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

func TestTxCreateEdge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := MustAddNode(t, ctx, tx, uid)
		n2 := MustAddNode(t, ctx, tx, uid)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := tx.CreateEdge(ctx, uid, e); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		n := &api.Edge{Label: "testEdge"}

		if err := tx.CreateEdge(ctx, "foo", n); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})
}

func TestTxFindEdgeByUID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := MustAddNode(t, ctx, tx, uid)
		n2 := MustAddNode(t, ctx, tx, uid)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := tx.CreateEdge(ctx, uid, e); err != nil {
			t.Fatal(err)
		}

		e2, err := tx.FindEdgeByUID(ctx, uid, e.UID)
		if err != nil {
			t.Fatal(err)
		}

		if e2.UID() != e.UID {
			t.Fatalf("expected edge: %s, got: %s", e.UID, e2.UID())
		}
	})

	t.Run("GraphNotFound", func(t *testing.T) {
		tx := MustOpenTx(t, context.TODO(), MemoryDSN)

		if _, err := tx.FindEdgeByUID(context.TODO(), "foo", "bar"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})
}

func TestTxFindEdges(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// NOTE(milosgajdos): these come from testdata/sample.json
	graphUID := "dd099040-9dab-4f3d-848e-3046912aa281"

	simpleSrc := int64(0)
	simpleTarg := int64(1)
	simpleLabel := "HasLang"

	randSrc := int64(200)
	randTarg := int64(300)
	randLabel := "randLabel"

	testCases := []struct {
		name       string
		uid        string
		filter     api.EdgeFilter
		expRes     int
		expMatches int
		expErr     bool
	}{
		{"GraphNotFound", "fooUID", api.EdgeFilter{}, 0, 0, true},
		{"EmptyFilter", graphUID, api.EdgeFilter{}, 9, 9, false},
		{"SrcTargetMatch", graphUID, api.EdgeFilter{Source: &simpleSrc, Target: &simpleTarg}, 1, 1, false},
		{"SrcTargetNoMatch", graphUID, api.EdgeFilter{Source: &randSrc, Target: &randTarg}, 0, 0, false},
		{"SrcTarget_LabelMatch", graphUID, api.EdgeFilter{Source: &simpleSrc, Target: &simpleTarg, Label: &simpleLabel}, 1, 1, false},
		{"SrcTarget_LabelNoMatch", graphUID, api.EdgeFilter{Source: &simpleSrc, Target: &simpleTarg, Label: &randLabel}, 0, 0, false},
		{"Src_LabelMatch", graphUID, api.EdgeFilter{Source: &simpleSrc, Label: &simpleLabel}, 2, 2, false},
		{"Src_LabelNoMatch", graphUID, api.EdgeFilter{Source: &simpleSrc, Label: &randLabel}, 0, 0, false},
		{"Target_LabelMatch", graphUID, api.EdgeFilter{Target: &simpleTarg, Label: &simpleLabel}, 2, 2, false},
		{"Target_LabelNoMatch", graphUID, api.EdgeFilter{Target: &simpleTarg, Label: &randLabel}, 0, 0, false},
		{"LabelOnly", graphUID, api.EdgeFilter{Label: &simpleLabel}, 4, 4, false},
		{"LabelOnly_NoMatch", graphUID, api.EdgeFilter{Label: &randLabel}, 0, 0, false},
		{"LabelOnly_WithOffsetLimit", graphUID, api.EdgeFilter{Label: &simpleLabel, Offset: 1, Limit: 1}, 1, 4, false},
		{"LabelOnly_WithNegOffset", graphUID, api.EdgeFilter{Label: &simpleLabel, Offset: -1}, 4, 4, false},
		{"LabelOnly_WithLargeOffset", graphUID, api.EdgeFilter{Label: &simpleLabel, Offset: 100}, 0, 4, false},
	}

	tx := MustOpenTx(t, context.TODO(), testDir)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ex, n, err := tx.FindEdges(context.TODO(), tc.uid, tc.filter)
			if !tc.expErr && err != nil {
				t.Fatal(err)
			}

			if tc.expRes != len(ex) {
				t.Errorf("expected results: %d, got: %d", tc.expRes, len(ex))
			}

			if n != tc.expMatches {
				t.Errorf("expected matches: %d, got: %d", tc.expMatches, n)
			}
		})
	}
}

func TestTxUpdateEdgeBetween(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := MustAddNode(t, ctx, tx, uid)
		n2 := MustAddNode(t, ctx, tx, uid)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := tx.CreateEdge(ctx, uid, e); err != nil {
			t.Fatal(err)
		}

		weight := 100.0
		label := "testLabel"
		attrs := map[string]interface{}{
			"foo": 10,
			"bar": "bar0",
		}

		update := api.EdgeUpdate{Weight: &weight, Label: &label, Attrs: attrs}

		e2, err := tx.UpdateEdgeBetween(ctx, uid, e.Source, e.Target, update)
		if err != nil {
			t.Fatal(err)
		}

		if w := e2.Weight(); w != weight {
			t.Errorf("expected weight: %f, got: %f", weight, w)
		}

		if l := e2.Label(); l != label {
			t.Errorf("expected label: %s, got: %s", label, l)
		}

		a := e2.Attrs()
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

	t.Run("EdgeNotFound", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := MustAddNode(t, ctx, tx, uid)
		n2 := MustAddNode(t, ctx, tx, uid)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := tx.CreateEdge(ctx, uid, e); err != nil {
			t.Fatal(err)
		}

		if _, err := tx.UpdateEdgeBetween(ctx, uid, -100, -200, api.EdgeUpdate{}); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})

	t.Run("NoUpdate", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := MustAddNode(t, ctx, tx, uid)
		n2 := MustAddNode(t, ctx, tx, uid)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Weight: 3.0,
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := tx.CreateEdge(ctx, uid, e); err != nil {
			t.Fatal(err)
		}

		e2, err := tx.UpdateEdgeBetween(ctx, uid, e.Source, e.Target, api.EdgeUpdate{})
		if err != nil {
			t.Fatal(err)
		}

		if e.Weight != e2.Weight() {
			t.Errorf("expected weight: %f, got: %f", e.Weight, e2.Weight())
		}

		if e.Label != e2.Label() {
			t.Errorf("expected label: %s, got: %s", e.Label, e2.Label())
		}
	})

	t.Run("GraphNotFound", func(t *testing.T) {
		tx := MustOpenTx(t, context.TODO(), MemoryDSN)

		if _, err := tx.UpdateEdgeBetween(context.TODO(), "foo", 0, 1, api.EdgeUpdate{}); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})

}

func TestTxDeleteEdge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := MustAddNode(t, ctx, tx, uid)
		n2 := MustAddNode(t, ctx, tx, uid)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := tx.CreateEdge(ctx, uid, e); err != nil {
			t.Fatal(err)
		}

		if err := tx.DeleteEdge(ctx, uid, e.UID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		if err := tx.DeleteEdge(ctx, "foo", "bar"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})
}

func TestTxDeleteEdgeBetween(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		g := MustGraph(t)
		tx := MustOpenTx(t, ctx, MemoryDSN)
		MustAddGraph(t, ctx, tx, g)

		uid := g.UID()
		n := MustAddNode(t, ctx, tx, uid)
		n2 := MustAddNode(t, ctx, tx, uid)

		e := &api.Edge{
			Source: n.ID(),
			Target: n2.ID(),
			Label:  "testEdge",
			Attrs:  map[string]interface{}{"foo": 1},
		}

		if err := tx.CreateEdge(ctx, uid, e); err != nil {
			t.Fatal(err)
		}

		if err := tx.DeleteEdgeBetween(ctx, uid, e.Source, e.Target); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrGraphNotFound", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		if err := tx.DeleteEdgeBetween(ctx, "foo", 0, 1); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})
}
