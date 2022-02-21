package marshal

import (
	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

// Graph stores graph.
type Graph struct {
	Graph api.Graph  `json:"graph"`
	Nodes []api.Node `json:"nodes"`
	Edges []api.Edge `json:"edges"`
}
