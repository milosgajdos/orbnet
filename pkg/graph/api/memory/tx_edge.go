package memory

import (
	"context"
	"fmt"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"

	gonum "gonum.org/v1/gonum/graph"
)

// TxEdge wraps memory.Edge
type TxEdge struct {
	*memory.Edge
}

// CreateEdge adds a new node to graph.
// nolint:revive
func (t *Tx) CreateEdge(ctx context.Context, uid string, e *api.Edge) error {
	t.db.Lock()
	defer t.db.Unlock()

	g, ok := t.db.db[uid]
	if !ok {
		return api.Errorf(api.ENOTFOUND, "graph %s not found", uid)
	}

	if e.Source == e.Target {
		return api.Errorf(api.EINVALID, "self-loops not allowed")
	}

	srcNode, err := findNodeByUID(ctx, g, e.Source)
	if err != nil {
		return api.Errorf(api.ENOTFOUND, "source node %s in graph %s", e.Source, uid)
	}

	source := g.Node(srcNode.ID())
	if source == nil {
		return api.Errorf(api.ENOTFOUND, "source node %s in graph %s", e.Source, uid)
	}

	trgNode, err := findNodeByUID(ctx, g, e.Target)
	if err != nil {
		return api.Errorf(api.ENOTFOUND, "target node: %s in graph: %s", e.Target, g.UID())
	}

	target := g.Node(trgNode.ID())
	if target == nil {
		return api.Errorf(api.ENOTFOUND, "target node %s not found in graph %s", e.Target, uid)
	}

	if e := g.Edge(source.ID(), target.ID()); e != nil {
		return nil
	}

	var opts []memory.Option

	if e.Label != "" {
		opts = append(opts, memory.WithLabel(e.Label))
	}

	if e.Attrs != nil {
		opts = append(opts, memory.WithAttrs(e.Attrs))
	}

	edge, err := memory.NewEdge(source, target, opts...)
	if err != nil {
		return err
	}

	g.SetWeightedEdge(edge)

	e.Label = edge.Label()
	e.Attrs = edge.Attrs()
	e.UID = edge.UID()
	e.Weight = edge.Weight()

	return nil
}

// findEdgeByUID is a helper function that finds a node with the given id ore it returns error.
// nolint:revive
func (t Tx) findEdgeByUID(ctx context.Context, g *memory.Graph, uid string) (*TxEdge, error) {
	edges := g.Edges()
	for edges.Next() {
		e := edges.Edge().(*memory.Edge)
		if e.UID() == uid {
			return &TxEdge{
				Edge: memory.EdgeDeepCopy(e),
			}, nil
		}
	}
	return nil, api.Errorf(api.ENOTFOUND, "edge %s not found", uid)
}

// FindEdgeByUID returns edge with the given UID.
// It returns error if the edge with the given uid could not be found.
func (t *Tx) FindEdgeByUID(ctx context.Context, guid, euid string) (*TxEdge, error) {
	t.db.RLock()
	defer t.db.RUnlock()

	g, ok := t.db.db[guid]
	if !ok {
		return nil, api.Errorf(api.ENOTFOUND, "graph %s not found", guid)
	}
	return t.findEdgeByUID(ctx, g, euid)
}

// nolint:revive
func filterNodeEdges(ctx context.Context, nodes gonum.Nodes, g *memory.Graph, filter api.EdgeFilter) ([]*TxEdge, error) {
	var ex []*TxEdge
	var e gonum.Edge
	var srcID, trgID int64

	if filter.Source != nil {
		srcNode, err := findNodeByUID(ctx, g, *filter.Source)
		if err != nil {
			// TODO: consider returning empty TxEdge slice
			return nil, api.Errorf(api.ENOTFOUND, "source node %s in graph %s", *filter.Source, g.UID())
		}
		srcID = srcNode.ID()
	}

	if filter.Target != nil {
		trgNode, err := findNodeByUID(ctx, g, *filter.Target)
		if err != nil {
			// TODO: consider returning empty TxEdge slice
			return nil, api.Errorf(api.ENOTFOUND, "target node: %s in graph: %s", *filter.Target, g.UID())
		}
		trgID = trgNode.ID()
	}

	for nodes.Next() {
		n := nodes.Node().(*memory.Node)
		if filter.Source != nil {
			e = g.Edge(srcID, n.ID())
		} else {
			e = g.Edge(n.ID(), trgID)
		}

		// NOTE: this should never happen
		if e == nil {
			continue
		}

		memEdge := e.(*memory.Edge)
		if l := filter.Label; l != nil {
			if memEdge.Label() == *l {
				txEdge := &TxEdge{Edge: memEdge}
				ex = append(ex, txEdge)
			}
		}
	}
	return ex, nil
}

// nolint:revive
func filterEdges(ctx context.Context, edges gonum.Edges, filter api.EdgeFilter) ([]*TxEdge, error) {
	var ex []*TxEdge
	for edges.Next() {
		e := edges.Edge().(*memory.Edge)
		if l := filter.Label; l != nil {
			if e.Label() == *l {
				txEdge := &TxEdge{
					Edge: memory.EdgeDeepCopy(e),
				}
				ex = append(ex, txEdge)
			}
			continue
		}
		txEdge := &TxEdge{
			Edge: memory.EdgeDeepCopy(e),
		}
		ex = append(ex, txEdge)
	}
	return ex, nil
}

