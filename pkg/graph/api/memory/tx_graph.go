package memory

import (
	"context"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"
)

// CreateGraph creates a new graph and returns it.
func (t *Tx) CreateGraph(ctx context.Context, g *api.Graph) error {
	t.db.Lock()
	defer t.db.Unlock()

	var opts []memory.Option

	if g.UID != "" {
		opts = append(opts, memory.WithUID(g.UID))
	}

	if g.Label != "" {
		opts = append(opts, memory.WithLabel(g.Label))
	}

	if g.Attrs != nil {
		opts = append(opts, memory.WithAttrs(g.Attrs))
	}

	// NOTE(milosgajdos): we are ignoring g.Type
	// as we only support weighted_directed graphs

	mg, err := memory.NewGraph(opts...)
	if err != nil {
		return err
	}

	if _, ok := t.db.db[mg.UID()]; ok {
		return api.Errorf(api.ECONFLICT, "graph %s already exists", mg.UID())
	}

	t.db.db[mg.UID()] = mg

	g.UID = mg.UID()
	g.Type = mg.Type()
	g.Nodes = mg.Nodes().Len()
	g.Edges = mg.Edges().Len()
	g.Label = mg.Label()
	g.Attrs = mg.Attrs()

	return nil
}

// FindGraphByUID returns graph with the given UID.
// It returns error if the graph with the given uid could not be found.
func (t *Tx) FindGraphByUID(ctx context.Context, uid string) (*memory.Graph, error) {
	t.db.RLock()
	defer t.db.RUnlock()

	g, ok := t.db.db[uid]
	if !ok {
		return nil, api.Errorf(api.ENOTFOUND, "graph %s not found", uid)
	}

	return memory.GraphDeepCopy(g), nil
}

// FindGraphs returns all graphs matching the filter.
func (t *Tx) FindGraphs(ctx context.Context, filter api.GraphFilter) ([]*memory.Graph, int, error) {
	t.db.RLock()
	defer t.db.RUnlock()

	// nolint:prealloc
	var graphs []*memory.Graph

	if uid := filter.UID; uid != nil {
		g, ok := t.db.db[*uid]
		if !ok {
			return graphs, 0, nil
		}

		if t := filter.Type; t != nil {
			if g.Type() != *t {
				return graphs, 0, nil
			}
		}

		if l := filter.Label; l != nil {
			if g.Label() != *l {
				return graphs, 0, nil
			}
		}

		cg := memory.GraphDeepCopy(g)
		return []*memory.Graph{cg}, 1, nil
	}

	for _, g := range t.db.db {
		if t := filter.Type; t != nil {
			if g.Type() != *t {
				continue
			}
		}

		if l := filter.Label; l != nil {
			if g.Label() != *l {
				continue
			}
		}

		cg := memory.GraphDeepCopy(g)
		graphs = append(graphs, cg)
	}

	return applyOffsetLimit(graphs, filter.Offset, filter.Limit).([]*memory.Graph), len(graphs), nil
}

// UpdateGraph updates graph in db.
func (t *Tx) UpdateGraph(ctx context.Context, uid string, update api.GraphUpdate) (*memory.Graph, error) {
	t.db.Lock()
	defer t.db.Unlock()

	g, ok := t.db.db[uid]
	if !ok {
		return nil, api.Errorf(api.ENOTFOUND, "graph %s not found", uid)
	}

	if l := update.Label; l != nil {
		g.SetLabel(*l)
	}

	if a := update.Attrs; a != nil {
		for k, v := range a {
			g.Attrs()[k] = v
		}
	}

	return memory.GraphDeepCopy(g), nil
}

// DeleteGraph deletes graph from db.
func (t *Tx) DeleteGraph(ctx context.Context, uid string) error {
	t.db.Lock()
	defer t.db.Unlock()
	delete(t.db.db, uid)
	return nil
}
