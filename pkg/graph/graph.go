package graph

import (
	"context"
	"image/color"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
)

// Graph is weighted graph.
type Graph interface {
	graph.Weighted
	// UID returns graph UID.
	UID() string
	// Edges returns graph edges iterator.
	Edges() graph.Edges
	// Label returns graph label.
	Label() string
	// Attrs are graph attributes.
	Attrs() map[string]interface{}
}

// Node is a graph node.
type Node interface {
	graph.Node
	// UID returns node UID.
	UID() string
	// Label returns node label.
	Label() string
	// Attrs returns node attributes.
	Attrs() map[string]interface{}
}

// Styler is used for styling.
type Styler interface {
	// Type returns the type of style.
	Type() string
	// Shape returns style shape.
	Shape() string
	// Color returns style color.
	Color() color.RGBA
}

// DOTNode is Graphviz DOT node.
type DOTNode interface {
	Node
	encoding.Attributer
	// DOTID returns DOT ID.
	DOTID() string
	// SetDOTID sets DOT ID.
	SetDOTID(dotid string)
}

// Edge is a graph edge.
type Edge interface {
	graph.WeightedEdge
	// UID returns edge UID.
	UID() string
	// Label returns edge label.
	Label() string
	// Attrs returns node attributes.
	Attrs() map[string]interface{}
}

// LabelSetter sets label.
type LabelSetter interface {
	SetLabel(string)
}

// UIDSetter sets UID.
type UIDSetter interface {
	SetUID(string)
}

// WeightSetter sets weight.
type WeightSetter interface {
	SetWeight(float64)
}

// DOTEdge is Graphviz DOT edge.
type DOTEdge interface {
	Edge
	encoding.Attributer
}

// Adder allows to add edges and nodes to graph.
type Adder interface {
	Graph
	graph.NodeAdder
	graph.WeightedEdgeAdder
}

// Remover allows to remove nodes and edges from graph.
type Remover interface {
	Graph
	graph.NodeRemover
	graph.EdgeRemover
}

// Updater allows to update graph.
type Updater interface {
	Adder
	Remover
}

// NodeUpdater adds and removes nodes.
type NodeUpdater interface {
	Graph
	graph.NodeAdder
	graph.NodeRemover
}

// EdgeUpdater adds and removes edges.
type EdgeUpdater interface {
	Graph
	graph.WeightedEdgeAdder
	graph.EdgeRemover
}

// Marshaler is used for marshaling graphs.
type Marshaler interface {
	// Marshal marshals graph into bytes.
	Marshal(g Graph) ([]byte, error)
}

// Unmarshaler is used for unmarshaling graphs.
type Unmarshaler interface {
	// Unmarshal unmarshals arbitrary bytes into graph.
	Unmarshal([]byte, Graph) error
}

// Syncer syncs the graph to a database or a filesystem.
type Syncer interface {
	Sync(context.Context, Graph) error
}

// Loader loads a graph from a databse or a filesystem.
type Loader interface {
	Load(context.Context, string) (Graph, error)
}
