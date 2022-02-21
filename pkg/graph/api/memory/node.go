package memory

import (
	"context"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

// NodeService manages graph nodes.
type NodeService struct {
	db *DB
}

// NewNodeService creates an instance of NodeService and returns it.
// Nodes managed by the node service belong to the graph with the given uid.
func NewNodeService(db *DB) (*NodeService, error) {
	return &NodeService{
		db: db,
	}, nil
}

// CreateNode creates a new node.
func (ns *NodeService) CreateNode(ctx context.Context, uid string, n *api.Node) error {
	tx, err := ns.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	return tx.CreateNode(ctx, uid, n)
}

// FindNodeByID returns a single node with the given id.
func (ns *NodeService) FindNodeByID(ctx context.Context, uid string, id int64) (*api.Node, error) {
	tx, err := ns.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	node, err := tx.FindNodeByID(ctx, uid, id)
	if err != nil {
		return nil, err
	}

	return &api.Node{
		ID:     node.ID(),
		UID:    node.UID(),
		Label:  node.Label(),
		Attrs:  node.Attrs(),
		DegOut: node.DegOut,
		DegIn:  node.DegIn,
	}, nil
}

// FindNodeByID returns a single node with the given uid.
func (ns *NodeService) FindNodeByUID(ctx context.Context, guid, nuid string) (*api.Node, error) {
	tx, err := ns.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	node, err := tx.FindNodeByUID(ctx, guid, nuid)
	if err != nil {
		return nil, err
	}

	return &api.Node{
		ID:     node.ID(),
		UID:    node.UID(),
		Label:  node.Label(),
		Attrs:  node.Attrs(),
		DegOut: node.DegOut,
		DegIn:  node.DegIn,
	}, nil
}

// FindNodes returns all nodes matching the filter.
// It also returns a count of total matching nodes which may differ from
// the number of returned nodes if the Limit field is set.
func (ns *NodeService) FindNodes(ctx context.Context, uid string, filter api.NodeFilter) ([]*api.Node, int, error) {
	tx, err := ns.db.BeginTx(ctx)
	if err != nil {
		return nil, 0, err
	}

	nx, count, err := tx.FindNodes(ctx, uid, filter)
	if err != nil {
		return nil, count, err
	}

	nodes := make([]*api.Node, len(nx))

	for i, n := range nx {
		nodes[i] = &api.Node{
			ID:     n.ID(),
			UID:    n.UID(),
			Label:  n.Label(),
			Attrs:  n.Attrs(),
			DegOut: n.DegOut,
			DegIn:  n.DegIn,
		}
	}

	return nodes, count, nil
}

// UpdateNode updates an existing node by ID.
func (ns *NodeService) UpdateNode(ctx context.Context, uid string, id int64, update api.NodeUpdate) (*api.Node, error) {
	tx, err := ns.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	node, err := tx.UpdateNode(ctx, uid, id, update)
	if err != nil {
		return nil, err
	}

	return &api.Node{
		ID:     node.ID(),
		UID:    node.UID(),
		Label:  node.Label(),
		Attrs:  node.Attrs(),
		DegOut: node.DegOut,
		DegIn:  node.DegIn,
	}, nil
}

// DeleteNodeByID permanently removes a node by ID.
// It automatically removes removed node's incoming and outgoing edges.
func (ns *NodeService) DeleteNodeByID(ctx context.Context, uid string, id int64) error {
	tx, err := ns.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	return tx.DeleteNodeByID(ctx, uid, id)
}

// DeleteNodeByUID permanently removes a node by UID.
// It automatically removes removed node's incoming and outgoing edges.
func (ns *NodeService) DeleteNodeByUID(ctx context.Context, guid, nuid string) error {
	tx, err := ns.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	return tx.DeleteNodeByUID(ctx, guid, nuid)
}
