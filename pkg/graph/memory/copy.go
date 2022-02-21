package memory

import (
	"sync"

	"github.com/milosgajdos/orbnet/pkg/graph/attrs"
	"gonum.org/v1/gonum/graph/simple"
)

// NodeDeepCopy makes a deep copy of Node and returns it.
func NodeDeepCopy(n *Node) *Node {
	return &Node{
		id:    n.id,
		uid:   n.uid,
		dotid: n.dotid,
		label: n.label,
		attrs: attrs.CopyFrom(n.attrs),
		style: n.style,
	}
}

// EdgeDeepCopy makes a deep copy of Edge and returns it
func EdgeDeepCopy(e *Edge) *Edge {
	return &Edge{
		uid:    e.uid,
		from:   NodeDeepCopy(e.From().(*Node)),
		to:     NodeDeepCopy(e.To().(*Node)),
		weight: e.weight,
		label:  e.label,
		attrs:  attrs.CopyFrom(e.attrs),
		style:  e.style,
	}
}

// GraphDeepCopy return s deep copy of a memory graph.
func GraphDeepCopy(g *Graph) *Graph {
	g.mu.RLock()
	defer g.mu.RUnlock()

	cg := &Graph{
		WeightedDirectedGraph: simple.NewWeightedDirectedGraph(DefaultWeight, 0.0),
		uid:                   g.uid,
		typ:                   g.typ,
		label:                 g.label,
		attrs:                 attrs.CopyFrom(g.attrs),
		mu:                    &sync.RWMutex{},
	}

	// copy all src nodes.
	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node().(*Node)
		node := NodeDeepCopy(n)
		cg.AddNode(node)
	}

	// copy all src edges.
	nodes.Reset()
	for nodes.Next() {
		nid := nodes.Node().ID()
		to := g.From(nid)
		for to.Next() {
			vid := to.Node().ID()
			e := g.WeightedEdge(nid, vid).(*Edge)
			edge := EdgeDeepCopy(e)
			cg.SetWeightedEdge(edge)
		}
	}

	return cg
}
