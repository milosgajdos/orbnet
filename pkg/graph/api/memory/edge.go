package memory

import (
	"context"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

// EdgeService lets you manage graph edges.
type EdgeService struct {
	db *DB
}

// NewEdgeService creates an instance of EdgeService and returns it.
// Nodes managed by the node service belong to the graph with the given uid.
func NewEdgeService(db *DB) (*EdgeService, error) {
	return &EdgeService{
		db: db,
	}, nil
}

// CreateEdge creates a new edge.
func (es *EdgeService) CreateEdge(ctx context.Context, uid string, e *api.Edge) error {
	tx, err := es.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	return tx.CreateEdge(ctx, uid, e)
}

// FindEdgeByUID returns a single edge with the given id.
func (es *EdgeService) FindEdgeByUID(ctx context.Context, guid, euid string) (*api.Edge, error) {
	tx, err := es.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	edge, err := tx.FindEdgeByUID(ctx, guid, euid)
	if err != nil {
		return nil, err
	}

	return &api.Edge{
		UID:    edge.UID(),
		Source: edge.From().ID(),
		Target: edge.To().ID(),
		Weight: edge.Weight(),
		Label:  edge.Label(),
		Attrs:  edge.Attrs(),
	}, nil
}

// FindEdges returns all edges matching the filter.
// It also returns a count of total matching edges which may differ from
// the number of returned edges if the Limit field is set.
func (es *EdgeService) FindEdges(ctx context.Context, guid string, filter api.EdgeFilter) ([]*api.Edge, int, error) {
	tx, err := es.db.BeginTx(ctx)
	if err != nil {
		return nil, 0, err
	}

	ex, count, err := tx.FindEdges(ctx, guid, filter)
	if err != nil {
		return nil, count, err
	}

	edges := make([]*api.Edge, len(ex))

	for i, e := range ex {
		edges[i] = &api.Edge{
			UID:   e.UID(),
			Label: e.Label(),
			Attrs: e.Attrs(),
		}
	}

	return edges, count, nil
}

// UpdateEdgeBetween updates an edge between two nodes.
func (es *EdgeService) UpdateEdgeBetween(ctx context.Context, guid string, source, target int64, update api.EdgeUpdate) (*api.Edge, error) {
	tx, err := es.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	edge, err := tx.UpdateEdgeBetween(ctx, guid, source, target, update)
	if err != nil {
		return nil, err
	}

	return &api.Edge{
		UID:    edge.UID(),
		Source: edge.From().ID(),
		Target: edge.To().ID(),
		Weight: edge.Weight(),
		Label:  edge.Label(),
		Attrs:  edge.Attrs(),
	}, nil
}

// DeleteEdge permanently removes an edge by UID.
func (es *EdgeService) DeleteEdge(ctx context.Context, guid, euid string) error {
	tx, err := es.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	return tx.DeleteEdge(ctx, guid, euid)
}

// DeleteEdgeBetween permanently deletes all edges between source and target nodes.
func (es *EdgeService) DeleteEdgeBetween(ctx context.Context, guid string, source, target int64) error {
	tx, err := es.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	return tx.DeleteEdgeBetween(ctx, guid, source, target)
}
