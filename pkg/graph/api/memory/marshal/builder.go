package marshal

import (
	"context"
	"sync"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/api"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"

	gonum "gonum.org/v1/gonum/graph/simple"
)

// Builder is builds in-memory API graph.
type Builder struct {
	nodes map[string]*memory.Node
	mu    *sync.RWMutex
}

// NewBuilder creates a new builder and returns it.
func NewBuilder() (*Builder, error) {
	return &Builder{
		nodes: make(map[string]*memory.Node),
		mu:    &sync.RWMutex{},
	}, nil
}

// addNodes adds nodes to graph g.
func (b *Builder) addNodes(_ context.Context, g graph.Adder, nodes []api.Node) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, node := range nodes {
		n := gonum.Node(node.ID)

		var opts []memory.Option
		if uid := node.UID; uid != "" {
			opts = append(opts, memory.WithUID(uid))
		}

		if label := node.Label; *label != "" {
			opts = append(opts, memory.WithLabel(*label))
		}

		if attrs := node.Attrs; attrs != nil {
			opts = append(opts, memory.WithAttrs(attrs))
		}

		memNode, err := memory.NewNode(n.ID(), opts...)
		if err != nil {
			return err
		}

		if _, ok := b.nodes[memNode.UID()]; !ok {
			g.AddNode(memNode)
			b.nodes[memNode.UID()] = memNode
		}
	}

	return nil
}

// addEdges adds edges to graph g.
func (b *Builder) addEdges(_ context.Context, g graph.Adder, edges []api.Edge) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, edge := range edges {
		from := g.Node(edge.Source)
		if from == nil {
			return graph.Errorf(graph.EINTERNAL, "edge adder: could not find source node: %d", edge.Source)
		}

		to := g.Node(edge.Target)
		if to == nil {
			return graph.Errorf(graph.EINTERNAL, "edge adder: could not find target node: %d", edge.Target)
		}

		// dont create a new edge if an edge already exists
		if e := g.Edge(from.ID(), to.ID()); e != nil {
			continue
		}

		var opts []memory.Option
		if uid := edge.UID; uid != "" {
			opts = append(opts, memory.WithUID(uid))
		}

		if label := edge.Label; label != "" {
			opts = append(opts, memory.WithLabel(label))
		}

		if attrs := edge.Attrs; attrs != nil {
			opts = append(opts, memory.WithAttrs(attrs))
		}

		e, err := memory.NewEdge(from, to, opts...)
		if err != nil {
			return err
		}

		g.SetWeightedEdge(e)
	}

	return nil
}

// Build builds a graph by adding nodes and edges from the data received on the given channel
func (b *Builder) Build(ctx context.Context, g graph.Adder, dataChan <-chan interface{}) error {
	for {
		select {
		case data := <-dataChan:
			// dataChan has been closed
			if data == nil {
				return nil
			}

			if d, ok := data.(Graph); ok {
				if l, ok := g.(graph.LabelSetter); ok {
					l.SetLabel(*d.Graph.Label)
				}

				if u, ok := g.(graph.UIDSetter); ok {
					u.SetUID(d.Graph.UID)
				}

				for k, v := range d.Graph.Attrs {
					g.Attrs()[k] = v
				}

				if err := b.addNodes(ctx, g, d.Nodes); err != nil {
					return err
				}

				if err := b.addEdges(ctx, g, d.Edges); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}
