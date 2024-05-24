package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

func MustGraphService(t *testing.T, dsn string) *GraphService {
	db := MustOpenDB(t, dsn)
	gs, err := NewGraphService(db)
	if err != nil {
		t.Fatal(err)
	}
	return gs
}

func MustClosedGraphService(t *testing.T, dsn string) *GraphService {
	db := MustDB(t, dsn)
	gs, err := NewGraphService(db)
	if err != nil {
		t.Fatal(err)
	}
	return gs
}

func TestCreateGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		gs := MustGraphService(t, MemoryDSN)
		ag := &api.Graph{
			Label: StringPtr("Foo"),
			Attrs: map[string]interface{}{
				"foo": "bar",
				"car": 10,
			},
		}

		if err := gs.CreateGraph(context.TODO(), ag); err != nil {
			t.Fatal(err)
		}

		if ag.UID == "" {
			t.Fatal("expected non-emnpty UID")
		}
	})

	t.Run("DuplicateUID", func(t *testing.T) {
		gs := MustGraphService(t, MemoryDSN)

		ag := &api.Graph{
			Label: StringPtr("Foo"),
		}

		if err := gs.CreateGraph(context.TODO(), ag); err != nil {
			t.Fatal(err)
		}

		ag2 := &api.Graph{
			UID:   ag.UID,
			Label: StringPtr("Foo"),
		}

		if err := gs.CreateGraph(context.TODO(), ag2); err == nil {
			t.Fatalf("expected error, got: %v", err)
		}
	})

	t.Run("ClosedDB", func(t *testing.T) {
		gs := MustClosedGraphService(t, MemoryDSN)

		ag := &api.Graph{
			Label: StringPtr("Foo"),
		}

		if err := gs.CreateGraph(context.TODO(), ag); !errors.Is(err, ErrDBClosed) {
			t.Fatalf("expected error: %v, got: %v", ErrDBClosed, err)
		}
	})
}

func TestFindGraphByUID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		gs := MustGraphService(t, MemoryDSN)

		ag := &api.Graph{
			Label: StringPtr("Foo"),
			Attrs: map[string]interface{}{
				"foo": "bar",
				"car": 10,
			},
		}

		if err := gs.CreateGraph(context.TODO(), ag); err != nil {
			t.Fatal(err)
		}

		g, err := gs.FindGraphByUID(context.TODO(), ag.UID)
		if err != nil {
			t.Fatal(err)
		}

		if g.UID != ag.UID {
			t.Fatalf("Expected UID: %s, got: %s", ag.UID, g.UID)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		gs := MustGraphService(t, MemoryDSN)

		if _, err := gs.FindGraphByUID(context.TODO(), "garbageUID"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("expected error: %s, got: %s", api.ENOTFOUND, api.ErrorCode(err))
		}
	})

	t.Run("ClosedDB", func(t *testing.T) {
		gs := MustClosedGraphService(t, MemoryDSN)

		if _, err := gs.FindGraphByUID(context.TODO(), "foo"); !errors.Is(err, ErrDBClosed) {
			t.Fatalf("expected error: %v, got: %v", ErrDBClosed, err)
		}
	})
}

