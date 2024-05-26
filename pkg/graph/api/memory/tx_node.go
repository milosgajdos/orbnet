package memory

import (
	"context"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"

	gonum "gonum.org/v1/gonum/graph"
)

// TxNode wraps memory.Node
type TxNode struct {
	*memory.Node
	DegIn  int
	DegOut int
}

// CreateNode adds a new node to graph.
// nolint:revive
func (t *Tx) CreateNode(ctx context.Context, uid string, n *api.Node) error {
	t.db.Lock()
	defer t.db.Unlock()

	g, ok := t.db.db[uid]
	if !ok {
		return api.Errorf(api.ENOTFOUND, "graph %s not found", uid)
	}

	var opts []memory.Option

	if *n.Label != "" {
		opts = append(opts, memory.WithLabel(*n.Label))
	}

	if n.Attrs != nil {
		opts = append(opts, memory.WithAttrs(n.Attrs))
	}

	gnode := g.NewNode()
	node, err := memory.NewNode(gnode.ID(), opts...)
	if err != nil {
		return err
	}

	g.AddNode(node)

	n.ID = node.ID()
	n.UID = node.UID()
	n.Label = StringPtr(node.Label())
	n.Attrs = node.Attrs()
	n.DegIn = 0
	n.DegOut = 0

	return nil
}

// findNodeByID is a helper function that finds a node with the given id or it returns error.
// nolint:revive
func (t Tx) findNodeByID(ctx context.Context, g *memory.Graph, id int64) (*TxNode, error) {
	n := g.Node(id)
	if n == nil {
		return nil, api.Errorf(api.ENOTFOUND, "node %d not found", id)
	}

	node := n.(*memory.Node)

	return &TxNode{
		Node:   memory.NodeDeepCopy(node),
		DegOut: g.From(node.ID()).Len(),
		DegIn:  g.To(node.ID()).Len(),
	}, nil
}

// FindNodeByID returns node with the given ID.
// It returns error if the node with the given id could not be found.
func (t *Tx) FindNodeByID(ctx context.Context, uid string, id int64) (*TxNode, error) {
	t.db.RLock()
	defer t.db.RUnlock()

	g, ok := t.db.db[uid]
	if !ok {
		return nil, api.Errorf(api.ENOTFOUND, "graph %s not found", uid)
	}

	return t.findNodeByID(ctx, g, id)
}

// findNodeByUID is a helper function that finds a node with the given uid or it returns error.
// nolint:revive
func (t Tx) findNodeByUID(ctx context.Context, g *memory.Graph, uid string) (*TxNode, error) {
	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node().(*memory.Node)
		if n.UID() == uid {
			return &TxNode{
				Node:   memory.NodeDeepCopy(n),
				DegOut: g.From(n.ID()).Len(),
				DegIn:  g.To(n.ID()).Len(),
			}, nil
		}
	}

	return nil, api.Errorf(api.ENOTFOUND, "node %s not found", uid)
}

// FindNodeByUID returns node with the given UID.
// It returns error if the node with the given uid could not be found.
func (t *Tx) FindNodeByUID(ctx context.Context, guid, uid string) (*TxNode, error) {
	t.db.RLock()
	defer t.db.RUnlock()

	g, ok := t.db.db[guid]
	if !ok {
		return nil, api.Errorf(api.ENOTFOUND, "graph %s not found", guid)
	}

	return t.findNodeByUID(ctx, g, uid)
}

// nolint:revive
func filterNodes(ctx context.Context, nodes gonum.Nodes, g *memory.Graph, filter api.NodeFilter) ([]*TxNode, error) {
	var nx []*TxNode
	for nodes.Next() {
		n := nodes.Node().(*memory.Node)
		if l := filter.Label; l != nil {
			if n.Label() == *l {
				txNode := &TxNode{
					Node:   memory.NodeDeepCopy(n),
					DegOut: g.From(n.ID()).Len(),
					DegIn:  g.To(n.ID()).Len(),
				}
				nx = append(nx, txNode)
			}
			continue
		}
		txNode := &TxNode{
			Node:   memory.NodeDeepCopy(n),
			DegOut: g.From(n.ID()).Len(),
			DegIn:  g.To(n.ID()).Len(),
		}
		nx = append(nx, txNode)
	}
	return nx, nil
}

// nolint:revive
func labeledNodes(ctx context.Context, g *memory.Graph, label string, node *TxNode) ([]*TxNode, int, error) {
	if node.Label() == label {
		txNode := &TxNode{
			Node:   memory.NodeDeepCopy(node.Node),
			DegOut: g.From(node.ID()).Len(),
			DegIn:  g.To(node.ID()).Len(),
		}
		// Matches ID and Label
		return []*TxNode{txNode}, 1, nil
	}
	// Does not match the label
	return []*TxNode{}, 0, nil
}

