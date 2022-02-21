package api

import "context"

// Edge is an edge between two graph nodes.
type Edge struct {
	// UID is edge UUID.
	UID string `json:"uid,omitempty"`
	// Source is an ID of the edge source node.
	Source int64 `json:"source"`
	// Target is an ID of the edge target node.
	Target int64 `json:"target"`
	// Weight is the edge weight.
	Weight float64 `json:"weight"`
	// Label is edge label
	Label string `json:"label"`
	// Attrs are edge attributes
	Attrs map[string]interface{} `json:"attributes,omitempty"`
}

// EdgeService represents a service for managing Edges.
type EdgeService interface {
	// CreateEdge creates a new edge.
	CreateEdge(ctx context.Context, uid string, e *Edge) error
	// FindEdgeByUID returns a single edge with the given id.
	FindEdgeByUID(ctx context.Context, guid, euid string) (*Edge, error)
	// FindEdges returns all edges matching the filter.
	// It also returns a count of total matching edges which may differ from
	// the number of returned edges if the Limit field is set.
	FindEdges(ctx context.Context, uid string, filter EdgeFilter) ([]*Edge, int, error)
	// UpdateEdgeBetween updates an edge between two nodes.
	UpdateEdgeBetween(ctx context.Context, uid string, source, target int64, update EdgeUpdate) (*Edge, error)
	// DeleteEdge permanently removes an edge by UID.
	DeleteEdge(ctx context.Context, guid, euid string) error
	// DeleteEdgeBetween permanently deletes all edges between two nodes.
	DeleteEdgeBetween(ctx context.Context, uid string, source, target int64) error
}

// EdgeFilter represents a filter used by FindEdges().
type EdgeFilter struct {
	// Filtering fields.
	Source *int64  `json:"source"`
	Target *int64  `json:"target"`
	Label  *string `json:"label"`
	// Restrict to subset of range.
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// EdgeUpdate represents a set of fields to update on an edge.
type EdgeUpdate struct {
	Weight *float64               `json:"weight"`
	Label  *string                `json:"label"`
	Attrs  map[string]interface{} `json:"attributes,omitempty"`
}
