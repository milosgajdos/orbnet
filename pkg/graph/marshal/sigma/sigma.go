package sigma

import (
	"encoding/json"
	"fmt"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"gonum.org/v1/gonum/graph/formats/sigmajs"
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

// Marshal marshals g into format that can be used by
// SigmaJS. See here for more: http://sigmajs.org/
func (m *Marshaler) Marshal(g graph.Graph) ([]byte, error) {
	c := sigmajs.Graph{
		Nodes: make([]sigmajs.Node, g.Nodes().Len()),
		Edges: make([]sigmajs.Edge, g.Edges().Len()),
	}

	nodes := g.Nodes()
	i := 0
	for nodes.Next() {
		n := nodes.Node().(graph.Node)

		c.Nodes[i] = sigmajs.Node{
			ID:         fmt.Sprint(n.ID()),
			Attributes: n.Attrs(),
		}

		i++
	}

	edges := g.Edges()
	i = 0
	for edges.Next() {
		e := edges.Edge().(graph.Edge)

		c.Edges[i] = sigmajs.Edge{
			ID:         fmt.Sprint(i),
			Source:     fmt.Sprint(e.From().ID()),
			Target:     fmt.Sprint(e.To().ID()),
			Attributes: e.Attrs(),
		}

		i++
	}

	return json.MarshalIndent(c, m.prefix, m.indent)
}
