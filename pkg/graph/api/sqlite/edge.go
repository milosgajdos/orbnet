package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

// EdgeService lets you manage edges.
type EdgeService struct {
	db *DB
}

// NewEdgeService creates an instance of EdgeService and returns it.
func NewEdgeService(db *DB) (*EdgeService, error) {
	return &EdgeService{
		db: db,
	}, nil
}

// CreateEdge creates a new edge.
func (es *EdgeService) CreateEdge(ctx context.Context, graphUID string, e *api.Edge) error {
	tx, err := es.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback()

	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt

	attrs, err := json.Marshal(e.Attrs)
	if err != nil {
		return err
	}

	// Execute insertion query.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO edges (
			uid,
			graph,
			source,
			target,
			label,
			weight,
			attrs,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		e.UID,
		graphUID,
		e.Source,
		e.Target,
		e.Label,
		e.Weight,
		string(attrs),
		(*NullTime)(&e.CreatedAt),
		(*NullTime)(&e.UpdatedAt),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// FindEdgeByUID returns a single edge with the given uid.
func (es *EdgeService) FindEdgeByUID(ctx context.Context, graphUID, edgeUID string) (*api.Edge, error) {
	tx, err := es.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	// nolint:errcheck
	defer tx.Rollback()

	edge, err := findEdgeByUID(ctx, tx, graphUID, edgeUID)
	if err != nil {
		return nil, err
	}

	return edge, nil
}

// FindEdges returns all edges matching the filter.
func (es *EdgeService) FindEdges(ctx context.Context, graphUID string, filter api.EdgeFilter) ([]*api.Edge, int, error) {
	tx, err := es.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	// nolint:errcheck
	defer tx.Rollback()

	return findEdges(ctx, tx, graphUID, filter)
}

// UpdateEdgeBetween updates an edge between two nodes.
func (es *EdgeService) UpdateEdgeBetween(ctx context.Context, graphUID, source, target string, update api.EdgeUpdate) (*api.Edge, error) {
	tx, err := es.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	// nolint:errcheck
	defer tx.Rollback()

	edge, err := updateEdgeBetween(ctx, tx, graphUID, source, target, update)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return edge, nil
}

// DeleteEdge permanently removes an edge by UID.
func (es *EdgeService) DeleteEdge(ctx context.Context, graphUID, edgeUID string) error {
	tx, err := es.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback()

	if err := deleteEdge(ctx, tx, graphUID, edgeUID); err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteEdgeBetween permanently deletes all edges between two nodes.
func (es *EdgeService) DeleteEdgeBetween(ctx context.Context, graphUID, source, target string) error {
	tx, err := es.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback()

	if err := deleteEdgeBetween(ctx, tx, graphUID, source, target); err != nil {
		return err
	}

	return tx.Commit()
}

// Helper functions

func findEdgeByUID(ctx context.Context, tx *Tx, graphUID, edgeUID string) (*api.Edge, error) {
	var edge api.Edge
	var attrString string
	if err := tx.QueryRowContext(ctx, `
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
		WHERE uid = ? AND graph = ?
	`, edgeUID, graphUID).Scan(
		&edge.UID,
		&edge.Source,
		&edge.Target,
		&edge.Label,
		&edge.Weight,
		&attrString,
		(*NullTime)(&edge.CreatedAt),
		(*NullTime)(&edge.UpdatedAt),
	); err != nil {
		return nil, &api.Error{Code: api.ENOTFOUND, Message: "Edge not found."}
	}

	if len(attrString) > 0 {
		attrs := map[string]interface{}{}
		decoder := json.NewDecoder(strings.NewReader(attrString))
		decoder.UseNumber()
		if err := decoder.Decode(&attrs); err != nil {
			return nil, err
		}
		// Convert json.Number to int or float64 as needed
		for key, value := range attrs {
			if numStr, ok := value.(json.Number); ok {
				if num, err := numStr.Int64(); err == nil {
					attrs[key] = num
				} else if num, err := numStr.Float64(); err == nil {
					attrs[key] = num
				}
			}
		}
		edge.Attrs = attrs
	}

	return &edge, nil
}

func findEdges(ctx context.Context, tx *Tx, graphUID string, filter api.EdgeFilter) (_ []*api.Edge, n int, err error) {
	where, args := []string{"graph = ?"}, []interface{}{graphUID}
	if v := filter.Source; v != nil {
		where, args = append(where, "source = ?"), append(args, *v)
	}
	if v := filter.Target; v != nil {
		where, args = append(where, "target = ?"), append(args, *v)
	}
	if v := filter.Label; v != nil {
		where, args = append(where, "label = ?"), append(args, *v)
	}

	query := `
		SELECT
		    uid,
		    source,
		    target,
		    label,
		    weight,
		    attrs,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM edges
		WHERE ` + strings.Join(where, " AND ") + `
		` + FormatLimitOffset(filter.Limit, filter.Offset)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, n, err
	}
	defer rows.Close()

	edges := make([]*api.Edge, 0)
	for rows.Next() {
		var edge api.Edge
		var attrString string
		if err := rows.Scan(
			&edge.UID,
			&edge.Source,
			&edge.Target,
			&edge.Label,
			&edge.Weight,
			&attrString,
			(*NullTime)(&edge.CreatedAt),
			(*NullTime)(&edge.UpdatedAt),
			&n,
		); err != nil {
			return nil, 0, err
		}

		if len(attrString) > 0 {
			attrs := map[string]interface{}{}
			decoder := json.NewDecoder(strings.NewReader(attrString))
			decoder.UseNumber()
			if err := decoder.Decode(&attrs); err != nil {
				return nil, 0, err
			}
			// Convert json.Number to int or float64 as needed
			for key, value := range attrs {
				if numStr, ok := value.(json.Number); ok {
					if num, err := numStr.Int64(); err == nil {
						attrs[key] = num
					} else if num, err := numStr.Float64(); err == nil {
						attrs[key] = num
					}
				}
			}
			edge.Attrs = attrs
		}

		edges = append(edges, &edge)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return edges, n, nil
}

func updateEdgeBetween(ctx context.Context, tx *Tx, graphUID, source, target string, update api.EdgeUpdate) (*api.Edge, error) {
	edge, err := findEdgeBetween(ctx, tx, graphUID, source, target)
	if err != nil {
		return nil, err
	}

	if update.Label != nil {
		edge.Label = *update.Label
	}
	if update.Weight != nil {
		edge.Weight = *update.Weight
	}

	edge.UpdatedAt = time.Now()

	// Prepare the SQL query for updating the edge.
	sqlQuery := "UPDATE edges SET label = ?, weight = ?, updated_at = ?"
	args := []interface{}{
		edge.Label,
		edge.Weight,
		(*NullTime)(&edge.UpdatedAt),
	}

	if len(update.Attrs) > 0 {
		if edge.Attrs == nil {
			edge.Attrs = make(map[string]interface{})
		}
		for k, v := range update.Attrs {
			edge.Attrs[k] = v
		}
		// Convert the new attributes to JSON.
		newAttrsJSON, err := json.Marshal(update.Attrs)
		if err != nil {
			return nil, err
		}
		// Use JSON functions to merge the attributes.
		sqlQuery += ", attrs = json_set(attrs"
		for key := range update.Attrs {
			sqlQuery += fmt.Sprintf(", '$.%s', json_extract(?, '$.%s')", key, key)
			args = append(args, string(newAttrsJSON))
		}
		sqlQuery += ")"
	}

	// Add the source and target to the arguments for the WHERE clause.
	sqlQuery += " WHERE source = ? AND target = ? AND graph = ?"
	args = append(args, source, target, graphUID)

	// Execute the SQL query.
	if _, err := tx.ExecContext(ctx, sqlQuery, args...); err != nil {
		return nil, err
	}

	return edge, nil
}

func deleteEdge(ctx context.Context, tx *Tx, graphUID, edgeUID string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM edges WHERE uid = ? AND graph = ?`, edgeUID, graphUID); err != nil {
		return err
	}
	return nil
}

func deleteEdgeBetween(ctx context.Context, tx *Tx, graphUID, source, target string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM edges WHERE source = ? AND target = ? AND graph = ?`, source, target, graphUID); err != nil {
		return err
	}
	return nil
}

func findEdgeBetween(ctx context.Context, tx *Tx, graphUID, source, target string) (*api.Edge, error) {
	var edge api.Edge
	var attrString string

	if err := tx.QueryRowContext(ctx, `
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
		WHERE source = ? AND target = ? AND graph = ?
	`, source, target, graphUID).Scan(
		&edge.UID,
		&edge.Source,
		&edge.Target,
		&edge.Label,
		&edge.Weight,
		&attrString,
		(*NullTime)(&edge.CreatedAt),
		(*NullTime)(&edge.UpdatedAt),
	); err != nil {
		return nil, &api.Error{Code: api.ENOTFOUND, Message: "Edge not found."}
	}

	if len(attrString) > 0 {
		attrs := map[string]interface{}{}
		decoder := json.NewDecoder(strings.NewReader(attrString))
		decoder.UseNumber()
		if err := decoder.Decode(&attrs); err != nil {
			return nil, err
		}
		// Convert json.Number to int or float64 as needed
		for key, value := range attrs {
			if numStr, ok := value.(json.Number); ok {
				if num, err := numStr.Int64(); err == nil {
					attrs[key] = num
				} else if num, err := numStr.Float64(); err == nil {
					attrs[key] = num
				}
			}
		}
		edge.Attrs = attrs
	}

	return &edge, nil
}