// FindEdges returns all graph nodes matching the filter.
func (t *Tx) FindEdges(ctx context.Context, uid string, filter api.EdgeFilter) ([]*TxEdge, int, error) {
	t.db.RLock()
	defer t.db.RUnlock()

	g, ok := t.db.db[uid]
	if !ok {
		return nil, 0, api.Errorf(api.ENOTFOUND, "graph %s not found", uid)
	}

	var srcID, trgtID int64

	if filter.Source != nil {
		srcNode, err := findNodeByUID(ctx, g, *filter.Source)
		if err != nil {
			return nil, 0, api.Errorf(api.ENOTFOUND, "source node %s in graph %s", *filter.Source, g.UID())
		}
		srcID = srcNode.ID()
	}

	if filter.Target != nil {
		trgNode, err := findNodeByUID(ctx, g, *filter.Target)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to find target node: %s in graph: %s", *filter.Target, g.UID())
		}
		trgtID = trgNode.ID()
	}

	var edges []*TxEdge
	var err error

	// Both Source and Target have been provided
	// we are looking for a single edge
	if filter.Source != nil && filter.Target != nil {
		e := g.Edge(srcID, trgtID)
		if e == nil {
			return edges, 0, nil
		}

		memEdge := e.(*memory.Edge)
		if l := filter.Label; l != nil {
			if memEdge.Label() == *l {
				txEdge := &TxEdge{Edge: memEdge}
				return []*TxEdge{txEdge}, 1, nil
			}
			return []*TxEdge{}, 0, nil
		}
		txEdge := &TxEdge{Edge: memEdge}
		return []*TxEdge{txEdge}, 1, nil
	}

	// Source has been provided
	if src := filter.Source; src != nil {
		edges, err = filterNodeEdges(ctx, g.From(srcID), g, filter)
		if err != nil {
			return nil, 0, err
		}
		return applyOffsetLimit(edges, filter.Offset, filter.Limit).([]*TxEdge), len(edges), nil
	}

	// Target has been provided
	if target := filter.Target; target != nil {
		edges, err = filterNodeEdges(ctx, g.To(trgtID), g, filter)
		if err != nil {
			return nil, 0, err
		}
		return applyOffsetLimit(edges, filter.Offset, filter.Limit).([]*TxEdge), len(edges), nil
	}

	edges, err = filterEdges(ctx, g.Edges(), filter)
	if err != nil {
		return nil, 0, err
	}
	return applyOffsetLimit(edges, filter.Offset, filter.Limit).([]*TxEdge), len(edges), nil
}

// UpdateEdgeBetween updates edge between two nodes.
// nolint:revive
func (t *Tx) UpdateEdgeBetween(ctx context.Context, uid string, source, target string, update api.EdgeUpdate) (*TxEdge, error) {
	t.db.Lock()
	defer t.db.Unlock()

	g, ok := t.db.db[uid]
	if !ok {
		return nil, api.Errorf(api.ENOTFOUND, "graph %s not found", uid)
	}

	srcNode, err := findNodeByUID(ctx, g, source)
	if err != nil {
		return nil, api.Errorf(api.ENOTFOUND, "source node: %s in graph: %s", source, g.UID())
	}

	trgNode, err := findNodeByUID(ctx, g, target)
	if err != nil {
		return nil, api.Errorf(api.ENOTFOUND, "target node: %s in graph: %s", target, g.UID())
	}

	e := g.Edge(srcNode.ID(), trgNode.ID())
	if e == nil {
		return nil, api.Errorf(api.ENOTFOUND, "edge %s->%s not found", source, target)
	}

	edge, ok := e.(*memory.Edge)
	if !ok {
		return nil, api.Errorf(api.EINTERNAL, "invalid edge data found")
	}

	if l := update.Label; l != nil {
		edge.SetLabel(*l)
	}

	if a := update.Attrs; a != nil {
		for k, v := range a {
			edge.Attrs()[k] = v
		}
	}

	if w := update.Weight; w != nil {
		edge.SetWeight(*w)
	}

	return &TxEdge{
		Edge: memory.EdgeDeepCopy(edge),
	}, nil
}

// DeleteEdge deletes edge with the given uid from graph.
// nolint:revive
func (t *Tx) DeleteEdge(ctx context.Context, guid, euid string) error {
	t.db.Lock()
	defer t.db.Unlock()

	g, ok := t.db.db[guid]
	if !ok {
		return api.Errorf(api.ENOTFOUND, "graph %s not found", guid)
	}

	edges := g.Edges()
	for edges.Next() {
		e := edges.Edge().(*memory.Edge)
		if e.UID() == euid {
			g.RemoveEdge(e.From().ID(), e.To().ID())
			return nil
		}
	}

	return api.Errorf(api.ENOTFOUND, "edge %s not found", euid)
}

// DeleteEdgeBetween deletes all edges between source and target nodes.
// nolint:revive
func (t *Tx) DeleteEdgeBetween(ctx context.Context, uid string, source, target string) error {
	t.db.Lock()
	defer t.db.Unlock()

	g, ok := t.db.db[uid]
	if !ok {
		return api.Errorf(api.ENOTFOUND, "graph %s not found", uid)
	}

	srcNode, err := findNodeByUID(ctx, g, source)
	if err != nil {
		return api.Errorf(api.ENOTFOUND, "source node: %s in graph: %s", source, g.UID())
	}

	trgNode, err := findNodeByUID(ctx, g, target)
	if err != nil {
		return api.Errorf(api.ENOTFOUND, "target node: %s in graph: %s", target, g.UID())
	}

	e := g.Edge(srcNode.ID(), trgNode.ID())
	g.RemoveEdge(e.From().ID(), e.To().ID())

	return nil
}

// nolint:revive
func findNodeByUID(ctx context.Context, g *memory.Graph, uid string) (*memory.Node, error) {
	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node().(*memory.Node)
		if n.UID() == uid {
			return n, nil
		}
	}

	return nil, api.Errorf(api.ENOTFOUND, "node %s not found", uid)
}
