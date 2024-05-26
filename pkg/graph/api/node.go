package api

import (
	"context"
	"time"
)

// Node is a graph node.
type Node struct {
	// ID is node ID.
	ID int64 `json:"id"`
	// UID is node UUID.
	UID string `json:"uid,omitempty"`
	// DegOut is the count of outgoing edges.
	DegOut int `json:"deg_out"`
	// DegIn is the count of incoming edges.
	DegIn int `json:"deg_in"`
	// Label is node label
	Label *string `json:"label"`
	// Attrs are node attributes
	Attrs map[string]interface{} `json:"attributes,omitempty"`
	// Timestamps for graph creation & last update.
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NodeService represents a service for managing graph nodes.
type NodeService interface {
	// CreateNode creates a new node.
	CreateNode(ctx context.Context, uid string, n *Node) error
	// FindNodeByID returns a single node with the given id.
	FindNodeByID(ctx context.Context, uid string, id int64) (*Node, error)
	// FindNodeByUID returns a single node with the given uid.
	FindNodeByUID(ctx context.Context, guid, nuid string) (*Node, error)
	// FindNodes returns all nodes matching the filter.
	// It also returns a count of total matching nodes which may differ from
	// the number of returned nodes if the Limit field is set.
	FindNodes(ctx context.Context, uid string, filter NodeFilter) ([]*Node, int, error)
	// UpdateNode updates an existing node by ID.
	UpdateNode(ctx context.Context, uid string, id int64, update NodeUpdate) (*Node, error)
	// DeleteNodeByUID permanently removes a node by UID.
	// It automatically removes removed node's incoming and outgoing edges.
	DeleteNodeByUID(ctx context.Context, guid, nuid string) error
	// DeleteNodeByID permanently removes a node by ID.
	// It automatically removes removed node's incoming and outgoing edges.
	DeleteNodeByID(ctx context.Context, uid string, id int64) error
}

// NodeFilter represents a filter used by FindNodes().
type NodeFilter struct {
	// Filtering fields.
	ID    *int64  `json:"id"`
	UID   *string `json:"uid"`
	Label *string `json:"label"`
	// To gets all nodes that
	// can reach this node.
	To *int64 `json:"to"`
	// From gets all nodes
	// reachable from this node.
	From *int64 `json:"from"`
	// Restrict to subset of range.
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// NodeUpdate represents a set of fields to update on a node.
type NodeUpdate struct {
	Label *string                `json:"label"`
	Attrs map[string]interface{} `json:"attributes,omitempty"`
}
