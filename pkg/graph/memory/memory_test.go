package memory

import (
	"testing"
)

func TestNewGraph(t *testing.T) {
	g, err := NewGraph()
	if err != nil {
		t.Fatalf("failed to create new stars graph: %v", err)
	}

	if uid := g.UID(); uid == "" {
		t.Error("expected non-empty UID")
	}

	if typ := g.Type(); typ != DefaultType {
		t.Errorf("expected type: %s, got: %s", DefaultType, typ)
	}

	if l := g.Label(); l != DefaultLabel {
		t.Errorf("expected label: %s, got: %s", DefaultLabel, l)
	}

	newLabel := "newLabel"
	g.SetLabel(newLabel)
	if l := g.Label(); l != newLabel {
		t.Errorf("expected label: %s, got: %s", newLabel, l)
	}

	newUID := "newUID"
	g.SetUID(newUID)
	if uid := g.UID(); uid != newUID {
		t.Errorf("expected UID: %s, got: %s", newUID, uid)
	}

	if a := g.Attrs(); a == nil {
		t.Error("expected non-empty attributes")
	}
}

func TestNewGraphWithOpts(t *testing.T) {
	uid := "FooID"
	label := "fooLabel"

	g, err := NewGraph(WithUID(uid), WithLabel(label))
	if err != nil {
		t.Fatalf("failed to create new stars graph: %v", err)
	}

	if u := g.UID(); u != uid {
		t.Errorf("expected uid: %s, got: %s", uid, u)
	}

	if typ := g.Type(); typ != DefaultType {
		t.Errorf("expected type: %s, got: %s", DefaultType, typ)
	}

	if l := g.Label(); l != label {
		t.Errorf("expected label: %s, got: %s", label, l)
	}

}
