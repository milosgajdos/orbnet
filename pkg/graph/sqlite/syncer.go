package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/milosgajdos/orbnet/pkg/graph"
)

// Syncer syncs graph to sqlite.
type Syncer struct {
	db *DB
}

// NewSyncer creates a new sqlite syncer and returns it.
func NewSyncer(db *DB) (*Syncer, error) {
	return &Syncer{
		db: db,
	}, nil
}

// Sync sync graph g to sqlite DB.
func (s *Syncer) Sync(ctx context.Context, g graph.Graph) error {
	tx, err := s.db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback()

	if err := s.syncGraph(ctx, tx, g); err != nil {
		return err
	}

	nodes := g.Nodes()
	for nodes.Next() {
		n, ok := nodes.Node().(graph.Node)
		if !ok {
			continue
		}
		if err := s.syncNode(ctx, tx, g.UID(), n); err != nil {
			return err
		}
	}

	edges := g.Edges()
	for edges.Next() {
		e, ok := edges.Edge().(graph.Edge)
		if !ok {
			continue
		}
		if err := s.syncEdge(ctx, tx, g.UID(), e); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// syncGraph initializes the graph entry in the database.
func (s *Syncer) syncGraph(ctx context.Context, tx *sql.Tx, g graph.Graph) error {
	attrs, err := json.Marshal(g.Attrs())
	if err != nil {
		return err
	}

	createdAt := time.Now()
	updatedAt := createdAt

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO graphs (
			uid,
			label,
			attrs,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		g.UID(),
		g.Label(),
		attrs,
		(*NullTime)(&createdAt),
		(*NullTime)(&updatedAt),
	); err != nil {
		return err
	}

	return nil
}

// syncNode stores node in the sqlite DB.
func (s *Syncer) syncNode(ctx context.Context, tx *sql.Tx, graphUID string, n graph.Node) error {
	createdAt := time.Now()
	updatedAt := createdAt

	attrs, err := json.Marshal(n.Attrs())
	if err != nil {
		return err
	}

	// Execute insertion query.
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO nodes (
			uid,
			graph,
			label,
			attrs,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		n.UID(),
		graphUID,
		n.Label(),
		string(attrs),
		(*NullTime)(&createdAt),
		(*NullTime)(&updatedAt),
	); err != nil {
		return err
	}

	return nil
}

// syncEdge stores edge in the sqlite DB.
func (s *Syncer) syncEdge(ctx context.Context, tx *sql.Tx, graphUID string, e graph.Edge) error {
	attrs, err := json.Marshal(e.Attrs())
	if err != nil {
		return err
	}

	// Execute insertion query.
	_, err = tx.ExecContext(ctx, `
		WITH source_node AS (
		    SELECT uid AS source_uid FROM nodes WHERE id = ?
		),
		target_node AS (
		    SELECT uid AS target_uid FROM nodes WHERE id = ?
		)
		INSERT INTO edges (uid, graph, source, target, label, weight, attrs)
		VALUES (?, ?,
			(SELECT source_uid FROM source_node),
			(SELECT target_uid FROM target_node),
		?, ?, ?);
	`,
		e.From().ID(),
		e.To().ID(),
		e.UID(),
		graphUID,
		e.Label(),
		e.Weight(),
		string(attrs),
	)
	if err != nil {
		return err
	}

	return nil
}
