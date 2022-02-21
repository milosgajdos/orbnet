package memory

import (
	"reflect"
	"testing"
)

func MustGraph(t *testing.T) *Graph {
	g, err := NewGraph(WithLabel("foo"))
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	n1, err := NewNode(1, WithLabel("node1"))
	if err != nil {
		t.Fatalf("failed to create graph node: %v", err)
	}
	g.AddNode(n1)
	n2, err := NewNode(2, WithLabel("node2"))
	if err != nil {
		t.Fatalf("failed to create graph node: %v", err)
	}
	g.AddNode(n2)
	e, err := NewEdge(n1, n2, WithLabel("node1-node2"))
	if err != nil {
		t.Fatalf("failed to create graph edge: %v", err)
	}
	g.SetWeightedEdge(e)

	return g
}

func TestCopy(t *testing.T) {
	t.Run("EmptyGraph", func(t *testing.T) {
		g, err := NewGraph(WithLabel("foo"))
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}

		g2 := GraphDeepCopy(g)

		if !reflect.DeepEqual(g, g2) {
			t.Fatal("expected graphs to be equal")
		}
	})

	t.Run("Non-EmptyGraph", func(t *testing.T) {
		g := MustGraph(t)
		g2 := GraphDeepCopy(g)

		if !reflect.DeepEqual(g, g2) {
			t.Fatalf("expected graphs to be equal g: %#v, g2: %#v", g, g2)
		}
	})
}
