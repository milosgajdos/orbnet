package memory

import (
	"reflect"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/style"
)

const (
	testWeight     = 2.0
	testNode1ID    = 1
	testNode2ID    = 2
	testNode1Label = "node1Label"
	testNode2Label = "node2Label"
)

func MustNode(t *testing.T, id int64, label string, attrs map[string]interface{}) *Node {
	n, err := NewNode(id, WithLabel(label), WithAttrs(attrs))
	if err != nil {
		t.Fatalf("failed to create new node: %v", err)
	}
	return n
}

func TestNewEdge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	attrs := map[string]interface{}{
		"foo": "bar",
	}

	from := MustNode(t, testNode1ID, testNode1Label, attrs)
	to := MustNode(t, testNode2ID, testNode2Label, attrs)

	e, err := NewEdge(from, to, WithLabel(testLabel), WithAttrs(attrs))
	if err != nil {
		t.Fatalf("failed to create new edge: %v", err)
	}

	if uid := e.UID(); uid == "" {
		t.Errorf("empty edge UID")
	}

	// TODO(milosgajdos): should we use big.NewFloat.Cmp?
	if w := e.Weight(); w != DefaultWeight {
		t.Errorf("expected weight: %f, got: %f", DefaultWeight, w)
	}

	newWeight := 20.0
	e.SetWeight(newWeight)
	if w := e.Weight(); w != newWeight {
		t.Errorf("expected weight: %f, got: %f", newWeight, w)
	}

	if f := e.From(); f.ID() != from.ID() {
		t.Errorf("expected from ID: %d, got: %d", from.ID(), f.ID())
	}

	if tx := e.To(); tx.ID() != to.ID() {
		t.Errorf("expected to ID: %d, got: %d", to.ID(), tx.ID())
	}

	re := e.ReversedEdge()

	if f := re.From(); f.ID() != to.ID() {
		t.Errorf("expected from ID: %d, got: %d", to.ID(), f.ID())
	}
	if tx := re.To(); tx.ID() != from.ID() {
		t.Errorf("expected to ID: %d, got: %d", tx.ID(), from.ID())
	}

	if l := e.Label(); l != testLabel {
		t.Errorf("expected label: %s, got: %s", testLabel, l)
	}

	newLabel := "newLabel"
	e.SetLabel(newLabel)
	if l := e.Label(); l != newLabel {
		t.Errorf("expected label: %s, got: %s", newLabel, l)
	}

	newUID := "newUID"
	e.SetUID(newUID)
	if uid := e.UID(); uid != newUID {
		t.Errorf("expected UID: %s, got: %s", newUID, uid)
	}

	if a := e.Attrs(); !reflect.DeepEqual(a, attrs) {
		t.Errorf("expected attrs: %v, got: %v", attrs, a)
	}

	if s := e.Style(); s != style.DefaultStyleType {
		t.Errorf("expected style: %s, got: %s", style.DefaultStyleType, s)
	}

	if s := e.Shape(); s != style.DefaultEdgeShape {
		t.Errorf("expected shape: %s, got: %s", style.DefaultEdgeShape, s)
	}

	if c := e.Color(); c != style.DefaultEdgeColor {
		t.Errorf("expected color: %v, got: %v", style.DefaultEdgeColor, c)
	}

	if a := e.Attributes(); len(a) != len(attrs) {
		t.Errorf("expected attributes count: %d, got: %d", len(attrs), len(a))
	}
}

func TestNewEdgeWithOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	attrs := map[string]interface{}{
		"foo": "bar",
	}

	from := MustNode(t, testNode1ID, testNode1Label, attrs)
	to := MustNode(t, testNode2ID, testNode2Label, attrs)

	e, err := NewEdge(from, to, WithLabel(testLabel), WithAttrs(attrs), WithWeight(testWeight))
	if err != nil {
		t.Fatalf("failed to create new edge: %v", err)
	}

	// TODO(milosgajdos): should we use big.NewFloat.Cmp?
	if w := e.Weight(); w != testWeight {
		t.Errorf("expected weight: %f, got: %f", testWeight, w)
	}
}
