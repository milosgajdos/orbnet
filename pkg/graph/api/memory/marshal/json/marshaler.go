package json

import (
	"encoding/json"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/api"
	"github.com/milosgajdos/orbnet/pkg/graph/api/memory/marshal"
	gonum "gonum.org/v1/gonum/graph"
)

// Marshaler implements graph.Marshaler.
type Marshaler struct {
	name   string
	prefix string
	indent string
}

// NewMarshaler creates a new Marshaler and returns it.
func NewMarshaler(name, prefix, indent string) (*Marshaler, error) {
	return &Marshaler{
		name:   name,
		prefix: prefix,
		indent: indent,
	}, nil
}

// Marshal marshals g into JSON api model.
func (m *Marshaler) Marshal(g graph.Graph) ([]byte, error) {
	ag := &marshal.Graph{
		Graph: api.Graph{
			UID:   g.UID(),
			Nodes: g.Nodes().Len(),
			Edges: g.Edges().Len(),
			Label: StringPtr(g.Label()),
			Attrs: g.Attrs(),
		},
		Nodes: make([]api.Node, g.Nodes().Len()),
		Edges: make([]api.Edge, g.Edges().Len()),
	}

	nodes := g.Nodes()
	i := 0
	for nodes.Next() {
		n := nodes.Node().(graph.Node)

		degOut := g.From(n.ID()).Len()
		degIn := degOut

		dg, ok := g.(gonum.Directed)
		if ok {
			degIn = dg.To(n.ID()).Len()
		}

		ag.Nodes[i] = api.Node{
			ID:     n.ID(),
			UID:    n.UID(),
			DegOut: degOut,
			DegIn:  degIn,
			Label:  n.Label(),
			Attrs:  n.Attrs(),
		}

		i++
	}

	edges := g.Edges()
	i = 0
	for edges.Next() {
		e := edges.Edge().(graph.Edge)

		ag.Edges[i] = api.Edge{
			UID:    e.UID(),
			Source: e.From().ID(),
			Target: e.To().ID(),
			Weight: e.Weight(),
			Label:  e.Label(),
			Attrs:  e.Attrs(),
		}

		i++
	}

	return json.MarshalIndent(ag, m.prefix, m.indent)
}

func StringPtr(s string) *string {
	return &s
}
