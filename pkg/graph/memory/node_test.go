package memory

import (
	"reflect"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/style"
)

const (
	testID    = 100
	testLabel = "testLabel"
	testDotID = "testDotID"
	testUID   = "testUID"
)

func TestNewNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	attrs := map[string]interface{}{
		"foo": "bar",
	}

	n, err := NewNode(testID, WithLabel(testLabel), WithAttrs(attrs))
	if err != nil {
		t.Fatalf("failed to create new node: %v", err)
	}

	if id := n.ID(); id != testID {
		t.Errorf("expected ID: %d, got: %d", testID, id)
	}

	if l := n.Label(); l != testLabel {
		t.Errorf("expected label: %s, got: %s", testLabel, l)
	}

	newLabel := "newLabel"
	n.SetLabel(newLabel)
	if l := n.Label(); l != newLabel {
		t.Errorf("expected label: %s, got: %s", newLabel, l)
	}

	if a := n.Attrs(); !reflect.DeepEqual(a, attrs) {
		t.Errorf("expected attrs: %v, got: %v", attrs, a)
	}

	if s := n.Type(); s != style.DefaultStyleType {
		t.Errorf("expected type: %s, got: %s", style.DefaultStyleType, s)
	}

	if s := n.Shape(); s != style.DefaultNodeShape {
		t.Errorf("expected shape: %s, got: %s", style.DefaultNodeShape, s)
	}

	if c := n.Color(); c != style.DefaultNodeColor {
		t.Errorf("expected color: %v, got: %v", style.DefaultNodeColor, c)
	}

	if d := n.DOTID(); d != n.UID() {
		t.Errorf("expected dotid: %s, got: %s", n.UID(), d)
	}

	n.SetDOTID(testDotID)

	if d := n.DOTID(); d != testDotID {
		t.Errorf("expected dotid: %s, got: %s", testDotID, d)
	}

	newUID := "newUID"
	n.SetUID(newUID)
	if uid := n.UID(); uid != newUID {
		t.Errorf("expected UID: %s, got: %s", newUID, uid)
	}

	if a := n.Attributes(); len(a) != len(attrs) {
		t.Errorf("expected attributes count: %d, got: %d", len(attrs), len(a))
	}
}

func TestNewNodeWithOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	attrs := map[string]interface{}{
		"foo": "bar",
	}

	style := style.Style{
		Type:  "foo",
		Shape: "bar",
		Color: style.DefaultNodeColor,
	}

	opts := []Option{
		WithLabel(testLabel),
		WithAttrs(attrs),
		WithUID(testUID),
		WithDotID(testDotID),
		WithStyle(style),
	}

	n, err := NewNode(testID, opts...)
	if err != nil {
		t.Fatalf("failed to create new node: %v", err)
	}

	if s := n.Type(); s != style.Type {
		t.Errorf("expected style: %s, got: %s", style.Type, s)
	}

	if s := n.Shape(); s != style.Shape {
		t.Errorf("expected shape: %s, got: %s", style.Shape, s)
	}

	if c := n.Color(); c != style.Color {
		t.Errorf("expected color: %v, got: %v", style.Color, c)
	}

	if d := n.DOTID(); d != testDotID {
		t.Errorf("expected dotid: %s, got: %s", testDotID, d)
	}

	if u := n.UID(); u != testUID {
		t.Errorf("expected uid: %s, got: %s", testUID, u)
	}
}