// FindNodes returns all graph nodes matching the filter.
func (t *Tx) FindNodes(ctx context.Context, uid string, filter api.NodeFilter) ([]*TxNode, int, error) {
	t.db.RLock()
	defer t.db.RUnlock()

	g, ok := t.db.db[uid]
	if !ok {
		return nil, 0, api.Errorf(api.ENOTFOUND, "graph %s not found", uid)
	}

	// ID has been provided
	if id := filter.ID; id != nil {
		node, err := t.findNodeByID(ctx, g, *id)
		if err != nil {
			if code := api.ErrorCode(err); code == api.ENOTFOUND {
				return []*TxNode{}, 0, nil
			}
			return nil, 0, err
		}
		if l := filter.Label; l != nil {
			return labeledNodes(ctx, g, *l, node)
		}
		// Matches ID, but no Label has been provided
		return []*TxNode{node}, 1, nil
	}

	// ID has been provided
	if uid := filter.UID; uid != nil {
		node, err := t.findNodeByUID(ctx, g, *uid)
		if err != nil {
			if code := api.ErrorCode(err); code == api.ENOTFOUND {
				return []*TxNode{}, 0, nil
			}
			return nil, 0, err
		}
		if l := filter.Label; l != nil {
			return labeledNodes(ctx, g, *l, node)
		}
		// Matches ID, but no Label has been provided
		return []*TxNode{node}, 1, nil
	}

	var nodes []*TxNode
	var err error

	// To has been provided
	if to := filter.Target; to != nil {
		// TODO: make this more efficient
		// This iterates over g.Nodes searching for to.
		// We then iterate over all its incoming nodes.
		node, err := t.findNodeByUID(ctx, g, *to)
		if err != nil {
			if code := api.ErrorCode(err); code == api.ENOTFOUND {
				return []*TxNode{}, 0, nil
			}
			return nil, 0, err
		}
		nodes, err = filterNodes(ctx, g.To(node.ID()), g, filter)
		if err != nil {
			return nil, 0, err
		}
		return applyOffsetLimit(nodes, filter.Offset, filter.Limit).([]*TxNode), len(nodes), nil
	}

	// From has been provided
	if from := filter.Source; from != nil {
		// TODO: make this more efficient
		// This iterates over g.Nodes searching for from
		// We then iterate over all its outgoing nodes.
		node, err := t.findNodeByUID(ctx, g, *from)
		if err != nil {
			if code := api.ErrorCode(err); code == api.ENOTFOUND {
				return []*TxNode{}, 0, nil
			}
			return nil, 0, err
		}
		nodes, err = filterNodes(ctx, g.From(node.ID()), g, filter)
		if err != nil {
			return nil, 0, err
		}
		return applyOffsetLimit(nodes, filter.Offset, filter.Limit).([]*TxNode), len(nodes), nil
	}

	nodes, err = filterNodes(ctx, g.Nodes(), g, filter)
	if err != nil {
		return nil, 0, err
	}
	return applyOffsetLimit(nodes, filter.Offset, filter.Limit).([]*TxNode), len(nodes), nil
}

// UpdateNode updates node.
// nolint:revive
func (t *Tx) UpdateNode(ctx context.Context, uid string, id int64, update api.NodeUpdate) (*TxNode, error) {
	t.db.Lock()
	defer t.db.Unlock()

	g, ok := t.db.db[uid]
	if !ok {
		return nil, api.Errorf(api.ENOTFOUND, "graph %s not found", uid)
	}

	n := g.Node(id)
	if n == nil {
		return nil, api.Errorf(api.ENOTFOUND, "node %d not found in graph %s", id, uid)
	}

	node, ok := n.(*memory.Node)
	if !ok {
		return nil, api.Errorf(api.EINTERNAL, "invalid node data found")
	}

	if l := update.Label; l != nil {
		node.SetLabel(*l)
	}

	if a := update.Attrs; a != nil {
		for k, v := range a {
			node.Attrs()[k] = v
		}
	}

	return &TxNode{
		Node:   memory.NodeDeepCopy(node),
		DegOut: g.From(n.ID()).Len(),
		DegIn:  g.To(n.ID()).Len(),
	}, nil
}

// DeleteNode deletes node from graph.
// nolint:revive
func (t *Tx) DeleteNodeByID(ctx context.Context, uid string, id int64) error {
	t.db.Lock()
	defer t.db.Unlock()

	g, ok := t.db.db[uid]
	if !ok {
		return api.Errorf(api.ENOTFOUND, "graph %s not found", uid)
	}

	g.RemoveNode(id)

	return nil
}

// DeleteNode deletes node from graph.
func (t *Tx) DeleteNodeByUID(ctx context.Context, guid, nuid string) error {
	t.db.Lock()
	defer t.db.Unlock()

	g, ok := t.db.db[guid]
	if !ok {
		return api.Errorf(api.ENOTFOUND, "graph %s not found", guid)
	}

	n, err := t.findNodeByUID(ctx, g, nuid)
	if err != nil {
		return err
	}

	g.RemoveNode(n.ID())

	return nil
}
