package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

// NodeService lets you manage graphs.
type NodeService struct {
	db *DB
}

// NewNodeService creates an instance of NodeService and returns it.
func NewNodeService(db *DB) (*NodeService, error) {
	return &NodeService{
		db: db,
	}, nil
}

// CreateNode creates a new node.
func (ns *NodeService) CreateNode(ctx context.Context, graphUID string, n *api.Node) error {
	tx, err := ns.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback()

	n.CreatedAt = time.Now()
	n.UpdatedAt = n.CreatedAt

	attrs, err := json.Marshal(n.Attrs)
	if err != nil {
		return err
	}

	// Execute insertion query.
	result, err := tx.ExecContext(ctx, `
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
		n.UID,
		graphUID,
		n.Label,
		string(attrs),
		(*NullTime)(&n.CreatedAt),
		(*NullTime)(&n.UpdatedAt),
	)
	if err != nil {
		return err
	}

	n.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return tx.Commit()
}

// FindNodes returns all nodes matching the filter.
// It also returns a count of total matching nodes which may differ from
// the number of returned nodes if the Limit field is set.
func (ns *NodeService) FindNodes(ctx context.Context, graphUID string, filter api.NodeFilter) ([]*api.Node, int, error) {
	tx, err := ns.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	// nolint:errcheck
	defer tx.Rollback()

	return findNodes(ctx, tx, graphUID, filter)
}

func findNodes(ctx context.Context, tx *Tx, graphUID string, filter api.NodeFilter) (_ []*api.Node, n int, err error) {
	where, args := []string{"graph = ?"}, []interface{}{graphUID}
	switch {
	case filter.ID != nil:
		where, args = append(where, "id = ?"), append(args, *filter.ID)
	case filter.UID != nil:
		where, args = append(where, "uid = ?"), append(args, *filter.UID)
	case filter.To != nil:
		where = append(where, "uid IN (SELECT source FROM edges WHERE target = (SELECT uid FROM nodes WHERE id = ? AND graph = ?) AND graph = ?)")
		args = append(args, *filter.To, graphUID, graphUID)
	case filter.From != nil:
		where = append(where, "uid IN (SELECT target FROM edges WHERE source = (SELECT uid FROM nodes WHERE id = ? AND graph = ?) AND graph = ?)")
		args = append(args, *filter.From, graphUID, graphUID)
	}
	if v := filter.Label; v != nil {
		where, args = append(where, "label = ?"), append(args, *v)
	}

	query := `
		SELECT
		    id,
		    uid,
		    label,
		    attrs,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM nodes
		WHERE ` + strings.Join(where, " AND ") + `
		` + FormatLimitOffset(filter.Limit, filter.Offset)

	// Execute query to fetch node rows.
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, n, err
	}
	defer rows.Close()

	// Deserialize rows into Node objects.
	nodes := make([]*api.Node, 0)
	for rows.Next() {
		var node api.Node
		var attrString string
		if err := rows.Scan(
			&node.ID,
			&node.UID,
			&node.Label,
			&attrString,
			(*NullTime)(&node.CreatedAt),
			(*NullTime)(&node.UpdatedAt),
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
			for key, value := range attrs {
				if numStr, ok := value.(json.Number); ok {
					if num, err := numStr.Int64(); err == nil {
						attrs[key] = num
					} else if num, err := numStr.Float64(); err == nil {
						attrs[key] = num
					}
				}
			}
			node.Attrs = attrs
		}

		nodes = append(nodes, &node)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Calculate degree counts
	for _, node := range nodes {
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM edges
			WHERE source = ?
			`, node.UID).Scan(&node.DegOut); err != nil {
			return nil, 0, err
		}
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM edges
			WHERE target = ?
			`, node.UID).Scan(&node.DegIn); err != nil {
			return nil, 0, err
		}
	}

	return nodes, n, nil
}

func findNodeByID(ctx context.Context, tx *Tx, graphUID string, id int64) (*api.Node, error) {
	nodes, _, err := findNodes(ctx, tx, graphUID, api.NodeFilter{ID: &id})
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &api.Error{Code: api.ENOTFOUND, Message: "Node not found."}
	}
	return nodes[0], nil
}

// FindNodeByUID returns a single node with the given uid.
func (ns *NodeService) FindNodeByID(ctx context.Context, graphUID string, id int64) (*api.Node, error) {
	tx, err := ns.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	// nolint:errcheck
	defer tx.Rollback()

	node, err := findNodeByID(ctx, tx, graphUID, id)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func findNodeByUID(ctx context.Context, tx *Tx, graphUID, nodeUID string) (*api.Node, error) {
	nodes, _, err := findNodes(ctx, tx, graphUID, api.NodeFilter{UID: &nodeUID})
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &api.Error{Code: api.ENOTFOUND, Message: "Node not found."}
	}
	return nodes[0], nil
}

// FindNodeByUID returns a single node with the given uid.
func (ns *NodeService) FindNodeByUID(ctx context.Context, graphUID, nodeUID string) (*api.Node, error) {
	tx, err := ns.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	// nolint:errcheck
	defer tx.Rollback()

	node, err := findNodeByUID(ctx, tx, graphUID, nodeUID)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func updateNode(ctx context.Context, tx *Tx, graphUID string, id int64, update api.NodeUpdate) (*api.Node, error) {
	node, err := findNodeByID(ctx, tx, graphUID, id)
	if err != nil {
		return nil, err
	}

	if update.Label != nil {
		node.Label = update.Label
	}

	node.UpdatedAt = time.Now()

	// Prepare the SQL query for updating the node.
	sqlQuery := "UPDATE nodes SET label = ?, updated_at = ?"
	args := []interface{}{
		node.Label,
		(*NullTime)(&node.UpdatedAt),
	}

	if len(update.Attrs) > 0 {
		if node.Attrs == nil {
			node.Attrs = make(map[string]interface{})
		}
		for k, v := range update.Attrs {
			node.Attrs[k] = v
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

	// Add the node ID to the arguments for the WHERE clause.
	sqlQuery += " WHERE id = ? AND graph = ?"
	args = append(args, id, graphUID)

	// Execute the SQL query.
	if _, err := tx.ExecContext(ctx, sqlQuery, args...); err != nil {
		return nil, err
	}

	return node, nil
}

// UpdateNode updates an existing node by ID.
func (ns *NodeService) UpdateNode(ctx context.Context, graphUID string, id int64, update api.NodeUpdate) (*api.Node, error) {
	tx, err := ns.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	// nolint:errcheck
	defer tx.Rollback()

	node, err := updateNode(ctx, tx, graphUID, id, update)
	if err != nil {
		return node, err
	}

	if err := tx.Commit(); err != nil {
		return node, err
	}

	return node, nil
}

func deleteNodeByUID(ctx context.Context, tx *Tx, graphUID, nodeUID string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM nodes WHERE uid = ? AND graph = ?`, nodeUID, graphUID); err != nil {
		return err
	}
	return nil
}

// DeleteNodeByUID permanently removes a node by UID.
// It automatically removes removed node's incoming and outgoing edges.
func (ns *NodeService) DeleteNodeByUID(ctx context.Context, graphUID, nodeUID string) error {
	tx, err := ns.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback()

	if err := deleteNodeByUID(ctx, tx, graphUID, nodeUID); err != nil {
		return err
	}

	return tx.Commit()
}

func deleteNodeByID(ctx context.Context, tx *Tx, graphUID string, id int64) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM nodes WHERE id = ? AND graph = ?`, id, graphUID); err != nil {
		return err
	}
	return nil
}

// DeleteNodeByUID permanently removes a node by UID.
// It automatically removes removed node's incoming and outgoing edges.
func (ns *NodeService) DeleteNodeByID(ctx context.Context, graphUID string, id int64) error {
	tx, err := ns.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback()

	if err := deleteNodeByID(ctx, tx, graphUID, id); err != nil {
		return err
	}

	return tx.Commit()
}
