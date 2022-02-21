package cytoscape

import (
	"encoding/json"
	"fmt"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"gonum.org/v1/gonum/graph/formats/cytoscapejs"
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
// CytoscapJS https://js.cytoscape.org/
func (m *Marshaler) Marshal(g graph.Graph) ([]byte, error) {
	c := cytoscapejs.Elements{
		Nodes: make([]cytoscapejs.Node, g.Nodes().Len()),
		Edges: make([]cytoscapejs.Edge, g.Edges().Len()),
	}

	nodes := g.Nodes()
	i := 0
	for nodes.Next() {
		n := nodes.Node().(graph.Node)

		ndata := cytoscapejs.NodeData{
			ID:         fmt.Sprint(n.ID()),
			Attributes: n.Attrs(),
		}

		c.Nodes[i] = cytoscapejs.Node{
			Data:       ndata,
			Selectable: true,
		}

		i++
	}

	edges := g.Edges()
	i = 0
	for edges.Next() {
		e := edges.Edge().(graph.Edge)

		edata := cytoscapejs.EdgeData{
			ID:         fmt.Sprint(i),
			Source:     fmt.Sprint(e.From().ID()),
			Target:     fmt.Sprint(e.To().ID()),
			Attributes: e.Attrs(),
		}

		c.Edges[i] = cytoscapejs.Edge{
			Data:       edata,
			Selectable: true,
		}

		i++
	}

	return json.MarshalIndent(c, m.prefix, m.indent)
}
