package gexf

import (
	"fmt"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/attrs"
	"gonum.org/v1/gonum/graph/formats/gexf12"
)

const (
	DefaultRelation = "Undefined"
)

func NewNode(n graph.Node) *gexf12.Node {
	node := &gexf12.Node{
		ID:    fmt.Sprint(n.ID()),
		Label: n.Label(),
	}

	sn, ok := n.(graph.Styler)
	if ok {
		node.Color = &gexf12.Color{
			R: sn.Color().R,
			G: sn.Color().G,
			B: sn.Color().B,
		}
	}

	a := n.Attrs()

	if d := attrs.ToString("date", a["date"]); d != "" {
		node.Start = d
	}

	name := n.UID()

	if n := attrs.ToString("name", a["name"]); n != "" {
		name = n
	}

	att := gexf12.AttValue{
		For:   nameAttr,
		Value: name,
	}

	node.AttValues = &gexf12.AttValues{AttValues: []gexf12.AttValue{att}}

	return node
}

func NewEdge(id int, e graph.Edge) *gexf12.Edge {
	edge := &gexf12.Edge{
		ID:     fmt.Sprint(id),
		Source: fmt.Sprint(e.From().ID()),
		Target: fmt.Sprint(e.To().ID()),
	}

	a := e.Attrs()

	relation := DefaultRelation

	if r := attrs.ToString("relation", a["relation"]); r != "" {
		edge.Label = r
		relation = r
	}

	if d := attrs.ToString("date", a["date"]); d != "" {
		edge.Start = d
	}

	if relation != "" {
		att := gexf12.AttValue{
			For:   relAttr,
			Value: relation,
		}
		edge.AttValues = &gexf12.AttValues{AttValues: []gexf12.AttValue{att}}
	}

	return edge
}
