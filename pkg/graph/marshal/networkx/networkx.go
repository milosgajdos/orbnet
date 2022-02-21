package networkx

import (
	"encoding/json"
	"fmt"

	"github.com/milosgajdos/orbnet/pkg/graph"
)

// Node is Graph node
type Node struct {
	// ID of the node
	ID string
	// Attributes define node attributes
	Attributes map[string]interface{}
}

// Link links source and target nodes
type Link struct {
	// ID of the link
	ID string
	// Source node ID
	Source string
	// Target node ID
	Target string
	// Attributes define link attributes
	Attributes map[string]interface{}
}

// Graph stores Nodes and Links.
type Graph struct {
	// Nodes are graph nodes
	Nodes []Node `json:"nodes"`
	// Links connect nodes
	Links []Link `json:"links"`
}

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
// networkx frameworkk https://networkx.org/
func (m *Marshaler) Marshal(g graph.Graph) ([]byte, error) {
	c := Graph{
		Nodes: make([]Node, g.Nodes().Len()),
		Links: make([]Link, g.Edges().Len()),
	}

	nodes := g.Nodes()
	i := 0
	for nodes.Next() {
		n := nodes.Node().(graph.Node)

		c.Nodes[i] = Node{
			ID:         fmt.Sprint(n.ID()),
			Attributes: n.Attrs(),
		}

		i++
	}

	edges := g.Edges()
	i = 0
	for edges.Next() {
		e := edges.Edge().(graph.Edge)

		c.Links[i] = Link{
			ID:         fmt.Sprint(i),
			Source:     fmt.Sprint(e.From().ID()),
			Target:     fmt.Sprint(e.To().ID()),
			Attributes: e.Attrs(),
		}

		i++
	}

	return json.MarshalIndent(c, m.prefix, m.indent)
}