func TestFindGraphs(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// NOTE(milosgajdos): these come from testdata/sample.json
	testUID := "cc099040-9dab-4f3d-848e-3046912aa281"
	testLabel := "test"

	randUID := "randUID"
	randLabel := "randLabel"

	testCases := []struct {
		name       string
		filter     api.GraphFilter
		expRes     int
		expMatches int
		expErr     bool
	}{
		{"UID", api.GraphFilter{UID: &randUID}, 0, 0, false},
		{"UID", api.GraphFilter{UID: &testUID}, 1, 1, false},
		{"UID", api.GraphFilter{UID: &testUID, Label: &testLabel}, 1, 1, false},
		{"UID", api.GraphFilter{UID: &testUID, Label: &randLabel}, 0, 0, false},
		{"NoTypeLabel", api.GraphFilter{}, 2, 2, false},
		{"TestLabel", api.GraphFilter{Label: &testLabel}, 2, 2, false},
		{"RandLabel", api.GraphFilter{Label: &randLabel}, 0, 0, false},
		{"TestLabelLimitOffset", api.GraphFilter{Label: &testLabel, Offset: 100}, 0, 2, false},
		{"TestLabelLimitOffset", api.GraphFilter{Label: &testLabel, Offset: -1}, 2, 2, false},
		{"LimitOffset", api.GraphFilter{Label: &testLabel, Offset: 1}, 1, 2, false},
		{"LimitOffset", api.GraphFilter{Label: &testLabel, Offset: 1, Limit: 1}, 1, 2, false},
		{"LimitOffset", api.GraphFilter{Label: &testLabel, Offset: 1, Limit: 10}, 1, 2, false},
	}

	gs := MustGraphService(t, testDir)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gx, n, err := gs.FindGraphs(context.TODO(), tc.filter)
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

	t.Run("ClosedDB", func(t *testing.T) {
		gs := MustClosedGraphService(t, MemoryDSN)

		if _, _, err := gs.FindGraphs(context.TODO(), api.GraphFilter{}); !errors.Is(err, ErrDBClosed) {
			t.Fatalf("expected error: %v, got: %v", ErrDBClosed, err)
		}
	})
}

func TestUpdateGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		gs := MustGraphService(t, MemoryDSN)

		ag := &api.Graph{
			Label: StringPtr("Foo"),
			Attrs: map[string]interface{}{
				"foo": "bar",
				"car": 10,
			},
		}

		if err := gs.CreateGraph(context.TODO(), ag); err != nil {
			t.Fatal(err)
		}

		newLabel := "NewLabel"
		fooKey, fooVal := "foo", "fooVal"
		update := api.GraphUpdate{
			Label: &newLabel,
			Attrs: map[string]interface{}{
				fooKey: fooVal,
			},
		}

		g, err := gs.UpdateGraph(context.TODO(), ag.UID, update)
		if err != nil {
			t.Fatal(err)
		}

		if *g.Label != newLabel {
			t.Fatalf("expected label: %s, got: %s", newLabel, *g.Label)
		}

		if val := g.Attrs[fooKey]; val != fooVal {
			t.Fatalf("expected value for key %s: %v, got: %v", fooKey, fooVal, val)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		gs := MustGraphService(t, MemoryDSN)

		ag := &api.Graph{
			Label: StringPtr("Foo"),
			Attrs: map[string]interface{}{
				"foo": "bar",
				"car": 10,
			},
		}

		if err := gs.CreateGraph(context.TODO(), ag); err != nil {
			t.Fatal(err)
		}

		newLabel := "NewLabel"
		update := api.GraphUpdate{
			Label: &newLabel,
		}

		if _, err := gs.UpdateGraph(context.TODO(), "garbageUID", update); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})

	t.Run("ClosedDB", func(t *testing.T) {
		gs := MustClosedGraphService(t, MemoryDSN)

		if _, err := gs.UpdateGraph(context.TODO(), "garbageUID", api.GraphUpdate{}); !errors.Is(err, ErrDBClosed) {
			t.Fatalf("expected error: %v, got: %v", ErrDBClosed, err)
		}
	})
}

func TestDeleteGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		gs := MustGraphService(t, MemoryDSN)

		ag := &api.Graph{
			Label: StringPtr("Foo"),
			Attrs: map[string]interface{}{
				"foo": "bar",
				"car": 10,
			},
		}

		if err := gs.CreateGraph(context.TODO(), ag); err != nil {
			t.Fatal(err)
		}

		if err := gs.DeleteGraph(context.TODO(), ag.UID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ClosedDB", func(t *testing.T) {
		gs := MustClosedGraphService(t, MemoryDSN)

		if err := gs.DeleteGraph(context.TODO(), "foo"); !errors.Is(err, ErrDBClosed) {
			t.Fatalf("expected error: %v, got: %v", ErrDBClosed, err)
		}
	})
}
