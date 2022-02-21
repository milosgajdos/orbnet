package memory

import (
	"context"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

// GraphService lets you manage graphs.
type GraphService struct {
	db *DB
}

// NewGraphService creates an instance of GraphService and returns it.
func NewGraphService(db *DB) (*GraphService, error) {
	return &GraphService{
		db: db,
	}, nil
}

// CreateGraph creates a new graph.
func (gs *GraphService) CreateGraph(ctx context.Context, g *api.Graph) error {
	tx, err := gs.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	return tx.CreateGraph(ctx, g)
}

// FindGraphByUID returns a single graph with the given uid.
func (gs *GraphService) FindGraphByUID(ctx context.Context, uid string) (*api.Graph, error) {
	tx, err := gs.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	dg, err := tx.FindGraphByUID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &api.Graph{
		UID:   dg.UID(),
		Type:  dg.Type(),
		Nodes: dg.Nodes().Len(),
		Edges: dg.Edges().Len(),
		Label: dg.Label(),
		Attrs: dg.Attrs(),
	}, nil
}

// FindGraphs returns all graphs matching the filter.
// It also returns a count of total matching graphs which may differ from
// the number of returned graphs if the Limit field is set.
func (gs *GraphService) FindGraphs(ctx context.Context, filter api.GraphFilter) ([]*api.Graph, int, error) {
	tx, err := gs.db.BeginTx(ctx)
	if err != nil {
		return nil, 0, err
	}

	gx, n, err := tx.FindGraphs(ctx, filter)
	if err != nil {
		return nil, n, err
	}

	graphs := make([]*api.Graph, len(gx))

	for i, fg := range gx {
		graphs[i] = &api.Graph{
			UID:   fg.UID(),
			Type:  fg.Type(),
			Nodes: fg.Nodes().Len(),
			Edges: fg.Edges().Len(),
			Label: fg.Label(),
			Attrs: fg.Attrs(),
		}
	}

	return graphs, n, nil
}

// UpdateGraph updates an existing graph by ID.
func (gs *GraphService) UpdateGraph(ctx context.Context, uid string, update api.GraphUpdate) (*api.Graph, error) {
	tx, err := gs.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	ug, err := tx.UpdateGraph(ctx, uid, update)
	if err != nil {
		return nil, err
	}

	return &api.Graph{
		UID:   ug.UID(),
		Type:  ug.Type(),
		Nodes: ug.Nodes().Len(),
		Edges: ug.Edges().Len(),
		Label: ug.Label(),
		Attrs: ug.Attrs(),
	}, nil
}

// DeleteGraph permanently removes a graph by ID.
func (gs *GraphService) DeleteGraph(ctx context.Context, uid string) error {
	tx, err := gs.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	return tx.DeleteGraph(ctx, uid)
}
