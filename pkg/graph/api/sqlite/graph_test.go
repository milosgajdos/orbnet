package sqlite

import (
	"context"
	"reflect"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

// MustCreateGraph creates a graph in the database. Fatal on error.
func MustCreateGraph(ctx context.Context, tb testing.TB, db *DB, graph *api.Graph) {
	tb.Helper()
	gs, err := NewGraphService(db)
	if err != nil {
		tb.Fatal(err)
	}
	if err = gs.CreateGraph(ctx, graph); err != nil {
		tb.Fatal(err)
	}
}

func MustGraphService(t *testing.T, db *DB) *GraphService {
	gs, err := NewGraphService(db)
	if err != nil {
		t.Fatal(err)
	}
	return gs
}

func TestGraphService_CreateGraph(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		gs := MustGraphService(t, db)

		g := &api.Graph{
			UID:   "Foobar",
			Label: StringPtr("Foo"),
			Attrs: map[string]interface{}{
				"foo": "bar",
				"car": 10,
			},
		}

		// Create new graph & verify ID and timestamps are set.
		if err := gs.CreateGraph(context.Background(), g); err != nil {
			t.Fatal(err)
		}
		if got, want := g.UID, "Foobar"; got != want {
			t.Fatalf("UID=%v, want %v", got, want)
		}
		if g.CreatedAt.IsZero() {
			t.Fatal("expected created at")
		}
		if g.UpdatedAt.IsZero() {
			t.Fatal("expected updated at")
		}

		g2 := &api.Graph{
			UID:   "Foobar2",
			Label: StringPtr("Foo"),
		}
		if err := gs.CreateGraph(context.Background(), g2); err != nil {
			t.Fatal(err)
		}
		if got, want := g2.UID, "Foobar2"; got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		}

		got, err := gs.FindGraphByUID(context.Background(), "Foobar")
		if err != nil {
			t.Fatal(err)
		}
		if got, want := *got.Label, *g.Label; got != want {
			t.Fatalf("label=%v, want %v", got, want)
		}
		for k, v := range got.Attrs {
			wantVal, ok := g.Attrs[k]
			if !ok {
				t.Fatalf("missing attribute: %v", k)
			}
			if intVal, ok := wantVal.(int); ok {
				if !reflect.DeepEqual(int64(intVal), v) {
					t.Fatalf("expected: %#v, got: %#v", wantVal, v)
				}
				continue
			}
			if !reflect.DeepEqual(wantVal, v) {
				t.Fatalf("expected: %#v, got: %#v", wantVal, v)
			}
		}
	})

	t.Run("ErrUIDRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		gs := MustGraphService(t, db)

		if err := gs.CreateGraph(context.Background(), &api.Graph{}); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGraphService_FindGraph(t *testing.T) {
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		gs := MustGraphService(t, db)

		if _, err := gs.FindGraphByUID(context.Background(), "garbage"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestGraphService_FindGraphs(t *testing.T) {
	t.Run("UID", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		gs := MustGraphService(t, db)

		ctx := context.Background()
		MustCreateGraph(ctx, t, db, &api.Graph{UID: "foo", Label: StringPtr("lfoo")})
		MustCreateGraph(ctx, t, db, &api.Graph{UID: "bar", Label: StringPtr("lbar")})
		MustCreateGraph(ctx, t, db, &api.Graph{UID: "car", Label: StringPtr("lcar")})
		MustCreateGraph(ctx, t, db, &api.Graph{UID: "dar", Label: StringPtr("ldar")})

		uid, label := "bar", "lbar"
		gx, n, err := gs.FindGraphs(ctx, api.GraphFilter{UID: &uid})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(gx), 1; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		}
		if got, want := *gx[0].Label, label; got != want {
			t.Fatalf("label=%v, want %v", got, want)
		}
		if got, want := n, 1; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})
}

func TestGraphService_UpdateGraph(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		gs := MustGraphService(t, db)

		ctx := context.Background()
		attrs := map[string]interface{}{
			"foo": "bar",
			"car": 5,
		}
		MustCreateGraph(ctx, t, db, &api.Graph{UID: "foo", Label: StringPtr("lfoo"), Attrs: attrs})

		newLabel := "labFoo"
		newAttrs := map[string]interface{}{
			"foo": "car",
		}

		ug, err := gs.UpdateGraph(ctx, "foo", api.GraphUpdate{
			Label: &newLabel,
			Attrs: newAttrs,
		})
		if err != nil {
			t.Fatal(err)
		}

		if *ug.Label != newLabel {
			t.Fatalf("label=%v, want=%v", *ug.Label, newLabel)
		}
		expAttrs := map[string]interface{}{
			"foo": "car",
			"car": int64(5),
		}
		if !reflect.DeepEqual(expAttrs, ug.Attrs) {
			t.Fatalf("expected: %#v, got: %#v", expAttrs, ug.Attrs)
		}

		got, err := gs.FindGraphByUID(context.Background(), "foo")
		if err != nil {
			t.Fatal(err)
		}
		if got, want := *got.Label, *ug.Label; got != want {
			t.Fatalf("label=%v, want %v", got, want)
		}
		if !reflect.DeepEqual(expAttrs, got.Attrs) {
			t.Fatalf("expected: %#v, got: %#v", expAttrs, got.Attrs)
		}
	})

	t.Run("NonExistent", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		gs := MustGraphService(t, db)

		_, err := gs.UpdateGraph(context.Background(), "foo", api.GraphUpdate{Label: StringPtr("foo")})
		if err == nil {
			t.Fatal("expeected error")
		}
	})
}

func TestGraphService_DeleteGraph(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		gs := MustGraphService(t, db)

		ctx := context.Background()
		MustCreateGraph(ctx, t, db, &api.Graph{UID: "foo", Label: StringPtr("lfoo")})

		if err := gs.DeleteGraph(ctx, "foo"); err != nil {
			t.Fatal(err)
		}

		if _, err := gs.FindGraphByUID(context.Background(), "foo"); api.ErrorCode(err) != api.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}
