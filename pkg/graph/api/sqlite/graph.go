package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	tx, err := gs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback()

	g.CreatedAt = time.Now()
	g.UpdatedAt = g.CreatedAt

	attrs, err := json.Marshal(g.Attrs)
	if err != nil {
		return err
	}

	// Execute insertion query.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO graphs (
			uid,
			label,
			attrs,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		g.UID,
		g.Label,
		string(attrs),
		(*NullTime)(&g.CreatedAt),
		(*NullTime)(&g.UpdatedAt),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// FindGraphByUID returns a single graph with the given uid.
func (gs *GraphService) FindGraphByUID(ctx context.Context, uid string) (*api.Graph, error) {
	tx, err := gs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	// nolint:errcheck
	defer tx.Rollback()

	graph, err := findGraphByUID(ctx, tx, uid)
	if err != nil {
		return nil, err
	}

	return graph, nil
}

// FindGraphs returns all graphs matching the filter.
// It also returns a count of total matching graphs which may differ from
// the number of returned graphs if the Limit field is set.
func (gs *GraphService) FindGraphs(ctx context.Context, filter api.GraphFilter) ([]*api.Graph, int, error) {
	tx, err := gs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	// nolint:errcheck
	defer tx.Rollback()

	return findGraphs(ctx, tx, filter)
}

func findGraphByUID(ctx context.Context, tx *Tx, uid string) (*api.Graph, error) {
	gx, _, err := findGraphs(ctx, tx, api.GraphFilter{UID: &uid})
	if err != nil {
		return nil, err
	} else if len(gx) == 0 {
		return nil, &api.Error{Code: api.ENOTFOUND, Message: "Graph not found."}
	}
	return gx[0], nil
}

func findGraphs(ctx context.Context, tx *Tx, filter api.GraphFilter) (_ []*api.Graph, n int, err error) {
	// Build WHERE clause.
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.UID; v != nil {
		where, args = append(where, "uid = ?"), append(args, *v)
	}
	if v := filter.Label; v != nil {
		where, args = append(where, "label = ?"), append(args, *v)
	}
	// Execute query to fetch user rows.
	rows, err := tx.QueryContext(ctx, `
		SELECT
		    uid,
		    label,
		    attrs,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM graphs
		WHERE `+strings.Join(where, " AND ")+`
		`+FormatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err != nil {
		return nil, n, err
	}
	defer rows.Close()

	// Deserialize rows into Graph objects.
	graphs := make([]*api.Graph, 0)
	for rows.Next() {
		var graph api.Graph
		var attrString string
		if err := rows.Scan(
			&graph.UID,
			&graph.Label,
			&attrString,
			(*NullTime)(&graph.CreatedAt),
			(*NullTime)(&graph.UpdatedAt),
			&n,
		); err != nil {
			return nil, 0, err
		}
		if len(attrString) > 0 {
			attrs := map[string]interface{}{}
			decoder := json.NewDecoder(strings.NewReader(attrString))
			decoder.UseNumber()
			err := decoder.Decode(&attrs)
			if err != nil {
				return nil, 0, err
			}
			// Convert json.Number to int or float64 as needed
			// NOTE: thanks golang
			for key, value := range attrs {
				if numStr, ok := value.(json.Number); ok {
					if num, err := numStr.Int64(); err == nil {
						attrs[key] = num
					} else if num, err := numStr.Float64(); err == nil {
						attrs[key] = num
					}
				}
			}
			graph.Attrs = attrs
		}

		graphs = append(graphs, &graph)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return graphs, n, nil
}

// UpdateGraph updates an existing graph by ID.
func (gs *GraphService) UpdateGraph(ctx context.Context, uid string, update api.GraphUpdate) (*api.Graph, error) {
	tx, err := gs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	// nolint:errcheck
	defer tx.Rollback()

	// Update graph & attach associated OAuth objects.
	graph, err := updateGraph(ctx, tx, uid, update)
	if err != nil {
		return graph, err
	}

	if err := tx.Commit(); err != nil {
		return graph, err
	}

	return graph, nil
}

func updateGraph(ctx context.Context, tx *Tx, uid string, update api.GraphUpdate) (*api.Graph, error) {
	graph, err := findGraphByUID(ctx, tx, uid)
	if err != nil {
		return nil, err
	}

	if update.Label != nil {
		graph.Label = update.Label
	}

	graph.UpdatedAt = time.Now()

	// Prepare the SQL query for updating the graph.
	sqlQuery := "UPDATE graphs SET label = ?, updated_at = ?"
	args := []interface{}{
		graph.Label,
		(*NullTime)(&graph.UpdatedAt),
	}

	if len(update.Attrs) > 0 {
		if graph.Attrs == nil {
			graph.Attrs = make(map[string]interface{})
		}
		for k, v := range update.Attrs {
			graph.Attrs[k] = v
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

	// Add the UID to the arguments for the WHERE clause.
	sqlQuery += " WHERE uid = ?"
	args = append(args, uid)

	// Execute the SQL query.
	if _, err := tx.ExecContext(ctx, sqlQuery, args...); err != nil {
		return nil, err
	}

	return graph, nil
}

// DeleteGraph permanently removes a graph by ID.
func (gs *GraphService) DeleteGraph(ctx context.Context, uid string) error {
	tx, err := gs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback()

	if err := deleteGraph(ctx, tx, uid); err != nil {
		return err
	}
	return tx.Commit()

}

func deleteGraph(ctx context.Context, tx *Tx, uid string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM graphs WHERE uid = ?`, uid); err != nil {
		return err
	}
	return nil
}
