package gexf

import (
	"bytes"
	"encoding/xml"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"gonum.org/v1/gonum/graph/formats/gexf12"
)

const (
	// nameAttr is used as a key for node name.
	nameAttr = "name"
	// relAttr stores the name of relation attribute
	relAttr = "relation"
)

// Marshaler is used for marshaling graphs.
type Marshaler struct {
	name   string
	prefix string
	indent string
}

// NewMarshaler creates a new graph marshaler and returns it.
func NewMarshaler(name, prefix, indent string) (*Marshaler, error) {
	return &Marshaler{
		name:   name,
		prefix: prefix,
		indent: indent,
	}, nil
}

// Marshal marshals g into format that can be used by
// Gephi https://gephi.org/
// To learn more about Gexf see here: https://gephi.org/gexf/format/
func (m *Marshaler) Marshal(g graph.Graph) ([]byte, error) {
	c := gexf12.Content{
		Graph: gexf12.Graph{
			TimeFormat:      "dateTime",
			DefaultEdgeType: "directed",
			Mode:            "dynamic",
			Attributes: []gexf12.Attributes{
				{
					Class: "edge",
					Mode:  "dynamic",
					Attributes: []gexf12.Attribute{{
						ID:    "relation",
						Title: "relation",
						Type:  "string",
					}}},
				{
					Class: "node",
					Mode:  "dynamic",
					Attributes: []gexf12.Attribute{{
						ID:    "name",
						Title: "name",
						Type:  "string",
					}}},
			},
		},
		Version: "1.2",
	}

	nodes := g.Nodes()
	c.Graph.Nodes.Count = nodes.Len()
	c.Graph.Nodes.Nodes = make([]gexf12.Node, 0, nodes.Len())
	for nodes.Next() {
		n := NewNode(nodes.Node().(graph.Node))
		c.Graph.Nodes.Nodes = append(c.Graph.Nodes.Nodes, *n)
	}

	edges := g.Edges()
	i := 0
	for edges.Next() {
		e := NewEdge(i, edges.Edge().(graph.Edge))
		c.Graph.Edges.Edges = append(c.Graph.Edges.Edges, *e)
		i++
	}
	c.Graph.Edges.Count = len(c.Graph.Edges.Edges)

	var b bytes.Buffer
	enc := xml.NewEncoder(&b)
	enc.Indent(m.prefix, m.indent)

	if err := enc.Encode(c); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
