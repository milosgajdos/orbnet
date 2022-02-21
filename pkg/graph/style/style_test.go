package style

import (
	"reflect"
	"testing"
)

func TestDefaultNode(t *testing.T) {
	dn := DefaultNode()

	s := Style{
		Type:  DefaultStyleType,
		Shape: DefaultNodeShape,
		Color: DefaultNodeColor,
	}

	if !reflect.DeepEqual(dn, s) {
		t.Fatal("unexpected default node style")
	}
}

func TestDefaultEdge(t *testing.T) {
	de := DefaultEdge()

	s := Style{
		Type:  DefaultStyleType,
		Shape: DefaultEdgeShape,
		Color: DefaultEdgeColor,
	}

	if !reflect.DeepEqual(de, s) {
		t.Fatal("unexpected default edge style")
	}
}
