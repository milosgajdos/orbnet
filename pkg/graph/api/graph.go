package api

import (
	"context"
	"time"
)

// Graph contains nodes connected by edges.
type Graph struct {
	// UID is graph UUID.
	UID string `json:"uid,omitempty"`
	// Nodes is the node count.
	Nodes int `json:"nodes"`
	// Edge is the edge count.
	Edges int `json:"edges"`
	// Label is graph label.
	Label *string `json:"label,omitempty"`
	// Attrs are graph attributes.
	Attrs map[string]interface{} `json:"attributes,omitempty"`
	// Timestamps for graph creation & last update.
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// GraphService represents a service for managing Graphs.
type GraphService interface {
	// CreateGraph creates a new graph.
	CreateGraph(ctx context.Context, g *Graph) error
	// FindGraphByUID returns a single graph with the given uid.
	FindGraphByUID(ctx context.Context, uid string) (*Graph, error)
	// FindGraphs returns all graphs matching the filter.
	// It also returns a count of total matching graphs which may differ from
	// the number of returned graphs if the Limit field is set.
	FindGraphs(ctx context.Context, filter GraphFilter) ([]*Graph, int, error)
	// UpdateGraph updates an existing graph by UID.
	UpdateGraph(ctx context.Context, uid string, update GraphUpdate) (*Graph, error)
	// DeleteGraph permanently removes a graph by UID.
	DeleteGraph(ctx context.Context, uid string) error
}

// GraphFilter represents a filter used by FindGraphs().
type GraphFilter struct {
	// Filtering fields.
	UID   *string `json:"uid"`
	Label *string `json:"label"`
	// Restrict to subset of range.
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// GraphUpdate represents a set of fields to update on a graph.
type GraphUpdate struct {
	Label *string                `json:"label"`
	Attrs map[string]interface{} `json:"attributes,omitempty"`
}
