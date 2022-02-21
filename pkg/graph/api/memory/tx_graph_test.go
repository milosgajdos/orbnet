package memory

import (
	"context"
	"reflect"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

func TestTxCreateGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)
		g := &api.Graph{Label: "testGraph", Attrs: map[string]interface{}{"foo": 1}}

		if err := tx.CreateGraph(ctx, g); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrAlreadyExists", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		g := &api.Graph{Label: "testGraph"}
		if err := tx.CreateGraph(ctx, g); err != nil {
			t.Fatal(err)
		}

		g2 := &api.Graph{Label: "testGraph", UID: g.UID}
		if err := tx.CreateGraph(ctx, g2); api.ErrorCode(err) != api.ECONFLICT {
			t.Fatal(err)
		}
	})
}

func TestTxFindGraphByUID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		g := &api.Graph{Label: "testGraph"}
		if err := tx.CreateGraph(ctx, g); err != nil {
			t.Fatal(err)
		}

		g2, err := tx.FindGraphByUID(ctx, g.UID)
		if err != nil {
			t.Fatal(err)
		}

		if g2.UID() != g.UID {
			t.Errorf("expected graph with UID: %s, got: %s", g.UID, g2.UID())
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		g := &api.Graph{Label: "testGraph"}
		if err := tx.CreateGraph(ctx, g); err != nil {
			t.Fatal(err)
		}

		if _, err := tx.FindGraphByUID(ctx, "randomUID"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})
}

func TestTxFindGraphs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// NOTE(milosgajdos): these come from testdata/sample.json
	testUID := "cc099040-9dab-4f3d-848e-3046912aa281"
	testLabel := "test"
	testType := "weighted_directed"

	randUID := "randUID"
	randLabel := "randLabel"
	randType := "randType"

	testCases := []struct {
		name       string
		filter     api.GraphFilter
		expRes     int
		expMatches int
		expErr     bool
	}{
		{"EmptyFilter", api.GraphFilter{}, 2, 2, false},
		{"UIDNonexist", api.GraphFilter{UID: &randUID}, 0, 0, false},
		{"UIDOk", api.GraphFilter{UID: &testUID}, 1, 1, false},
		{"UIDOkLabelOk", api.GraphFilter{UID: &testUID, Label: &testLabel}, 1, 1, false},
		{"UIDOkLabelOkTypeOk", api.GraphFilter{UID: &testUID, Label: &testLabel, Type: &testType}, 1, 1, false},
		{"UIDOkLabelNonexist", api.GraphFilter{UID: &testUID, Label: &randLabel}, 0, 0, false},
		{"UIDOkTypeNonexist", api.GraphFilter{UID: &testUID, Type: &randType}, 0, 0, false},
		{"TypeNoneLabelNone", api.GraphFilter{}, 2, 2, false},
		{"TypeNoneLabelOk", api.GraphFilter{Label: &testLabel}, 2, 2, false},
		{"TypeOkLabelOk", api.GraphFilter{Label: &testLabel, Type: &testType}, 2, 2, false},
		{"TypeNoneLabelNonexist", api.GraphFilter{Label: &randLabel}, 0, 0, false},
		{"TypeNonexistLabelNone", api.GraphFilter{Type: &randType}, 0, 0, false},
		{"LimitNoneOffsetLarge", api.GraphFilter{Type: &testType, Offset: 100}, 0, 2, false},
		{"LimitNoneOffsetNegative", api.GraphFilter{Type: &testType, Offset: -1}, 2, 2, false},
		{"LimitNoneOffsetOk", api.GraphFilter{Type: &testType, Offset: 1}, 1, 2, false},
		{"LimitOkOffsetOk", api.GraphFilter{Type: &testType, Offset: 1, Limit: 1}, 1, 2, false},
		{"LimitLargeOffsetOk", api.GraphFilter{Type: &testType, Offset: 1, Limit: 10}, 1, 2, false},
	}

	ctx := context.TODO()
	tx := MustOpenTx(t, ctx, testDir)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gx, n, err := tx.FindGraphs(ctx, tc.filter)
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

func TestTxUpdateGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("NotFound", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		g := &api.Graph{Label: "testGraph"}
		if err := tx.CreateGraph(ctx, g); err != nil {
			t.Fatal(err)
		}

		if _, err := tx.UpdateGraph(ctx, "randomUID", api.GraphUpdate{}); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatal(err)
		}
	})

	t.Run("LabelAttrs", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		g := &api.Graph{Label: "testGraph"}
		if err := tx.CreateGraph(ctx, g); err != nil {
			t.Fatal(err)
		}

		label := "testLabel"
		attrs := map[string]interface{}{
			"foo": 10,
			"bar": "bar0",
		}

		g2, err := tx.UpdateGraph(ctx, g.UID, api.GraphUpdate{Label: &label, Attrs: attrs})
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
		tx := MustOpenTx(t, ctx, MemoryDSN)

		g := &api.Graph{Label: "testGraph"}
		if err := tx.CreateGraph(ctx, g); err != nil {
			t.Fatal(err)
		}

		if _, err := tx.UpdateGraph(ctx, g.UID, api.GraphUpdate{}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestTxDeleteGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("OK", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		g := &api.Graph{Label: "testGraph"}
		if err := tx.CreateGraph(ctx, g); err != nil {
			t.Fatal(err)
		}

		if err := tx.DeleteGraph(ctx, g.UID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("NotExist", func(t *testing.T) {
		ctx := context.TODO()
		tx := MustOpenTx(t, ctx, MemoryDSN)

		if err := tx.DeleteGraph(ctx, "randomUID"); err != nil {
			t.Fatal(err)
		}
	})
}
