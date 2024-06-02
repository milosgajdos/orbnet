package sqlite

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"
)

// Loader loads graph from sqlite DB.
type Loader struct {
	db *DB
}

// NewLoader creates a new loader and returns it.
func NewLoader(db *DB) (*Loader, error) {
	return &Loader{
		db: db,
	}, nil
}

// Load loads the graph from sqlite DB and returns it.
func (l *Loader) Load(ctx context.Context, uid string) (graph.Graph, error) {
	tx, err := l.db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	// nolint:errcheck
	defer tx.Rollback()

	var (
		label     string
		attrsJSON string
		createdAt time.Time
		updatedAt time.Time
	)

	err = tx.QueryRowContext(ctx, `
		SELECT
			label,
			attrs,
			created_at,
			updated_at
		FROM graphs
		WHERE uid = ?
	`, uid).Scan(&label, &attrsJSON, (*NullTime)(&createdAt), (*NullTime)(&updatedAt))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve graph: %w", err)
	}

	attrs, err := AttrsFromString(attrsJSON)
	if err != nil {
		return nil, err
	}

	// Create the in-memory graph
	g, err := memory.NewGraph(
		memory.WithUID(uid),
		memory.WithLabel(label),
		memory.WithAttrs(attrs),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create in-memory graph: %w", err)
	}

	// Retrieve nodes
	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			uid,
			label,
			attrs,
			created_at,
			updated_at
		FROM nodes
		WHERE graph = ?
	`, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve nodes: %w", err)
	}
	defer rows.Close()

	nodeMap := make(map[string]*memory.Node) // Map to store nodes by their UID

	for rows.Next() {
		var (
			id            int64
			nodeUID       string
			nodeLabel     string
			nodeAttrsJSON string
			createdAt     time.Time
			updatedAt     time.Time
		)

		if err := rows.Scan(&id, &nodeUID, &nodeLabel, &nodeAttrsJSON, (*NullTime)(&createdAt), (*NullTime)(&updatedAt)); err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}

		nodeAttrs, err := AttrsFromString(nodeAttrsJSON)
		if err != nil {
			return nil, err
		}

		// Create node and add it to the graph
		node, err := memory.NewNode(id,
			memory.WithUID(nodeUID),
			memory.WithLabel(nodeLabel),
			memory.WithAttrs(nodeAttrs),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create node: %w", err)
		}
		g.AddNode(node)
		nodeMap[nodeUID] = node
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row error: %w", err)
	}

	// Retrieve edges
	edgeRows, err := tx.QueryContext(ctx, `
		SELECT
			uid,
			source,
			target,
			label,
			weight,
			attrs,
			created_at,
			updated_at
		FROM edges
		WHERE graph = ?
	`, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve edges: %w", err)
	}
	defer edgeRows.Close()

	for edgeRows.Next() {
		var (
			edgeUID       string
			sourceUID     string
			targetUID     string
			edgeLabel     string
			weight        float64
			edgeAttrsJSON string
			createdAt     time.Time
			updatedAt     time.Time
		)

		err := edgeRows.Scan(&edgeUID, &sourceUID, &targetUID, &edgeLabel,
			&weight, &edgeAttrsJSON, (*NullTime)(&createdAt), (*NullTime)(&updatedAt))
		if err != nil {
			return nil, fmt.Errorf("failed to scan edge: %w", err)
		}

		edgeAttrs, err := AttrsFromString(edgeAttrsJSON)
		if err != nil {
			return nil, err
		}

		// Create edge and add it to the graph
		sourceNode, sourceExists := nodeMap[sourceUID]
		targetNode, targetExists := nodeMap[targetUID]

		if !sourceExists || !targetExists {
			return nil, errors.New("source or target node does not exist")
		}

		edge, err := memory.NewEdge(sourceNode, targetNode,
			memory.WithUID(edgeUID),
			memory.WithLabel(edgeLabel),
			memory.WithWeight(weight),
			memory.WithAttrs(edgeAttrs),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create edge: %w", err)
		}
		g.SetWeightedEdge(edge)
	}

	if err := edgeRows.Err(); err != nil {
		return nil, fmt.Errorf("row error: %w", err)
	}

	return g, nil
}
