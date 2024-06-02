package memory

import (
	"fmt"
	"sync"

	"github.com/google/uuid"

	gonum "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

const (
	// DefaultLabel is memory graph default label.
	DefaultLabel = "InMemoryGraph"
	// DefaultType is default graph type.
	DefaultType = "weighted_directed"
	// DefaultWeight is default edge weight.
	DefaultWeight = 1.0
)

// Graph is an in-memory graph.
type Graph struct {
	*simple.WeightedDirectedGraph
	uid   string
	typ   string
	label string
	attrs map[string]interface{}
	mu    *sync.RWMutex
}

// NewGraph creates a new graph and returns it.
func NewGraph(opts ...Option) (*Graph, error) {
	gopts := Options{
		UID:   uuid.New().String(),
		Label: DefaultLabel,
		Attrs: make(map[string]interface{}),
		Type:  DefaultType,
	}

	for _, apply := range opts {
		apply(&gopts)
	}

	if gopts.Type != DefaultType {
		return nil, fmt.Errorf("unsupported graph type: %s", gopts.Type)
	}

	return &Graph{
		WeightedDirectedGraph: simple.NewWeightedDirectedGraph(DefaultWeight, 0.0),
		uid:                   gopts.UID,
		typ:                   gopts.Type,
		label:                 gopts.Label,
		attrs:                 gopts.Attrs,
		mu:                    &sync.RWMutex{},
	}, nil
}

// UID returns graph UID.
func (g Graph) UID() string {
	return g.uid
}

// Type returns the type of graph.
func (g Graph) Type() string {
	return g.typ
}

// Label returns graph label.
func (g Graph) Label() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.label
}

// SetLabel sets label.
func (g *Graph) SetLabel(l string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.label = l
}

// SetUID sets UID.
func (g *Graph) SetUID(uid string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.uid = uid
}

// Attrs returns graph attributes.
func (g *Graph) Attrs() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.attrs
}

// HasEdgeFromTo returns whether an edge exist between two nodoes with the given IDs.
func (g Graph) HasEdgeFromTo(uid, vid int64) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.WeightedDirectedGraph.HasEdgeBetween(uid, vid)
}

// To returns all nodes that can reach directly to the node with the given ID.
func (g Graph) To(id int64) gonum.Nodes {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.WeightedDirectedGraph.To(id)
}
